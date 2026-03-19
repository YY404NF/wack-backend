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

type courseRequest struct {
	TermID      uint64 `json:"term_id"`
	Grade       int    `json:"grade"`
	CourseName  string `json:"course_name"`
	TeacherName string `json:"teacher_name"`
}

func (h *apiHandler) listCourses(c *gin.Context) {
	page, pageSize := parsePage(c)
	classID, _ := strconv.ParseUint(c.DefaultQuery("class_id", "0"), 10, 64)
	items, total, err := h.courses.ListCourses(c.Query("term"), c.Query("teacher_name"), strings.TrimSpace(c.Query("keyword")), classID, page, pageSize)
	if err != nil {
		fail(c, 500, "list courses failed")
		return
	}
	ok(c, pageResult[query.CourseListItem]{Items: items, Page: page, PageSize: pageSize, Total: total})
}

func (h *apiHandler) createCourse(c *gin.Context) {
	var req courseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "invalid request")
		return
	}
	course, err := h.courses.CreateCourse(model.Course{
		TermID:      req.TermID,
		Grade:       req.Grade,
		CourseName:  req.CourseName,
		TeacherName: req.TeacherName,
	})
	if err != nil {
		switch {
		case service.IsServiceError(err, service.ErrInvalidInput), service.IsServiceError(err, service.ErrTermNotFound):
			fail(c, 400, "invalid request")
		default:
			fail(c, 400, "create course failed")
		}
		return
	}
	ok(c, course)
}

func (h *apiHandler) listCourseGroups(c *gin.Context) {
	courseID, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	items, err := h.courses.ListCourseGroups(courseID)
	if err != nil {
		switch {
		case service.IsServiceError(err, service.ErrCourseNotFound):
			fail(c, 404, "course not found")
		default:
			fail(c, 500, "list course groups failed")
		}
		return
	}
	ok(c, items)
}

func (h *apiHandler) createCourseGroup(c *gin.Context) {
	courseID, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	group, err := h.courses.CreateCourseGroup(courseID)
	if err != nil {
		switch {
		case service.IsServiceError(err, service.ErrCourseNotFound):
			fail(c, 404, "course not found")
		default:
			fail(c, 400, "create course group failed")
		}
		return
	}
	ok(c, group)
}

func (h *apiHandler) getCourseGroup(c *gin.Context) {
	courseID, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	groupID, err := parseUintParam(c, "group_id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	group, students, lessons, err := h.courses.GetCourseGroup(courseID, groupID)
	if err != nil {
		switch {
		case service.IsServiceError(err, service.ErrCourseGroupNotFound):
			fail(c, 404, "course group not found")
		default:
			fail(c, 500, "load course group failed")
		}
		return
	}
	ok(c, gin.H{
		"course_group": group,
		"students":     students,
		"sessions":     lessons,
	})
}

func (h *apiHandler) listCourseGroupSessions(c *gin.Context) {
	courseID, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	groupID, err := parseUintParam(c, "group_id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	lessons, err := h.courses.GetCourseGroupLessons(courseID, groupID)
	if err != nil {
		switch {
		case service.IsServiceError(err, service.ErrCourseGroupNotFound):
			fail(c, 404, "course group not found")
		default:
			fail(c, 500, "load course group sessions failed")
		}
		return
	}
	ok(c, lessons)
}

func (h *apiHandler) createCourseGroupSession(c *gin.Context) {
	courseID, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	groupID, err := parseUintParam(c, "group_id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	var req dto.CourseGroupLessonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "invalid request")
		return
	}
	lesson, err := h.courses.CreateCourseGroupLesson(courseID, groupID, model.CourseGroupLesson{
		WeekNo:       req.WeekNo,
		Weekday:      req.Weekday,
		Section:      req.Section,
		BuildingName: req.BuildingName,
		RoomName:     req.RoomName,
	})
	if err != nil {
		switch {
		case service.IsServiceError(err, service.ErrCourseGroupNotFound):
			fail(c, 404, "course group not found")
		case service.IsServiceError(err, service.ErrInvalidInput):
			fail(c, 400, "invalid request")
		default:
			fail(c, 400, "create course group session failed")
		}
		return
	}
	ok(c, lesson)
}

func (h *apiHandler) updateCourseGroupSession(c *gin.Context) {
	courseID, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	groupID, err := parseUintParam(c, "group_id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	sessionID, err := parseUintParam(c, "session_id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	var req dto.CourseGroupLessonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "invalid request")
		return
	}
	lesson, err := h.courses.UpdateCourseGroupLesson(courseID, groupID, sessionID, model.CourseGroupLesson{
		WeekNo:       req.WeekNo,
		Weekday:      req.Weekday,
		Section:      req.Section,
		BuildingName: req.BuildingName,
		RoomName:     req.RoomName,
	})
	if err != nil {
		switch {
		case service.IsServiceError(err, service.ErrCourseGroupNotFound):
			fail(c, 404, "course group not found")
		case service.IsServiceError(err, service.ErrCourseGroupLessonNotFound):
			fail(c, 404, "course group session not found")
		case service.IsServiceError(err, service.ErrInvalidInput):
			fail(c, 400, "invalid request")
		default:
			fail(c, 400, "update course group session failed")
		}
		return
	}
	ok(c, lesson)
}

func (h *apiHandler) deleteCourseGroupSession(c *gin.Context) {
	courseID, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	groupID, err := parseUintParam(c, "group_id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	sessionID, err := parseUintParam(c, "session_id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	if err := h.courses.DeleteCourseGroupLesson(courseID, groupID, sessionID); err != nil {
		switch {
		case service.IsServiceError(err, service.ErrCourseGroupNotFound):
			fail(c, 404, "course group not found")
		case service.IsServiceError(err, service.ErrCourseGroupLessonNotFound):
			fail(c, 404, "course group session not found")
		default:
			fail(c, 400, "delete course group session failed")
		}
		return
	}
	ok(c, gin.H{})
}

func (h *apiHandler) deleteCourseGroup(c *gin.Context) {
	courseID, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	groupID, err := parseUintParam(c, "group_id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	if err := h.courses.DeleteCourseGroup(courseID, groupID); err != nil {
		switch {
		case service.IsServiceError(err, service.ErrCourseGroupNotFound):
			fail(c, 404, "course group not found")
		default:
			fail(c, 400, "delete course group failed")
		}
		return
	}
	ok(c, gin.H{})
}

func (h *apiHandler) listCourseGroupStudents(c *gin.Context) {
	courseID, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	groupID, err := parseUintParam(c, "group_id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	items, err := h.courses.GetCourseGroupStudents(courseID, groupID)
	if err != nil {
		switch {
		case service.IsServiceError(err, service.ErrCourseGroupNotFound):
			fail(c, 404, "course group not found")
		default:
			fail(c, 500, "load course group students failed")
		}
		return
	}
	ok(c, items)
}

func (h *apiHandler) listAvailableCourseGroupClasses(c *gin.Context) {
	courseID, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	groupID, err := parseUintParam(c, "group_id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	items, err := h.courses.ListAvailableCourseGroupClasses(courseID, groupID, c.Query("keyword"))
	if err != nil {
		switch {
		case service.IsServiceError(err, service.ErrCourseGroupNotFound):
			fail(c, 404, "course group not found")
		default:
			fail(c, 500, "load available course group classes failed")
		}
		return
	}
	ok(c, items)
}

func (h *apiHandler) addCourseGroupClasses(c *gin.Context) {
	courseID, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	groupID, err := parseUintParam(c, "group_id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	var req dto.CourseGroupClassesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "invalid request")
		return
	}
	if err := h.courses.AddCourseGroupClasses(courseID, groupID, req.ClassIDs); err != nil {
		switch {
		case service.IsServiceError(err, service.ErrCourseGroupNotFound):
			fail(c, 404, "course group not found")
		case service.IsServiceError(err, service.ErrClassNotFound):
			fail(c, 404, "class not found")
		case service.IsServiceError(err, service.ErrInvalidInput):
			fail(c, 400, "invalid request")
		default:
			fail(c, 400, "add course group classes failed")
		}
		return
	}
	ok(c, gin.H{})
}

func (h *apiHandler) removeCourseGroupClass(c *gin.Context) {
	courseID, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	groupID, err := parseUintParam(c, "group_id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	classID, err := parseUintParam(c, "class_id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	if err := h.courses.RemoveCourseGroupClass(courseID, groupID, classID); err != nil {
		switch {
		case service.IsServiceError(err, service.ErrCourseGroupNotFound):
			fail(c, 404, "course group not found")
		case service.IsServiceError(err, service.ErrClassNotFound):
			fail(c, 404, "class not found")
		default:
			fail(c, 400, "remove course group class failed")
		}
		return
	}
	ok(c, gin.H{})
}

func (h *apiHandler) listAvailableCourseGroupStudents(c *gin.Context) {
	courseID, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	groupID, err := parseUintParam(c, "group_id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	items, err := h.courses.ListAvailableCourseGroupStudents(courseID, groupID, c.Query("keyword"))
	if err != nil {
		switch {
		case service.IsServiceError(err, service.ErrCourseGroupNotFound):
			fail(c, 404, "course group not found")
		default:
			fail(c, 500, "load available course group students failed")
		}
		return
	}
	ok(c, items)
}

func (h *apiHandler) addCourseGroupStudents(c *gin.Context) {
	courseID, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	groupID, err := parseUintParam(c, "group_id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	var req dto.CourseGroupStudentsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "invalid request")
		return
	}
	if err := h.courses.AddCourseGroupStudents(courseID, groupID, req.StudentIDs); err != nil {
		switch {
		case service.IsServiceError(err, service.ErrCourseGroupNotFound):
			fail(c, 404, "course group not found")
		case service.IsServiceError(err, service.ErrStudentNotFound):
			fail(c, 404, "student not found")
		case service.IsServiceError(err, service.ErrInvalidInput):
			fail(c, 400, "invalid request")
		default:
			fail(c, 400, "add course group students failed")
		}
		return
	}
	ok(c, gin.H{})
}

func (h *apiHandler) removeCourseGroupStudent(c *gin.Context) {
	courseID, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	groupID, err := parseUintParam(c, "group_id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	studentID, err := parseUintParam(c, "student_id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	if err := h.courses.RemoveCourseGroupStudent(courseID, groupID, studentID); err != nil {
		switch {
		case service.IsServiceError(err, service.ErrCourseGroupNotFound):
			fail(c, 404, "course group not found")
		case service.IsServiceError(err, service.ErrStudentNotFound):
			fail(c, 404, "student not found")
		default:
			fail(c, 400, "remove course group student failed")
		}
		return
	}
	ok(c, gin.H{})
}

func (h *apiHandler) updateCourse(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	var req courseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "invalid request")
		return
	}
	course, err := h.courses.UpdateCourse(id, model.Course{
		TermID:      req.TermID,
		Grade:       req.Grade,
		CourseName:  req.CourseName,
		TeacherName: req.TeacherName,
		Status:      1,
	})
	if err != nil {
		switch {
		case service.IsServiceError(err, service.ErrCourseNotFound):
			fail(c, 404, "course not found")
		case service.IsServiceError(err, service.ErrInvalidInput), service.IsServiceError(err, service.ErrTermNotFound):
			fail(c, 400, "invalid request")
		default:
			fail(c, 400, "update course failed")
		}
		return
	}
	ok(c, course)
}

func (h *apiHandler) deleteCourse(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	if err := h.courses.DeleteCourse(id); err != nil {
		switch {
		case service.IsServiceError(err, service.ErrCourseNotFound):
			fail(c, 404, "course not found")
		default:
			fail(c, 400, "delete course failed")
		}
		return
	}
	ok(c, gin.H{})
}

func (h *apiHandler) adminCourseCalendar(c *gin.Context) {
	items, err := h.courses.CourseCalendar(c.Query("week_no"), c.Query("term"))
	if err != nil {
		fail(c, 500, "load course calendar failed")
		return
	}
	ok(c, items)
}
