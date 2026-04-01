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
		Page:      page,
		PageSize:  pageSize,
		ClassID:   classID,
		Keyword:   c.Query("keyword"),
		StudentID: c.Query("student_id"),
		RealName:  c.Query("real_name"),
		ClassName: c.Query("class_name"),
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
		Page:      page,
		PageSize:  pageSize,
		ClassID:   classID,
		Keyword:   c.Query("keyword"),
		StudentID: c.Query("student_id"),
		RealName:  c.Query("real_name"),
		ClassName: c.Query("class_name"),
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
