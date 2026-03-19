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

type AttendanceStatusInput struct {
	AttendanceRecordID uint64
	Status             int
}

type SubmitAttendanceStatusesResult struct {
	AppliedCount int `json:"applied_count"`
	IgnoredCount int `json:"ignored_count"`
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

func (s *AttendanceService) AvailableCourseGroupLessons(weekday, weekNo int) ([]query.SessionWithCourse, error) {
	return s.attendance.AvailableCourseGroupLessons(weekday, weekNo)
}

func (s *AttendanceService) AvailableCourseGroupLessonsForClass(weekday, weekNo int, classID uint64) ([]query.SessionWithCourse, error) {
	return s.attendance.AvailableCourseGroupLessonsForClass(weekday, weekNo, classID)
}

func (s *AttendanceService) AttendanceRecords(sessionID uint64) ([]query.AttendanceRecordItem, error) {
	var lesson model.CourseGroupLesson
	if err := s.db.First(&lesson, sessionID).Error; err != nil {
		return nil, ErrCourseGroupLessonNotFound
	}
	return s.attendance.AttendanceSessionRecords(sessionID)
}

func (s *AttendanceService) AttendanceRecordLogs(recordID uint64) ([]query.AttendanceRecordLogItem, error) {
	return s.attendance.AttendanceRecordLogsByID(recordID)
}

func (s *AttendanceService) EnterAttendanceSession(courseGroupLessonID uint64, user model.User, withinDeadline func(model.CourseGroupLesson, time.Time) bool) (uint64, error) {
	var lesson model.CourseGroupLesson
	if err := s.db.First(&lesson, courseGroupLessonID).Error; err != nil {
		return 0, ErrCourseGroupLessonNotFound
	}
	if !withinDeadline(lesson, time.Now()) {
		return 0, ErrAttendanceDeadlinePassed
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		var detailCount int64
		if err := tx.Model(&model.AttendanceRecord{}).Where("course_group_lesson_id = ?", lesson.ID).Count(&detailCount).Error; err != nil {
			return err
		}
		if detailCount > 0 {
			return nil
		}

		var group model.CourseGroup
		if err := tx.First(&group, lesson.CourseGroupID).Error; err != nil {
			return err
		}

		type studentRow struct {
			StudentID uint64  `gorm:"column:student_id"`
			ClassID   *uint64 `gorm:"column:class_id"`
		}
		var students []studentRow
		if err := tx.Table("course_group_student").
			Select("course_group_student.student_id, course_group_student.class_id").
			Where("course_group_student.course_group_id = ? AND course_group_student.status = 1 AND student.status = 1", lesson.CourseGroupID).
			Joins("JOIN student ON student.id = course_group_student.student_id").
			Order("student.student_no ASC, student.id ASC").
			Scan(&students).Error; err != nil {
			return err
		}
		records := make([]model.AttendanceRecord, 0, len(students))
		for _, student := range students {
			records = append(records, model.AttendanceRecord{
				TermID:              group.TermID,
				CourseID:            group.CourseID,
				CourseGroupLessonID: lesson.ID,
				StudentID:           student.StudentID,
				ClassID:             student.ClassID,
				AttendanceStatus:    model.AttendanceUnset,
			})
		}
		if len(records) > 0 {
			return tx.Create(&records).Error
		}
		return nil
	})
	return courseGroupLessonID, err
}

func (s *AttendanceService) UpdateAttendanceStatus(detailID uint64, status int, operatorUserID uint64, allowOverwrite bool) error {
	var record model.AttendanceRecord
	if err := s.db.First(&record, detailID).Error; err != nil {
		return ErrAttendanceRecordNotFound
	}
	if !allowOverwrite && record.AttendanceStatus != model.AttendanceUnset {
		return ErrAttendanceRecordLocked
	}
	now := time.Now()
	oldStatus := record.AttendanceStatus
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&record).Updates(map[string]interface{}{
			"attendance_status":  status,
			"updated_by_user_id": operatorUserID,
			"updated_at":         now,
		}).Error; err != nil {
			return err
		}
		return s.audit.logAttendanceStatusChange(tx, record, operatorUserID, oldStatus, status, now)
	})
}

func (s *AttendanceService) SubmitAttendanceStatuses(checkID uint64, operatorUserID uint64, items []AttendanceStatusInput, withinDeadline func(model.CourseGroupLesson, time.Time) bool) (SubmitAttendanceStatusesResult, error) {
	return s.SubmitAttendanceStatusesForClass(checkID, operatorUserID, items, nil, withinDeadline)
}

func (s *AttendanceService) SubmitAttendanceStatusesForClass(checkID uint64, operatorUserID uint64, items []AttendanceStatusInput, classID *uint64, withinDeadline func(model.CourseGroupLesson, time.Time) bool) (SubmitAttendanceStatusesResult, error) {
	var lesson model.CourseGroupLesson
	if err := s.db.First(&lesson, checkID).Error; err != nil {
		return SubmitAttendanceStatusesResult{}, ErrCourseGroupLessonNotFound
	}
	if !withinDeadline(lesson, time.Now()) {
		return SubmitAttendanceStatusesResult{}, ErrAttendanceDeadlinePassed
	}

	result := SubmitAttendanceStatusesResult{}
	now := time.Now()
	err := s.db.Transaction(func(tx *gorm.DB) error {
		for _, item := range items {
			var record model.AttendanceRecord
			if err := tx.Where("id = ? AND course_group_lesson_id = ?", item.AttendanceRecordID, lesson.ID).First(&record).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					result.IgnoredCount++
					continue
				}
				return err
			}
			if classID != nil {
				allowed, err := s.attendance.AttendanceRecordBelongsToClass(checkID, record.ID, *classID)
				if err != nil {
					return err
				}
				if !allowed {
					result.IgnoredCount++
					continue
				}
			}
			if record.AttendanceStatus != model.AttendanceUnset {
				result.IgnoredCount++
				continue
			}

			update := map[string]interface{}{
				"attendance_status":  item.Status,
				"updated_by_user_id": operatorUserID,
				"updated_at":         now,
			}
			txResult := tx.Model(&model.AttendanceRecord{}).
				Where("id = ? AND course_group_lesson_id = ? AND attendance_status = ?", record.ID, lesson.ID, model.AttendanceUnset).
				Updates(update)
			if txResult.Error != nil {
				return txResult.Error
			}
			if txResult.RowsAffected == 0 {
				result.IgnoredCount++
				continue
			}
			if err := s.audit.logAttendanceStatusChange(tx, record, operatorUserID, model.AttendanceUnset, item.Status, now); err != nil {
				return err
			}
			result.AppliedCount++
		}
		return nil
	})
	if err != nil {
		return SubmitAttendanceStatusesResult{}, err
	}
	return result, nil
}

func (s *AttendanceService) GetAttendanceSession(sessionID uint64) (model.CourseGroupLesson, model.Course, []query.AttendanceRecordItem, error) {
	return s.GetAttendanceSessionForClass(sessionID, nil)
}

func (s *AttendanceService) GetAttendanceSessionForClass(sessionID uint64, classID *uint64) (model.CourseGroupLesson, model.Course, []query.AttendanceRecordItem, error) {
	var lesson model.CourseGroupLesson
	if err := s.db.First(&lesson, sessionID).Error; err != nil {
		return model.CourseGroupLesson{}, model.Course{}, nil, ErrCourseGroupLessonNotFound
	}

	var group model.CourseGroup
	if err := s.db.First(&group, lesson.CourseGroupID).Error; err != nil {
		return model.CourseGroupLesson{}, model.Course{}, nil, ErrCourseGroupNotFound
	}

	var course model.Course
	if err := s.db.First(&course, group.CourseID).Error; err != nil {
		return model.CourseGroupLesson{}, model.Course{}, nil, ErrCourseNotFound
	}
	if classID != nil {
		belongs, err := s.attendance.CourseGroupLessonBelongsToClass(lesson.ID, *classID)
		if err != nil {
			return model.CourseGroupLesson{}, model.Course{}, nil, err
		}
		if !belongs {
			return model.CourseGroupLesson{}, model.Course{}, nil, ErrCourseGroupLessonNotFound
		}
	}

	var records []query.AttendanceRecordItem
	var err error
	if classID != nil {
		records, err = s.attendance.AttendanceSessionRecordsForClass(lesson.ID, *classID)
	} else {
		records, err = s.attendance.AttendanceSessionRecords(lesson.ID)
	}
	if err != nil {
		return model.CourseGroupLesson{}, model.Course{}, nil, err
	}

	return lesson, course, records, nil
}

func (s *AttendanceService) AttendanceClassGroups(checkID uint64) ([]query.AttendanceClassGroupItem, error) {
	return s.AttendanceClassGroupsForClass(checkID, nil)
}

func (s *AttendanceService) AttendanceClassGroupsForClass(checkID uint64, classID *uint64) ([]query.AttendanceClassGroupItem, error) {
	var lesson model.CourseGroupLesson
	if err := s.db.First(&lesson, checkID).Error; err != nil {
		return nil, ErrCourseGroupLessonNotFound
	}
	if classID != nil {
		return s.attendance.AttendanceClassGroupsForClass(checkID, *classID)
	}
	return s.attendance.AttendanceClassGroups(checkID)
}

func (s *AttendanceService) CourseGroupLessonBelongsToClass(courseGroupLessonID uint64, classID uint64) (bool, error) {
	return s.attendance.CourseGroupLessonBelongsToClass(courseGroupLessonID, classID)
}

func (s *AttendanceService) CompleteAttendanceSession(checkID uint64, operatorUserID uint64, withinDeadline func(model.CourseGroupLesson, time.Time) bool) error {
	var lesson model.CourseGroupLesson
	if err := s.db.First(&lesson, checkID).Error; err != nil {
		return ErrCourseGroupLessonNotFound
	}
	if !withinDeadline(lesson, time.Now()) {
		return ErrAttendanceDeadlinePassed
	}
	return nil
}

func (s *AttendanceService) AbandonAttendanceSession(checkID uint64, withinDeadline func(model.CourseGroupLesson, time.Time) bool) error {
	var lesson model.CourseGroupLesson
	if err := s.db.First(&lesson, checkID).Error; err != nil {
		return ErrCourseGroupLessonNotFound
	}
	if !withinDeadline(lesson, time.Now()) {
		return ErrAttendanceDeadlinePassed
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		var recordIDs []uint64
		if err := tx.Model(&model.AttendanceRecord{}).
			Where("course_group_lesson_id = ?", lesson.ID).
			Pluck("id", &recordIDs).Error; err != nil {
			return err
		}
		if len(recordIDs) > 0 {
			if err := tx.Where("attendance_record_id IN ?", recordIDs).Delete(&model.AttendanceRecordLog{}).Error; err != nil {
				return err
			}
		}
		if err := tx.Where("course_group_lesson_id = ?", lesson.ID).Delete(&model.AttendanceRecord{}).Error; err != nil {
			return err
		}
		return nil
	})
}
