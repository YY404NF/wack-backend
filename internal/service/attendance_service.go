package service

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"wack-backend/internal/model"
	"wack-backend/internal/query"
)

type AttendanceService struct {
	db         *gorm.DB
	attendance *query.AttendanceQuery
	audit      *auditLogger
}

func NewAttendanceService(db *gorm.DB) *AttendanceService {
	return &AttendanceService{db: db, attendance: query.NewAttendanceQuery(db), audit: newAuditLogger()}
}

func (s *AttendanceService) DashboardSummary(weekNo, term, courseID string) (query.AttendanceDashboardSummary, error) {
	return s.attendance.DashboardSummary(weekNo, term, courseID)
}

func (s *AttendanceService) AttendanceResults(weekNo, courseID, status string, page, pageSize int) ([]query.AttendanceResultItem, int64, error) {
	return s.attendance.AttendanceResults(weekNo, courseID, status, page, pageSize)
}

func (s *AttendanceService) AvailableSessions(weekday, weekNo int) ([]query.SessionWithCourse, error) {
	return s.attendance.AvailableSessions(weekday, weekNo)
}

func (s *AttendanceService) AttendanceDetails(checkID uint64) ([]query.AttendanceDetailItem, error) {
	return s.attendance.AttendanceCheckDetails(checkID)
}

func (s *AttendanceService) AttendanceDetailLogs(detailID uint64) ([]query.AttendanceDetailLogItem, error) {
	return s.attendance.AttendanceDetailLogs(detailID)
}

func (s *AttendanceService) EnterAttendanceCheck(courseSessionID uint64, user model.User, withinDeadline func(model.CourseSession, time.Time) bool) (model.AttendanceCheck, error) {
	var session model.CourseSession
	if err := s.db.First(&session, courseSessionID).Error; err != nil {
		return model.AttendanceCheck{}, ErrCourseSessionNotFound
	}
	if !withinDeadline(session, time.Now()) {
		return model.AttendanceCheck{}, ErrAttendanceDeadlinePassed
	}

	var attendanceCheck model.AttendanceCheck
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&attendanceCheck, "course_session_id = ?", courseSessionID).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			attendanceCheck = model.AttendanceCheck{
				CourseSessionID: courseSessionID,
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
					StudentID:         student.StudentID,
					RealName:          student.RealName,
					Status:            model.AttendanceUnset,
				})
			}
			if len(details) > 0 {
				return tx.Create(&details).Error
			}
		}
		return nil
	})
	return attendanceCheck, err
}

func (s *AttendanceService) UpdateAttendanceStatus(detailID uint64, status int, operatorUserID uint64, writeAdminLog bool) error {
	var detail model.AttendanceDetail
	if err := s.db.First(&detail, detailID).Error; err != nil {
		return ErrAttendanceDetailNotFound
	}
	now := time.Now()
	oldStatus := detail.Status
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&detail).Updates(map[string]interface{}{
			"status":                status,
			"status_set_by_user_id": operatorUserID,
			"status_set_at":         now,
		}).Error; err != nil {
			return err
		}
		return s.audit.logAttendanceStatusChange(tx, detail, operatorUserID, oldStatus, status, now, writeAdminLog)
	})
}

func (s *AttendanceService) GetAttendanceCheck(checkID uint64) (query.AttendanceCheckPayload, model.CourseSession, model.Course, []query.AttendanceDetailItem, error) {
	var attendanceCheck model.AttendanceCheck
	if err := s.db.First(&attendanceCheck, checkID).Error; err != nil {
		return query.AttendanceCheckPayload{}, model.CourseSession{}, model.Course{}, nil, ErrAttendanceCheckNotFound
	}

	var session model.CourseSession
	if err := s.db.First(&session, attendanceCheck.CourseSessionID).Error; err != nil {
		return query.AttendanceCheckPayload{}, model.CourseSession{}, model.Course{}, nil, ErrCourseSessionNotFound
	}

	var course model.Course
	if err := s.db.First(&course, session.CourseID).Error; err != nil {
		return query.AttendanceCheckPayload{}, model.CourseSession{}, model.Course{}, nil, ErrCourseNotFound
	}

	var starter model.User
	if err := s.db.Select("id, student_id").First(&starter, attendanceCheck.StartedByUserID).Error; err != nil {
		return query.AttendanceCheckPayload{}, model.CourseSession{}, model.Course{}, nil, ErrUserNotFound
	}

	details, err := s.attendance.AttendanceCheckDetails(attendanceCheck.ID)
	if err != nil {
		return query.AttendanceCheckPayload{}, model.CourseSession{}, model.Course{}, nil, err
	}

	return query.AttendanceCheckPayload{
		ID:                 attendanceCheck.ID,
		CourseSessionID:    attendanceCheck.CourseSessionID,
		StartedByUserID:    attendanceCheck.StartedByUserID,
		StartedByStudentID: starter.StudentID,
		StartedAt:          attendanceCheck.StartedAt,
	}, session, course, details, nil
}

func (s *AttendanceService) CompleteAttendanceCheck(checkID uint64, withinDeadline func(model.CourseSession, time.Time) bool) error {
	var attendanceCheck model.AttendanceCheck
	if err := s.db.First(&attendanceCheck, checkID).Error; err != nil {
		return ErrAttendanceCheckNotFound
	}

	var session model.CourseSession
	if err := s.db.First(&session, attendanceCheck.CourseSessionID).Error; err != nil {
		return ErrCourseSessionNotFound
	}
	if !withinDeadline(session, time.Now()) {
		return ErrAttendanceDeadlinePassed
	}
	return nil
}
