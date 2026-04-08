package httpserver

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"wack-backend/internal/httpserver/dto"
	"wack-backend/internal/model"
	"wack-backend/internal/query"
	"wack-backend/internal/service"
)

func (h *apiHandler) listStudentOptions(c *gin.Context) {
	onlyUnbound := strings.EqualFold(c.DefaultQuery("binding", ""), "unbound")
	items, err := h.students.ListStudentOptions(c.Query("keyword"), onlyUnbound)
	if err != nil {
		fail(c, 500, "list student options failed")
		return
	}
	ok(c, items)
}

func (h *apiHandler) listStudents(c *gin.Context) {
	requestPage, pageSize := parsePage(c)
	page := requestPage
	classID, _ := strconv.ParseUint(c.DefaultQuery("class_id", "0"), 10, 64)
	focusStudentID, _ := strconv.ParseUint(c.DefaultQuery("focus_student_ref_id", "0"), 10, 64)
	listInput := service.ListStudentsInput{
		Page:                    page,
		PageSize:                pageSize,
		ClassID:                 classID,
		Keyword:                 c.Query("keyword"),
		StudentID:               c.Query("student_id"),
		RealName:                c.Query("real_name"),
		ClassName:               c.Query("class_name"),
		Term:                    c.Query("term"),
		AttendanceSummaryStatus: c.Query("attendance_summary_status"),
	}
	var focusResult *query.FocusPageResult
	if focusStudentID > 0 {
		located, err := h.students.LocateStudentPage(listInput, focusStudentID)
		if err != nil {
			fail(c, 500, "locate student page failed")
			return
		}
		focusResult = &located
		if located.Found {
			page = located.Page
			listInput.Page = page
		}
	}
	items, total, err := h.students.ListStudents(service.ListStudentsInput{
		Page:                    page,
		PageSize:                pageSize,
		ClassID:                 classID,
		Keyword:                 c.Query("keyword"),
		StudentID:               c.Query("student_id"),
		RealName:                c.Query("real_name"),
		ClassName:               c.Query("class_name"),
		Term:                    c.Query("term"),
		AttendanceSummaryStatus: c.Query("attendance_summary_status"),
	})
	if err != nil {
		fail(c, 500, "list students failed")
		return
	}
	result := pageResult[query.StudentItem]{Items: items, Page: page, PageSize: pageSize, Total: total}
	if focusResult != nil {
		result.FocusFound = &focusResult.Found
		if focusResult.Found {
			result.FocusPage = &focusResult.Page
			result.FocusRowKey = &focusResult.RowKey
		}
	}
	ok(c, result)
}

func (h *apiHandler) listStudentAttendanceRecords(c *gin.Context) {
	studentID, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	page, pageSize := parsePage(c)
	student, items, total, err := h.students.GetStudentAttendancePage(studentID, service.ListStudentAttendanceInput{
		Page:         page,
		PageSize:     pageSize,
		Term:         c.Query("term"),
		LessonDate:   c.Query("lesson_date"),
		Section:      c.Query("section"),
		CourseName:   c.Query("course_name"),
		TeacherName:  c.Query("teacher_name"),
		Status:       c.Query("status"),
		OperatorName: c.Query("operator_name"),
		OperatedDate: c.Query("operated_date"),
	})
	if err != nil {
		switch {
		case service.IsServiceError(err, service.ErrStudentNotFound):
			fail(c, 404, "student not found")
		default:
			fail(c, 500, "load student attendance records failed")
		}
		return
	}
	ok(c, gin.H{
		"student":            student,
		"attendance_records": items,
		"page":               page,
		"page_size":          pageSize,
		"total":              total,
	})
}

func (h *apiHandler) bulkUpdateStudentAttendanceRecordStatuses(c *gin.Context) {
	user, _ := currentUser(c)
	studentID, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	var req dto.BulkUpdateAttendanceRecordStatusesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "invalid request")
		return
	}
	if req.Status == nil || len(req.AttendanceRecordIDs) == 0 {
		fail(c, 400, "invalid request")
		return
	}
	if *req.Status < model.AttendancePresent || *req.Status > model.AttendanceOnLeave {
		fail(c, 400, "invalid status")
		return
	}
	if _, err := h.students.GetStudent(studentID); err != nil {
		switch {
		case service.IsServiceError(err, service.ErrStudentNotFound):
			fail(c, 404, "student not found")
		default:
			fail(c, 500, "load student attendance records failed")
		}
		return
	}
	result, err := h.attendance.BulkUpdateAttendanceRecordStatuses(req.AttendanceRecordIDs, *req.Status, user.ID)
	if err != nil {
		fail(c, 500, "update attendance statuses failed")
		return
	}
	ok(c, result)
}

func (h *apiHandler) createStudent(c *gin.Context) {
	var req dto.StudentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "invalid request")
		return
	}
	item, err := h.students.CreateStudent(model.Student{
		StudentNo:   req.StudentID,
		StudentName: req.RealName,
		ClassID:     req.ClassID,
	})
	if err != nil {
		switch {
		case service.IsServiceError(err, service.ErrClassNotFound):
			fail(c, 404, "class not found")
		case service.IsServiceError(err, service.ErrStudentNoAlreadyExists):
			fail(c, 409, "student no already exists")
		case service.IsServiceError(err, service.ErrInvalidInput):
			fail(c, 400, "invalid request")
		default:
			fail(c, 400, "create student failed")
		}
		return
	}
	ok(c, item)
}

func (h *apiHandler) updateStudent(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	var req dto.StudentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "invalid request")
		return
	}
	item, err := h.students.UpdateStudent(id, model.Student{
		StudentNo:   req.StudentID,
		StudentName: req.RealName,
		ClassID:     req.ClassID,
	})
	if err != nil {
		switch {
		case service.IsServiceError(err, service.ErrStudentNotFound):
			fail(c, 404, "student not found")
		case service.IsServiceError(err, service.ErrClassNotFound):
			fail(c, 404, "class not found")
		case service.IsServiceError(err, service.ErrInvalidInput):
			fail(c, 400, "invalid request")
		default:
			fail(c, 400, "update student failed")
		}
		return
	}
	ok(c, item)
}

func (h *apiHandler) deleteStudent(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	if err := h.students.DeleteStudent(id); err != nil {
		switch {
		case service.IsServiceError(err, service.ErrStudentNotFound):
			fail(c, 404, "student not found")
		default:
			fail(c, 400, "delete student failed")
		}
		return
	}
	ok(c, gin.H{})
}
