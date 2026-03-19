package httpserver

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"

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
