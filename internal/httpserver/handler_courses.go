package httpserver

import (
	"strings"

	"github.com/gin-gonic/gin"

	"wack-backend/internal/httpserver/dto"
	"wack-backend/internal/model"
	"wack-backend/internal/query"
)

func (h *apiHandler) listCourses(c *gin.Context) {
	page, pageSize := parsePage(c)
	items, total, err := h.courses.ListCourses(c.Query("term"), c.Query("teacher_name"), strings.TrimSpace(c.Query("keyword")), page, pageSize)
	if err != nil {
		fail(c, 500, "list courses failed")
		return
	}
	ok(c, pageResult[query.CourseListItem]{Items: items, Page: page, PageSize: pageSize, Total: total})
}

func (h *apiHandler) createCourse(c *gin.Context) {
	var course model.Course
	if err := c.ShouldBindJSON(&course); err != nil {
		fail(c, 400, "invalid request")
		return
	}
	course, err := h.courses.CreateCourse(course)
	if err != nil {
		fail(c, 400, "create course failed")
		return
	}
	ok(c, course)
}

func (h *apiHandler) getCourse(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	course, students, classes, sessions, err := h.courses.GetCourse(id)
	if err != nil {
		fail(c, 404, "course not found")
		return
	}
	ok(c, gin.H{
		"course":   course,
		"students": students,
		"classes":  classes,
		"sessions": sessions,
	})
}

func (h *apiHandler) updateCourse(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	var req model.Course
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "invalid request")
		return
	}
	course, err := h.courses.UpdateCourse(id, req)
	if err != nil {
		fail(c, 400, "update course failed")
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
		fail(c, 400, "delete course failed")
		return
	}
	ok(c, gin.H{})
}

func (h *apiHandler) replaceCourseStudents(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	var req dto.ReplaceCourseStudentsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "invalid request")
		return
	}
	students := make([]model.CourseStudent, 0, len(req.Students))
	for _, item := range req.Students {
		students = append(students, model.CourseStudent{
			StudentID: item.StudentID,
			RealName:  item.RealName,
		})
	}
	if err := h.courses.ReplaceCourseStudents(id, students); err != nil {
		fail(c, 400, "replace course students failed")
		return
	}
	ok(c, gin.H{})
}

func (h *apiHandler) replaceCourseClasses(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	var req dto.ReplaceCourseClassesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "invalid request")
		return
	}
	if err := h.courses.ReplaceCourseClasses(id, req.ClassIDs); err != nil {
		fail(c, 400, "replace course classes failed")
		return
	}
	ok(c, gin.H{})
}

func (h *apiHandler) replaceCourseSessions(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	var req dto.ReplaceCourseSessionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "invalid request")
		return
	}
	if err := h.courses.ReplaceCourseSessions(id, req.Sessions); err != nil {
		fail(c, 400, "replace course sessions failed")
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
