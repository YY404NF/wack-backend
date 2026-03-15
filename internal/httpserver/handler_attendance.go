package httpserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"wack-backend/internal/model"
)

func (h *apiHandler) adminAttendanceDashboard(c *gin.Context) {
	type summary struct {
		Present int64 `json:"present"`
		Late    int64 `json:"late"`
		Absent  int64 `json:"absent"`
		Leave   int64 `json:"leave"`
		Unset   int64 `json:"unset"`
	}
	result := summary{}
	base := h.db.Table("attendance_detail").
		Joins("JOIN attendance_check ON attendance_check.id = attendance_detail.attendance_check_id").
		Joins("JOIN course_session ON course_session.id = attendance_check.course_session_id").
		Joins("JOIN course ON course.id = course_session.course_id")
	if weekNo := c.Query("week_no"); weekNo != "" {
		base = base.Where("course_session.week_no = ?", weekNo)
	}
	if term := c.Query("term"); term != "" {
		base = base.Where("course.term = ?", term)
	}
	if courseID := c.Query("course_id"); courseID != "" {
		base = base.Where("course.id = ?", courseID)
	}
	statuses := map[string]int{"present": 1, "late": 2, "absent": 3, "leave": 4, "unset": 0}
	for key, status := range statuses {
		var count int64
		_ = base.Where("attendance_detail.status = ?", status).Count(&count).Error
		switch key {
		case "present":
			result.Present = count
		case "late":
			result.Late = count
		case "absent":
			result.Absent = count
		case "leave":
			result.Leave = count
		case "unset":
			result.Unset = count
		}
	}
	ok(c, result)
}

func (h *apiHandler) adminAttendanceResults(c *gin.Context) {
	page, pageSize := parsePage(c)
	type item struct {
		AttendanceCheckID  uint64 `json:"attendance_check_id"`
		AttendanceDetailID uint64 `json:"attendance_detail_id"`
		CourseID           uint64 `json:"course_id"`
		CourseName         string `json:"course_name"`
		TeacherName        string `json:"teacher_name"`
		WeekNo             int    `json:"week_no"`
		SessionNo          int    `json:"session_no"`
		StudentID          string `json:"student_id"`
		Status             int    `json:"status"`
	}
	query := h.db.Table("attendance_detail").
		Select("attendance_check.id AS attendance_check_id, attendance_detail.id AS attendance_detail_id, course.id AS course_id, course.course_name, course.teacher_name, course_session.week_no, course_session.session_no, user.student_id, attendance_detail.status").
		Joins("JOIN attendance_check ON attendance_check.id = attendance_detail.attendance_check_id").
		Joins("JOIN course_session ON course_session.id = attendance_check.course_session_id").
		Joins("JOIN course ON course.id = course_session.course_id").
		Joins("JOIN user ON user.id = attendance_detail.user_id")
	if courseID := c.Query("course_id"); courseID != "" {
		query = query.Where("course.id = ?", courseID)
	}
	if weekNo := c.Query("week_no"); weekNo != "" {
		query = query.Where("course_session.week_no = ?", weekNo)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("attendance_detail.status = ?", status)
	}
	var items []item
	total, err := paginate(query.Order("attendance_detail.id DESC"), page, pageSize, &items)
	if err != nil {
		fail(c, 500, "load attendance results failed")
		return
	}
	ok(c, pageResult{Items: items, Page: page, PageSize: pageSize, Total: total})
}

func (h *apiHandler) adminFreeTimeCalendar(c *gin.Context) {
	query := h.db.Table("student_free_time").
		Select("student_free_time.id, student_free_time.term, student_free_time.user_id, user.student_id, student_free_time.weekday, student_free_time.section, student_free_time.free_weeks, student_free_time.created_at, student_free_time.updated_at").
		Joins("JOIN user ON user.id = student_free_time.user_id")
	if term := c.Query("term"); term != "" {
		query = query.Where("student_free_time.term = ?", term)
	}
	type item struct {
		ID        uint64    `json:"id"`
		Term      string    `json:"term"`
		UserID    uint64    `json:"user_id"`
		StudentID string    `json:"student_id"`
		Weekday   int       `json:"weekday"`
		Section   int       `json:"section"`
		FreeWeeks string    `json:"free_weeks"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}
	var items []item
	if err := query.Order("weekday, section, user.student_id").Scan(&items).Error; err != nil {
		fail(c, 500, "load free time calendar failed")
		return
	}
	ok(c, items)
}

func (h *apiHandler) studentAvailableCourses(c *gin.Context) {
	type result struct {
		CourseSessionID   uint64     `json:"course_session_id"`
		CourseID          uint64     `json:"course_id"`
		CourseName        string     `json:"course_name"`
		TeacherName       string     `json:"teacher_name"`
		WeekNo            int        `json:"week_no"`
		Weekday           int        `json:"weekday"`
		Section           int        `json:"section"`
		BuildingName      string     `json:"building_name"`
		RoomName          string     `json:"room_name"`
		StartedAt         *time.Time `json:"started_at"`
		CanEnter          bool       `json:"can_enter"`
		EnterDeadline     string     `json:"enter_deadline"`
		AttendanceCheckID *uint64    `json:"attendance_check_id"`
	}

	var sessions []struct {
		model.CourseSession
		CourseName        string     `json:"course_name"`
		TeacherName       string     `json:"teacher_name"`
		AttendanceCheckID *uint64    `json:"attendance_check_id"`
		StartedAt         *time.Time `json:"started_at"`
	}

	weekday := int(time.Now().Weekday())
	if weekday == 0 {
		weekday = 7
	}

	if err := h.db.Table("course_session").
		Select("course_session.*, course.course_name, course.teacher_name, attendance_check.id AS attendance_check_id, attendance_check.started_at").
		Joins("JOIN course ON course.id = course_session.course_id").
		Joins("LEFT JOIN attendance_check ON attendance_check.course_session_id = course_session.id").
		Where("course_session.weekday = ?", weekday).
		Order("course_session.section ASC").
		Scan(&sessions).Error; err != nil {
		fail(c, 500, "load available courses failed")
		return
	}

	items := make([]result, 0, len(sessions))
	for _, session := range sessions {
		deadline, err := h.attendanceDeadline(session.CourseSession, time.Now())
		if err != nil {
			continue
		}
		items = append(items, result{
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
	var req struct {
		CourseSessionID uint64 `json:"course_session_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "invalid request")
		return
	}

	var session model.CourseSession
	if err := h.db.First(&session, req.CourseSessionID).Error; err != nil {
		fail(c, 404, "course session not found")
		return
	}
	if !h.withinDeadline(session, time.Now()) {
		fail(c, 403, "attendance entry deadline passed")
		return
	}

	var attendanceCheck model.AttendanceCheck
	err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&attendanceCheck, "course_session_id = ?", req.CourseSessionID).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			attendanceCheck = model.AttendanceCheck{
				CourseSessionID: req.CourseSessionID,
				StartedByUserID: user.ID,
				StartedAt:       time.Now(),
			}
			if err := tx.Create(&attendanceCheck).Error; err != nil {
				return err
			}
			var students []model.CourseStudent
			if err := tx.Where("course_id = ?", session.CourseID).Find(&students).Error; err != nil {
				return err
			}
			details := make([]model.AttendanceDetail, 0, len(students))
			for _, student := range students {
				details = append(details, model.AttendanceDetail{
					AttendanceCheckID: attendanceCheck.ID,
					UserID:            student.UserID,
					Status:            model.AttendanceUnset,
				})
			}
			if len(details) > 0 {
				return tx.Create(&details).Error
			}
			return nil
		}
		return nil
	})
	if err != nil {
		fail(c, 500, "enter attendance check failed")
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
	var attendanceCheck model.AttendanceCheck
	if err := h.db.First(&attendanceCheck, id).Error; err != nil {
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
	var course model.Course
	var starter model.User
	type detail struct {
		ID                uint64     `json:"id"`
		AttendanceCheckID uint64     `json:"attendance_check_id"`
		UserID            uint64     `json:"user_id"`
		StudentID         string     `json:"student_id"`
		Status            int        `json:"status"`
		StatusSetByUserID *uint64    `json:"status_set_by_user_id"`
		StatusSetAt       *time.Time `json:"status_set_at"`
	}
	var details []detail
	_ = h.db.First(&course, session.CourseID).Error
	_ = h.db.Select("id, student_id").First(&starter, attendanceCheck.StartedByUserID).Error
	_ = h.db.Table("attendance_detail").
		Select("attendance_detail.id, attendance_detail.attendance_check_id, attendance_detail.user_id, user.student_id, attendance_detail.status, attendance_detail.status_set_by_user_id, attendance_detail.status_set_at").
		Joins("JOIN user ON user.id = attendance_detail.user_id").
		Where("attendance_check_id = ?", attendanceCheck.ID).
		Order("attendance_detail.id ASC").
		Scan(&details).Error
	ok(c, gin.H{
		"attendance_check": gin.H{
			"id":                    attendanceCheck.ID,
			"course_session_id":     attendanceCheck.CourseSessionID,
			"started_by_user_id":    attendanceCheck.StartedByUserID,
			"started_by_student_id": starter.StudentID,
			"started_at":            attendanceCheck.StartedAt,
		},
		"course_session": session,
		"course":         course,
		"students":       details,
	})
}

func (h *apiHandler) studentUpdateAttendanceStatus(c *gin.Context) {
	user, _ := currentUser(c)
	id, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	var req struct {
		Status int `json:"status" binding:"required"`
	}
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

	if err := h.updateAttendanceStatus(detail, req.Status, user.ID, false); err != nil {
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
	var attendanceCheck model.AttendanceCheck
	if err := h.db.First(&attendanceCheck, id).Error; err != nil {
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
	ok(c, gin.H{"attendance_check_id": id, "completed": true})
}

func (h *apiHandler) adminGetAttendanceCheck(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	var attendanceCheck model.AttendanceCheck
	if err := h.db.First(&attendanceCheck, id).Error; err != nil {
		fail(c, 404, "attendance check not found")
		return
	}
	type detail struct {
		ID                uint64     `json:"id"`
		AttendanceCheckID uint64     `json:"attendance_check_id"`
		UserID            uint64     `json:"user_id"`
		StudentID         string     `json:"student_id"`
		Status            int        `json:"status"`
		StatusSetByUserID *uint64    `json:"status_set_by_user_id"`
		StatusSetAt       *time.Time `json:"status_set_at"`
	}
	type attendanceCheckPayload struct {
		ID                 uint64    `json:"id"`
		CourseSessionID    uint64    `json:"course_session_id"`
		StartedByUserID    uint64    `json:"started_by_user_id"`
		StartedByStudentID string    `json:"started_by_student_id"`
		StartedAt          time.Time `json:"started_at"`
	}
	var details []detail
	_ = h.db.Table("attendance_detail").
		Select("attendance_detail.id, attendance_detail.attendance_check_id, attendance_detail.user_id, user.student_id, attendance_detail.status, attendance_detail.status_set_by_user_id, attendance_detail.status_set_at").
		Joins("JOIN user ON user.id = attendance_detail.user_id").
		Where("attendance_detail.attendance_check_id = ?", id).
		Order("attendance_detail.id ASC").
		Scan(&details).Error
	var starter model.User
	_ = h.db.Select("id, student_id").First(&starter, attendanceCheck.StartedByUserID).Error
	ok(c, gin.H{
		"attendance_check": attendanceCheckPayload{
			ID:                 attendanceCheck.ID,
			CourseSessionID:    attendanceCheck.CourseSessionID,
			StartedByUserID:    attendanceCheck.StartedByUserID,
			StartedByStudentID: starter.StudentID,
			StartedAt:          attendanceCheck.StartedAt,
		},
		"details": details,
	})
}

func (h *apiHandler) adminUpdateAttendanceStatus(c *gin.Context) {
	user, _ := currentUser(c)
	id, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	var req struct {
		Status int `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "invalid request")
		return
	}
	var detail model.AttendanceDetail
	if err := h.db.First(&detail, id).Error; err != nil {
		fail(c, 404, "attendance detail not found")
		return
	}
	if err := h.updateAttendanceStatus(detail, req.Status, user.ID, true); err != nil {
		fail(c, 500, "update attendance status failed")
		return
	}
	ok(c, gin.H{})
}

func (h *apiHandler) updateAttendanceStatus(detail model.AttendanceDetail, newStatus int, operatorUserID uint64, writeAdminLog bool) error {
	now := time.Now()
	oldStatus := detail.Status
	statusSetBy := operatorUserID
	return h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&detail).Updates(map[string]interface{}{
			"status":                newStatus,
			"status_set_by_user_id": statusSetBy,
			"status_set_at":         now,
		}).Error; err != nil {
			return err
		}
		logItem := model.AttendanceDetailLog{
			AttendanceDetailID: detail.ID,
			AttendanceCheckID:  detail.AttendanceCheckID,
			UserID:             detail.UserID,
			OperatorUserID:     operatorUserID,
			OldStatus:          &oldStatus,
			NewStatus:          newStatus,
			OperationType:      "set_status",
			OperatedAt:         now,
		}
		if err := tx.Create(&logItem).Error; err != nil {
			return err
		}
		if writeAdminLog {
			oldValueBytes, _ := json.Marshal(gin.H{"status": oldStatus})
			newValueBytes, _ := json.Marshal(gin.H{"status": newStatus})
			oldValue := string(oldValueBytes)
			newValue := string(newValueBytes)
			adminLog := model.AdminOperationLog{
				OperatorUserID: operatorUserID,
				TargetTable:    "attendance_detail",
				TargetID:       detail.ID,
				ActionType:     "update",
				OldValue:       &oldValue,
				NewValue:       &newValue,
			}
			if err := tx.Create(&adminLog).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
