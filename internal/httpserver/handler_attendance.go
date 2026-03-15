package httpserver

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"

	"wack-backend/internal/httpserver/dto"
	"wack-backend/internal/model"
	"wack-backend/internal/query"
	"wack-backend/internal/service"
)

func (h *apiHandler) adminAttendanceDashboard(c *gin.Context) {
	result, err := h.attendance.DashboardSummary(c.Query("week_no"), c.Query("term"), c.Query("course_id"))
	if err != nil {
		fail(c, 500, "load attendance dashboard failed")
		return
	}
	ok(c, result)
}

func (h *apiHandler) adminAttendanceResults(c *gin.Context) {
	page, pageSize := parsePage(c)
	items, total, err := h.attendance.AttendanceResults(c.Query("week_no"), c.Query("course_id"), c.Query("status"), page, pageSize)
	if err != nil {
		fail(c, 500, "load attendance results failed")
		return
	}
	ok(c, pageResult[query.AttendanceResultItem]{Items: items, Page: page, PageSize: pageSize, Total: total})
}

func (h *apiHandler) adminFreeTimeCalendar(c *gin.Context) {
	items, err := h.freeTimes.FreeTimeCalendar(c.Query("term"))
	if err != nil {
		fail(c, 500, "load free time calendar failed")
		return
	}
	ok(c, items)
}

func (h *apiHandler) studentAvailableCourses(c *gin.Context) {
	weekday := int(time.Now().Weekday())
	if weekday == 0 {
		weekday = 7
	}

	sessions, err := h.attendance.AvailableSessions(weekday)
	if err != nil {
		fail(c, 500, "load available courses failed")
		return
	}

	items := make([]query.AvailableCourseItem, 0, len(sessions))
	for _, session := range sessions {
		deadline, err := h.attendanceDeadline(session.CourseSession, time.Now())
		if err != nil {
			continue
		}
		items = append(items, query.AvailableCourseItem{
			CourseSessionID:   session.ID,
			CourseID:          session.CourseID,
			CourseName:        session.CourseName,
			TeacherName:       session.TeacherName,
			WeekNo:            session.WeekNo,
			Weekday:           session.Weekday,
			Section:           session.Section,
			BuildingName:      session.BuildingName,
			RoomName:          session.RoomName,
			StartedAt:         session.StartedAt,
			CanEnter:          time.Now().Before(deadline),
			EnterDeadline:     deadline.Format("2006-01-02 15:04:05"),
			AttendanceCheckID: session.AttendanceCheckID,
		})
	}
	ok(c, items)
}

func (h *apiHandler) studentEnterAttendanceCheck(c *gin.Context) {
	user, _ := currentUser(c)
	var req dto.EnterAttendanceCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "invalid request")
		return
	}

	attendanceCheck, err := h.attendance.EnterAttendanceCheck(req.CourseSessionID, user, h.withinDeadline)
	if err != nil {
		switch {
		case service.IsServiceError(err, service.ErrCourseSessionNotFound):
			fail(c, 404, "course session not found")
		case service.IsServiceError(err, service.ErrAttendanceDeadlinePassed):
			fail(c, 403, "attendance entry deadline passed")
		default:
			fail(c, 500, "enter attendance check failed")
		}
		return
	}
	c.Params = append(c.Params, gin.Param{Key: "id", Value: fmt.Sprintf("%d", attendanceCheck.ID)})
	h.studentGetAttendanceCheck(c)
}

func (h *apiHandler) studentGetAttendanceCheck(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	attendanceCheck, session, course, details, err := h.attendance.GetAttendanceCheck(id)
	if err != nil {
		switch {
		case service.IsServiceError(err, service.ErrAttendanceCheckNotFound):
			fail(c, 404, "attendance check not found")
		case service.IsServiceError(err, service.ErrCourseSessionNotFound):
			fail(c, 404, "course session not found")
		default:
			fail(c, 500, "load attendance check failed")
		}
		return
	}
	if !h.withinDeadline(session, time.Now()) {
		fail(c, 403, "attendance entry deadline passed")
		return
	}
	ok(c, gin.H{
		"attendance_check": attendanceCheck,
		"course_session":   session,
		"course":           course,
		"students":         details,
	})
}

func (h *apiHandler) studentUpdateAttendanceStatus(c *gin.Context) {
	user, _ := currentUser(c)
	id, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	var req dto.UpdateAttendanceStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "invalid request")
		return
	}
	if req.Status < model.AttendanceUnset || req.Status > model.AttendanceOnLeave {
		fail(c, 400, "invalid status")
		return
	}

	var detail model.AttendanceDetail
	if err := h.db.First(&detail, id).Error; err != nil {
		fail(c, 404, "attendance detail not found")
		return
	}
	var attendanceCheck model.AttendanceCheck
	if err := h.db.First(&attendanceCheck, detail.AttendanceCheckID).Error; err != nil {
		fail(c, 404, "attendance check not found")
		return
	}
	var session model.CourseSession
	if err := h.db.First(&session, attendanceCheck.CourseSessionID).Error; err != nil {
		fail(c, 404, "course session not found")
		return
	}
	if !h.withinDeadline(session, time.Now()) {
		fail(c, 403, "attendance entry deadline passed")
		return
	}

	if err := h.attendance.UpdateAttendanceStatus(id, req.Status, user.ID, false); err != nil {
		fail(c, 500, "update attendance status failed")
		return
	}
	ok(c, gin.H{})
}

func (h *apiHandler) studentCompleteAttendanceCheck(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	if err := h.attendance.CompleteAttendanceCheck(id, h.withinDeadline); err != nil {
		switch {
		case service.IsServiceError(err, service.ErrAttendanceCheckNotFound):
			fail(c, 404, "attendance check not found")
		case service.IsServiceError(err, service.ErrCourseSessionNotFound):
			fail(c, 404, "course session not found")
		case service.IsServiceError(err, service.ErrAttendanceDeadlinePassed):
			fail(c, 403, "attendance entry deadline passed")
		default:
			fail(c, 500, "complete attendance check failed")
		}
		return
	}
	ok(c, gin.H{"attendance_check_id": id, "completed": true})
}

func (h *apiHandler) adminGetAttendanceCheck(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	attendanceCheck, _, _, details, err := h.attendance.GetAttendanceCheck(id)
	if err != nil {
		switch {
		case service.IsServiceError(err, service.ErrAttendanceCheckNotFound):
			fail(c, 404, "attendance check not found")
		default:
			fail(c, 500, "load attendance check failed")
		}
		return
	}
	ok(c, gin.H{
		"attendance_check": attendanceCheck,
		"details":          details,
	})
}

func (h *apiHandler) adminUpdateAttendanceStatus(c *gin.Context) {
	user, _ := currentUser(c)
	id, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	var req dto.UpdateAttendanceStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "invalid request")
		return
	}
	if err := h.attendance.UpdateAttendanceStatus(id, req.Status, user.ID, true); err != nil {
		fail(c, 500, "update attendance status failed")
		return
	}
	ok(c, gin.H{})
}
