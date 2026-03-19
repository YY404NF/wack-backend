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

func (h *apiHandler) adminAttendanceSessions(c *gin.Context) {
	page, pageSize := parsePage(c)
	items, total, err := h.attendance.AttendanceSessionSummaries(c.Query("keyword"), c.Query("week_no"), c.Query("status"), page, pageSize)
	if err != nil {
		fail(c, 500, "load attendance sessions failed")
		return
	}
	ok(c, pageResult[query.AttendanceSessionSummaryItem]{Items: items, Page: page, PageSize: pageSize, Total: total})
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
	user, _ := currentUser(c)
	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}

	setting, err := h.systemSettings.GetSystemSetting()
	if err != nil {
		fail(c, 500, "load system settings failed")
		return
	}
	currentWeek, hasWeek := academicWeek(setting.CurrentTermStartDate, now)
	if !hasWeek {
		ok(c, []query.AvailableCourseItem{})
		return
	}

	var sessions []query.SessionWithCourse
	if user.Role == model.RoleCommissioner {
		if user.ManagedClassID == nil {
			ok(c, []query.AvailableCourseItem{})
			return
		}
		sessions, err = h.attendance.AvailableCourseGroupLessonsForClass(weekday, currentWeek, *user.ManagedClassID)
	} else {
		sessions, err = h.attendance.AvailableCourseGroupLessons(weekday, currentWeek)
	}
	if err != nil {
		fail(c, 500, "load available courses failed")
		return
	}

	items := make([]query.AvailableCourseItem, 0, len(sessions))
	for _, session := range sessions {
		lesson := model.CourseGroupLesson{
			ID:            session.ID,
			CourseGroupID: 0,
			WeekNo:        session.WeekNo,
			Weekday:       session.Weekday,
			Section:       session.Section,
			BuildingName:  session.BuildingName,
			RoomName:      session.RoomName,
		}
		canEnter := h.withinDeadline(lesson, now)
		if user.Role != model.RoleCommissioner && !canEnter {
			continue
		}
		deadline, err := h.attendanceDeadline(lesson, now)
		if err != nil {
			continue
		}
		items = append(items, query.AvailableCourseItem{
			CourseGroupLessonID: session.ID,
			CourseID:            session.CourseID,
			CourseName:          session.CourseName,
			TeacherName:         session.TeacherName,
			WeekNo:              session.WeekNo,
			Weekday:             session.Weekday,
			Section:             session.Section,
			BuildingName:        session.BuildingName,
			RoomName:            session.RoomName,
			CanEnter:            canEnter,
			EnterDeadline:       deadline.Format("2006-01-02 15:04:05"),
		})
	}
	ok(c, items)
}

func (h *apiHandler) studentEnterAttendanceSession(c *gin.Context) {
	user, _ := currentUser(c)
	var req dto.EnterAttendanceSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "invalid request")
		return
	}

	if user.Role == model.RoleCommissioner {
		if user.ManagedClassID == nil {
			fail(c, 403, "commissioner class is not configured")
			return
		}
		belongs, err := h.attendance.CourseGroupLessonBelongsToClass(req.CourseGroupLessonID, *user.ManagedClassID)
		if err != nil {
			fail(c, 500, "check course group lesson scope failed")
			return
		}
		if !belongs {
			fail(c, 403, "course group lesson is out of scope")
			return
		}
	}

	lessonID, err := h.attendance.EnterAttendanceSession(req.CourseGroupLessonID, user, h.withinDeadline)
	if err != nil {
		switch {
		case service.IsServiceError(err, service.ErrCourseGroupLessonNotFound):
			fail(c, 404, "course group lesson not found")
		case service.IsServiceError(err, service.ErrAttendanceDeadlinePassed):
			fail(c, 403, "attendance entry deadline passed")
		default:
			fail(c, 500, "enter attendance session failed")
		}
		return
	}
	c.Params = append(c.Params, gin.Param{Key: "id", Value: fmt.Sprintf("%d", lessonID)})
	h.studentGetAttendanceSession(c)
}

func (h *apiHandler) studentGetAttendanceSession(c *gin.Context) {
	user, _ := currentUser(c)
	id, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	var managedClassID *uint64
	if user.Role == model.RoleCommissioner {
		if user.ManagedClassID == nil {
			fail(c, 403, "commissioner class is not configured")
			return
		}
		managedClassID = user.ManagedClassID
	}
	session, course, records, err := h.attendance.GetAttendanceSessionForClass(id, managedClassID)
	if err != nil {
		switch {
		case service.IsServiceError(err, service.ErrCourseGroupLessonNotFound):
			fail(c, 404, "course group lesson not found")
		default:
			fail(c, 500, "load attendance session failed")
		}
		return
	}
	if !h.withinDeadline(session, time.Now()) {
		fail(c, 403, "attendance entry deadline passed")
		return
	}
	classGroups, err := h.attendance.AttendanceClassGroupsForClass(id, managedClassID)
	if err != nil {
		fail(c, 500, "load attendance classes failed")
		return
	}
	ok(c, gin.H{
		"course_group_lesson": session,
		"course":              course,
		"class_groups":        classGroups,
		"students":            records,
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
	if req.Status < model.AttendancePresent || req.Status > model.AttendanceOnLeave {
		fail(c, 400, "invalid status")
		return
	}

	var detail model.AttendanceRecord
	if err := h.db.First(&detail, id).Error; err != nil {
		fail(c, 404, "attendance record not found")
		return
	}
	var lesson model.CourseGroupLesson
	if err := h.db.First(&lesson, detail.CourseGroupLessonID).Error; err != nil {
		fail(c, 404, "course group lesson not found")
		return
	}
	if !h.withinDeadline(lesson, time.Now()) {
		fail(c, 403, "attendance entry deadline passed")
		return
	}

	if err := h.attendance.UpdateAttendanceStatus(id, req.Status, user.ID, false); err != nil {
		switch {
		case service.IsServiceError(err, service.ErrAttendanceRecordNotFound):
			fail(c, 404, "attendance record not found")
		case service.IsServiceError(err, service.ErrAttendanceRecordLocked):
			ok(c, gin.H{"applied": false, "ignored": true})
		default:
			fail(c, 500, "update attendance status failed")
		}
		return
	}
	ok(c, gin.H{})
}

func (h *apiHandler) studentSubmitAttendanceStatuses(c *gin.Context) {
	user, _ := currentUser(c)
	id, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}

	var req dto.SubmitAttendanceStatusesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "invalid request")
		return
	}

	items := make([]service.AttendanceStatusInput, 0, len(req.Items))
	for _, item := range req.Items {
		if item.Status < model.AttendancePresent || item.Status > model.AttendanceOnLeave {
			fail(c, 400, "invalid status")
			return
		}
		items = append(items, service.AttendanceStatusInput{
			StudentRefID: item.StudentRefID,
			Status:       item.Status,
		})
	}

	var managedClassID *uint64
	if user.Role == model.RoleCommissioner {
		if user.ManagedClassID == nil {
			fail(c, 403, "commissioner class is not configured")
			return
		}
		managedClassID = user.ManagedClassID
	}
	result, err := h.attendance.SubmitAttendanceStatusesForClass(id, user.ID, items, managedClassID, h.withinDeadline)
	if err != nil {
		switch {
		case service.IsServiceError(err, service.ErrCourseGroupLessonNotFound):
			fail(c, 404, "course group lesson not found")
		case service.IsServiceError(err, service.ErrAttendanceDeadlinePassed):
			fail(c, 403, "attendance entry deadline passed")
		default:
			fail(c, 500, "submit attendance statuses failed")
		}
		return
	}

	ok(c, result)
}

func (h *apiHandler) studentCompleteAttendanceSession(c *gin.Context) {
	user, _ := currentUser(c)
	id, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	if err := h.attendance.CompleteAttendanceSession(id, user.ID, h.withinDeadline); err != nil {
		switch {
		case service.IsServiceError(err, service.ErrCourseGroupLessonNotFound):
			fail(c, 404, "course group lesson not found")
		case service.IsServiceError(err, service.ErrAttendanceDeadlinePassed):
			fail(c, 403, "attendance entry deadline passed")
		default:
			fail(c, 500, "complete attendance session failed")
		}
		return
	}
	ok(c, gin.H{"course_group_lesson_id": id, "completed": true})
}

func (h *apiHandler) studentAbandonAttendanceSession(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	if err := h.attendance.AbandonAttendanceSession(id, h.withinDeadline); err != nil {
		switch {
		case service.IsServiceError(err, service.ErrCourseGroupLessonNotFound):
			fail(c, 404, "course group lesson not found")
		case service.IsServiceError(err, service.ErrAttendanceDeadlinePassed):
			fail(c, 403, "attendance entry deadline passed")
		default:
			fail(c, 500, "abandon attendance session failed")
		}
		return
	}
	ok(c, gin.H{"course_group_lesson_id": id, "abandoned": true})
}

func (h *apiHandler) adminGetAttendanceSession(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	page, pageSize := parsePage(c)
	session, course, records, total, err := h.attendance.GetAttendanceSessionPage(id, c.Query("keyword"), c.Query("status"), page, pageSize)
	if err != nil {
		switch {
		case service.IsServiceError(err, service.ErrCourseGroupLessonNotFound):
			fail(c, 404, "course group lesson not found")
		default:
			fail(c, 500, "load attendance session failed")
		}
		return
	}
	ok(c, gin.H{
		"course_group_lesson": session,
		"course":              course,
		"attendance_records":  records,
		"page":                page,
		"page_size":           pageSize,
		"total":               total,
	})
}

func (h *apiHandler) adminUpsertAttendanceStatus(c *gin.Context) {
	user, _ := currentUser(c)
	sessionID, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	studentID, err := parseUintParam(c, "student_id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	var req dto.UpdateAttendanceStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "invalid request")
		return
	}
	if req.Status < model.AttendancePresent || req.Status > model.AttendanceOnLeave {
		fail(c, 400, "invalid status")
		return
	}
	if err := h.attendance.UpsertAttendanceStatusForStudent(sessionID, studentID, req.Status, user.ID, true); err != nil {
		fail(c, 500, "update attendance status failed")
		return
	}
	ok(c, gin.H{})
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
	if req.Status < model.AttendancePresent || req.Status > model.AttendanceOnLeave {
		fail(c, 400, "invalid status")
		return
	}
	if err := h.attendance.UpdateAttendanceStatus(id, req.Status, user.ID, true); err != nil {
		fail(c, 500, "update attendance status failed")
		return
	}
	ok(c, gin.H{})
}
