package httpserver

import (
	"github.com/gin-gonic/gin"

	"wack-backend/internal/httpserver/dto"
	"wack-backend/internal/model"
	"wack-backend/internal/query"
	"wack-backend/internal/service"
)

func (h *apiHandler) listClasses(c *gin.Context) {
	page, pageSize := parsePage(c)
	classes, total, err := h.classes.ListClasses(page, pageSize)
	if err != nil {
		fail(c, 500, "list classes failed")
		return
	}
	ok(c, pageResult[model.Class]{Items: classes, Page: page, PageSize: pageSize, Total: total})
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
