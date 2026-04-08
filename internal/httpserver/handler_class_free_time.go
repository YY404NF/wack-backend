package httpserver

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"wack-backend/internal/httpserver/dto"
	"wack-backend/internal/model"
	"wack-backend/internal/query"
	"wack-backend/internal/service"
)

func (h *apiHandler) listClasses(c *gin.Context) {
	requestPage, pageSize := parsePage(c)
	page := requestPage
	focusClassID, _ := strconv.ParseUint(c.DefaultQuery("focus_class_id", "0"), 10, 64)
	var focusResult *query.FocusPageResult
	if focusClassID > 0 {
		located, err := h.classes.LocateClassPage(
			c.Query("grade"),
			c.Query("major_name"),
			c.Query("class_name"),
			c.Query("term"),
			c.Query("attendance_summary_status"),
			focusClassID,
			pageSize,
		)
		if err != nil {
			fail(c, 500, "locate class page failed")
			return
		}
		focusResult = &located
		if located.Found {
			page = located.Page
		}
	}
	classes, total, err := h.classes.ListClasses(
		c.Query("grade"),
		c.Query("major_name"),
		c.Query("class_name"),
		c.Query("term"),
		c.Query("attendance_summary_status"),
		page,
		pageSize,
	)
	if err != nil {
		fail(c, 500, "list classes failed")
		return
	}
	result := pageResult[model.Class]{Items: classes, Page: page, PageSize: pageSize, Total: total}
	if focusResult != nil {
		result.FocusFound = &focusResult.Found
		if focusResult.Found {
			result.FocusPage = &focusResult.Page
			result.FocusRowKey = &focusResult.RowKey
		}
	}
	ok(c, result)
}

func (h *apiHandler) listClassOptions(c *gin.Context) {
	items, err := h.classes.ListClassOptions(c.Query("keyword"))
	if err != nil {
		fail(c, 500, "list class options failed")
		return
	}
	ok(c, items)
}

func (h *apiHandler) createClass(c *gin.Context) {
	var classItem model.Class
	if err := c.ShouldBindJSON(&classItem); err != nil {
		fail(c, 400, "invalid request")
		return
	}
	classItem, err := h.classes.CreateClass(classItem)
	if err != nil {
		switch {
		case service.IsServiceError(err, service.ErrInvalidInput):
			fail(c, 400, "invalid request")
		default:
			fail(c, 400, "create class failed")
		}
		return
	}
	ok(c, classItem)
}

func (h *apiHandler) getClass(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	classItem, err := h.classes.GetClass(id)
	if err != nil {
		fail(c, 404, "class not found")
		return
	}
	ok(c, classItem)
}

func (h *apiHandler) updateClass(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	var req model.Class
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "invalid request")
		return
	}
	classItem, err := h.classes.UpdateClass(id, req)
	if err != nil {
		switch {
		case service.IsServiceError(err, service.ErrClassNotFound):
			fail(c, 404, "class not found")
		case service.IsServiceError(err, service.ErrInvalidInput):
			fail(c, 400, "invalid request")
		default:
			fail(c, 400, "update class failed")
		}
		return
	}
	ok(c, classItem)
}

func (h *apiHandler) deleteClass(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	if err := h.classes.DeleteClass(id); err != nil {
		fail(c, 400, "delete class failed")
		return
	}
	ok(c, gin.H{})
}

func (h *apiHandler) getClassAttendanceRecords(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	page, pageSize := parsePage(c)
	classItem, records, total, err := h.classes.GetClassAttendancePage(id, service.ListClassAttendanceInput{
		Page:         page,
		PageSize:     pageSize,
		Term:         c.Query("term"),
		LessonDate:   c.Query("lesson_date"),
		Section:      c.Query("section"),
		CourseName:   c.Query("course_name"),
		TeacherName:  c.Query("teacher_name"),
		StudentID:    c.Query("student_id"),
		RealName:     c.Query("real_name"),
		Status:       c.Query("status"),
		OperatorName: c.Query("operator_name"),
		OperatedDate: c.Query("operated_date"),
	})
	if err != nil {
		switch {
		case service.IsServiceError(err, service.ErrClassNotFound):
			fail(c, 404, "class not found")
		default:
			fail(c, 500, "load class attendance records failed")
		}
		return
	}
	ok(c, gin.H{
		"class":              classItem,
		"attendance_records": records,
		"page":               page,
		"page_size":          pageSize,
		"total":              total,
	})
}

func (h *apiHandler) bulkUpdateClassAttendanceRecordStatuses(c *gin.Context) {
	user, _ := currentUser(c)
	id, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	if _, err := h.classes.GetClass(id); err != nil {
		switch {
		case service.IsServiceError(err, service.ErrClassNotFound):
			fail(c, 404, "class not found")
		default:
			fail(c, 500, "load class attendance records failed")
		}
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
	result, err := h.attendance.BulkUpdateAttendanceRecordStatuses(req.AttendanceRecordIDs, *req.Status, user.ID)
	if err != nil {
		fail(c, 500, "update attendance statuses failed")
		return
	}
	ok(c, result)
}

func (h *apiHandler) getClassStudents(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	users, err := h.classes.GetClassStudents(id)
	if err != nil {
		switch {
		case service.IsServiceError(err, service.ErrClassNotFound):
			fail(c, 404, "class not found")
		default:
			fail(c, 500, "get class students failed")
		}
		return
	}
	ok(c, users)
}

func (h *apiHandler) studentManagedClass(c *gin.Context) {
	user, exists := currentUser(c)
	if !exists {
		fail(c, 401, "unauthorized")
		return
	}
	if user.Role != model.RoleCommissioner {
		ok(c, gin.H{
			"managed_class":  nil,
			"class_students": []query.ClassStudentItem{},
		})
		return
	}
	if user.ManagedClassID == nil {
		ok(c, gin.H{
			"managed_class":  nil,
			"class_students": []query.ClassStudentItem{},
		})
		return
	}

	classItem, err := h.classes.GetClass(*user.ManagedClassID)
	if err != nil {
		switch {
		case service.IsServiceError(err, service.ErrClassNotFound):
			fail(c, 404, "managed class not found")
		default:
			fail(c, 500, "load managed class failed")
		}
		return
	}
	students, err := h.classes.GetClassStudents(*user.ManagedClassID)
	if err != nil {
		switch {
		case service.IsServiceError(err, service.ErrClassNotFound):
			fail(c, 404, "managed class not found")
		default:
			fail(c, 500, "load managed class students failed")
		}
		return
	}

	ok(c, gin.H{
		"managed_class":  classItem,
		"class_students": students,
	})
}

func (h *apiHandler) listClassStudentCandidates(c *gin.Context) {
	items, err := h.classes.ListStudentCandidates()
	if err != nil {
		fail(c, 500, "list class student candidates failed")
		return
	}
	ok(c, items)
}

func (h *apiHandler) createClassStudent(c *gin.Context) {
	classID, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	var req dto.ClassStudentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "invalid request")
		return
	}
	student, err := h.classes.CreateClassStudent(classID, model.Student{
		StudentNo:   req.StudentID,
		StudentName: req.RealName,
	})
	if err != nil {
		switch {
		case service.IsServiceError(err, service.ErrClassNotFound):
			fail(c, 404, "class not found")
		case service.IsServiceError(err, service.ErrInvalidInput):
			fail(c, 400, "invalid request")
		default:
			fail(c, 400, "create class student failed")
		}
		return
	}
	ok(c, student)
}

func (h *apiHandler) importClassStudents(c *gin.Context) {
	classID, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		fail(c, 400, "missing file")
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		fail(c, 400, "open file failed")
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = true

	records, err := reader.ReadAll()
	if err != nil && err != io.EOF {
		fail(c, 400, "invalid csv file")
		return
	}
	if len(records) == 0 {
		fail(c, 400, "empty csv file")
		return
	}
	if len(records[0]) < 2 {
		fail(c, 400, "invalid template header")
		return
	}

	headerStudentID := strings.TrimSpace(strings.TrimPrefix(records[0][0], "\uFEFF"))
	headerRealName := strings.TrimSpace(records[0][1])
	if headerStudentID != "学号" || headerRealName != "姓名" {
		fail(c, 400, "invalid template header")
		return
	}

	students := make([]model.Student, 0, len(records)-1)
	seenStudentIDs := make(map[string]int)

	for index, record := range records[1:] {
		lineNo := index + 2
		if len(record) == 0 {
			continue
		}

		for len(record) < 2 {
			record = append(record, "")
		}

		studentID := strings.TrimSpace(strings.TrimPrefix(record[0], "\uFEFF"))
		realName := strings.TrimSpace(record[1])
		if studentID == "" && realName == "" {
			continue
		}
		if studentID == "" || realName == "" {
			fail(c, 400, fmt.Sprintf("line %d is invalid", lineNo))
			return
		}
		if previousLine, exists := seenStudentIDs[studentID]; exists {
			fail(c, 400, fmt.Sprintf("duplicate student id at line %d, first seen at line %d", lineNo, previousLine))
			return
		}
		seenStudentIDs[studentID] = lineNo
		students = append(students, model.Student{
			StudentNo:   studentID,
			StudentName: realName,
		})
	}

	importedCount, err := h.classes.ImportClassStudents(classID, students)
	if err != nil {
		switch {
		case service.IsServiceError(err, service.ErrClassNotFound):
			fail(c, 404, "class not found")
		case service.IsServiceError(err, service.ErrInvalidInput):
			fail(c, 400, "invalid request")
		default:
			fail(c, 400, "import class students failed")
		}
		return
	}

	ok(c, dto.ClassStudentImportResponse{ImportedCount: importedCount})
}

func (h *apiHandler) updateClassStudent(c *gin.Context) {
	classID, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	studentID, err := parseUintParam(c, "student_id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	var req dto.ClassStudentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "invalid request")
		return
	}
	student, err := h.classes.UpdateClassStudent(classID, studentID, model.Student{
		StudentNo:   req.StudentID,
		StudentName: req.RealName,
	})
	if err != nil {
		switch {
		case service.IsServiceError(err, service.ErrClassNotFound):
			fail(c, 404, "class or class student not found")
		case service.IsServiceError(err, service.ErrInvalidInput):
			fail(c, 400, "invalid request")
		default:
			fail(c, 400, "update class student failed")
		}
		return
	}
	ok(c, student)
}

func (h *apiHandler) deleteClassStudent(c *gin.Context) {
	classID, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	studentID, err := parseUintParam(c, "student_id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	if err := h.classes.DeleteClassStudent(classID, studentID); err != nil {
		switch {
		case service.IsServiceError(err, service.ErrClassNotFound):
			fail(c, 404, "class student not found")
		default:
			fail(c, 400, "delete class student failed")
		}
		return
	}
	ok(c, gin.H{})
}
