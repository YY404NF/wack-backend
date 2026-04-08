package service

import (
	"errors"
	"math"
	"strings"
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
	StudentRefID uint64
	Status       int
}

type SubmitAttendanceStatusesResult struct {
	AcceptedItems []uint64 `json:"accepted_items"`
	IgnoredItems  []uint64 `json:"ignored_items"`
	AppliedCount  int      `json:"applied_count"`
	IgnoredCount  int      `json:"ignored_count"`
}

type AdminBulkUpdateAttendanceStatusesResult struct {
	AppliedItems []uint64 `json:"applied_items"`
	FailedItems  []uint64 `json:"failed_items"`
	AppliedCount int      `json:"applied_count"`
	FailedCount  int      `json:"failed_count"`
}

func NewAttendanceService(db *gorm.DB) *AttendanceService {
	return &AttendanceService{db: db, attendance: query.NewAttendanceQuery(db), audit: newAuditLogger()}
}

func paginateOverviewItems[T any](items []T, offset, limit int) ([]T, int64, bool) {
	total := len(items)
	if limit <= 0 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	if offset >= total {
		return []T{}, int64(total), false
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return items[offset:end], int64(total), end < total
}

func reverseOverviewItems[T any](items []T) []T {
	if len(items) == 0 {
		return []T{}
	}
	reversed := make([]T, len(items))
	for index := range items {
		reversed[index] = items[len(items)-1-index]
	}
	return reversed
}

func overviewRateRange[T any](items []T, getter func(T) float64) (float64, float64) {
	if len(items) == 0 {
		return 0, 0
	}
	minRate := getter(items[0])
	maxRate := minRate
	for _, item := range items[1:] {
		rate := getter(item)
		if rate < minRate {
			minRate = rate
		}
		if rate > maxRate {
			maxRate = rate
		}
	}
	return minRate, maxRate
}

func overviewDisplayRateKey(rate float64) int {
	return int(math.Round(rate * 1000))
}

func uniqueUint64s(values []uint64) []uint64 {
	if len(values) == 0 {
		return []uint64{}
	}
	result := make([]uint64, 0, len(values))
	seen := make(map[uint64]struct{}, len(values))
	for _, value := range values {
		if value == 0 {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func assignDenseOverviewRanks[T any](items []T, getter func(T) float64, setter func(*T, int)) {
	currentRank := 0
	lastKey := 0
	first := true
	for index := range items {
		key := overviewDisplayRateKey(getter(items[index]))
		if first || key != lastKey {
			currentRank += 1
			lastKey = key
			first = false
		}
		setter(&items[index], currentRank)
	}
}

func (s *AttendanceService) AdminOverview() (query.AdminOverviewData, error) {
	term, err := s.resolveActiveTerm(time.Now())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return query.AdminOverviewData{
				Term:                   "",
				CourseRankings:         []query.OverviewCourseRankingItem{},
				ClassRankings:          []query.OverviewClassRankingItem{},
				StudentRankings:        []query.OverviewStudentRankingItem{},
				RecentSessions:         []query.OverviewRecentSessionItem{},
				RecentAbnormalStudents: []query.OverviewRecentAbnormalItem{},
			}, nil
		}
		return query.AdminOverviewData{}, err
	}

	courseRankings, err := s.attendance.OverviewCourseRankings(term.Name)
	if err != nil {
		return query.AdminOverviewData{}, err
	}
	assignDenseOverviewRanks(courseRankings,
		func(item query.OverviewCourseRankingItem) float64 { return item.AttendanceRate },
		func(item *query.OverviewCourseRankingItem, rank int) { item.Rank = rank },
	)
	classRankings, err := s.attendance.OverviewClassRankings(term.Name)
	if err != nil {
		return query.AdminOverviewData{}, err
	}
	assignDenseOverviewRanks(classRankings,
		func(item query.OverviewClassRankingItem) float64 { return item.AttendanceRate },
		func(item *query.OverviewClassRankingItem, rank int) { item.Rank = rank },
	)
	studentRankings, err := s.attendance.OverviewStudentRankings(term.Name)
	if err != nil {
		return query.AdminOverviewData{}, err
	}
	assignDenseOverviewRanks(studentRankings,
		func(item query.OverviewStudentRankingItem) float64 { return item.AttendanceRate },
		func(item *query.OverviewStudentRankingItem, rank int) { item.Rank = rank },
	)
	recentSessions, err := s.attendance.OverviewRecentSessions(term.Name)
	if err != nil {
		return query.AdminOverviewData{}, err
	}
	recentAbnormalStudents, err := s.attendance.OverviewRecentAbnormalStudents(term.Name)
	if err != nil {
		return query.AdminOverviewData{}, err
	}

	return query.AdminOverviewData{
		Term:                term.Name,
		CourseRankings:      courseRankings,
		CourseRankingsTotal: int64(len(courseRankings)),
		CourseRankingsMinRate: func() float64 {
			minRate, _ := overviewRateRange(courseRankings, func(item query.OverviewCourseRankingItem) float64 { return item.AttendanceRate })
			return minRate
		}(),
		CourseRankingsMaxRate: func() float64 {
			_, maxRate := overviewRateRange(courseRankings, func(item query.OverviewCourseRankingItem) float64 { return item.AttendanceRate })
			return maxRate
		}(),
		ClassRankings:      classRankings,
		ClassRankingsTotal: int64(len(classRankings)),
		ClassRankingsMinRate: func() float64 {
			minRate, _ := overviewRateRange(classRankings, func(item query.OverviewClassRankingItem) float64 { return item.AttendanceRate })
			return minRate
		}(),
		ClassRankingsMaxRate: func() float64 {
			_, maxRate := overviewRateRange(classRankings, func(item query.OverviewClassRankingItem) float64 { return item.AttendanceRate })
			return maxRate
		}(),
		StudentRankings:      studentRankings,
		StudentRankingsTotal: int64(len(studentRankings)),
		StudentRankingsMinRate: func() float64 {
			minRate, _ := overviewRateRange(studentRankings, func(item query.OverviewStudentRankingItem) float64 { return item.AttendanceRate })
			return minRate
		}(),
		StudentRankingsMaxRate: func() float64 {
			_, maxRate := overviewRateRange(studentRankings, func(item query.OverviewStudentRankingItem) float64 { return item.AttendanceRate })
			return maxRate
		}(),
		RecentSessions:      recentSessions,
		RecentSessionsTotal: int64(len(recentSessions)),
		RecentSessionsMinRate: func() float64 {
			minRate, _ := overviewRateRange(recentSessions, func(item query.OverviewRecentSessionItem) float64 { return item.AttendanceRate })
			return minRate
		}(),
		RecentSessionsMaxRate: func() float64 {
			_, maxRate := overviewRateRange(recentSessions, func(item query.OverviewRecentSessionItem) float64 { return item.AttendanceRate })
			return maxRate
		}(),
		RecentAbnormalStudents: recentAbnormalStudents,
		RecentAbnormalTotal:    int64(len(recentAbnormalStudents)),
	}, nil
}

func (s *AttendanceService) AdminOverviewSection(section string, offset, limit int, order string) (query.AdminOverviewData, error) {
	term, err := s.resolveActiveTerm(time.Now())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return query.AdminOverviewData{
				Term:                   "",
				CourseRankings:         []query.OverviewCourseRankingItem{},
				ClassRankings:          []query.OverviewClassRankingItem{},
				StudentRankings:        []query.OverviewStudentRankingItem{},
				RecentSessions:         []query.OverviewRecentSessionItem{},
				RecentAbnormalStudents: []query.OverviewRecentAbnormalItem{},
			}, nil
		}
		return query.AdminOverviewData{}, err
	}

	normalizedOrder := strings.ToLower(strings.TrimSpace(order))
	ascending := normalizedOrder == "asc"

	result := query.AdminOverviewData{Term: term.Name}
	switch section {
	case "course_rankings":
		items, err := s.attendance.OverviewCourseRankings(term.Name)
		if err != nil {
			return query.AdminOverviewData{}, err
		}
		assignDenseOverviewRanks(items,
			func(item query.OverviewCourseRankingItem) float64 { return item.AttendanceRate },
			func(item *query.OverviewCourseRankingItem, rank int) { item.Rank = rank },
		)
		if ascending {
			items = reverseOverviewItems(items)
		}
		result.CourseRankings, result.CourseRankingsTotal, result.CourseRankingsHasMore = paginateOverviewItems(items, offset, limit)
		result.CourseRankingsMinRate, result.CourseRankingsMaxRate = overviewRateRange(items, func(item query.OverviewCourseRankingItem) float64 { return item.AttendanceRate })
	case "class_rankings":
		items, err := s.attendance.OverviewClassRankings(term.Name)
		if err != nil {
			return query.AdminOverviewData{}, err
		}
		assignDenseOverviewRanks(items,
			func(item query.OverviewClassRankingItem) float64 { return item.AttendanceRate },
			func(item *query.OverviewClassRankingItem, rank int) { item.Rank = rank },
		)
		if ascending {
			items = reverseOverviewItems(items)
		}
		result.ClassRankings, result.ClassRankingsTotal, result.ClassRankingsHasMore = paginateOverviewItems(items, offset, limit)
		result.ClassRankingsMinRate, result.ClassRankingsMaxRate = overviewRateRange(items, func(item query.OverviewClassRankingItem) float64 { return item.AttendanceRate })
	case "student_rankings":
		items, err := s.attendance.OverviewStudentRankings(term.Name)
		if err != nil {
			return query.AdminOverviewData{}, err
		}
		assignDenseOverviewRanks(items,
			func(item query.OverviewStudentRankingItem) float64 { return item.AttendanceRate },
			func(item *query.OverviewStudentRankingItem, rank int) { item.Rank = rank },
		)
		if ascending {
			items = reverseOverviewItems(items)
		}
		result.StudentRankings, result.StudentRankingsTotal, result.StudentRankingsHasMore = paginateOverviewItems(items, offset, limit)
		result.StudentRankingsMinRate, result.StudentRankingsMaxRate = overviewRateRange(items, func(item query.OverviewStudentRankingItem) float64 { return item.AttendanceRate })
	case "recent_sessions":
		items, err := s.attendance.OverviewRecentSessions(term.Name)
		if err != nil {
			return query.AdminOverviewData{}, err
		}
		result.RecentSessions, result.RecentSessionsTotal, result.RecentSessionsHasMore = paginateOverviewItems(items, offset, limit)
		result.RecentSessionsMinRate, result.RecentSessionsMaxRate = overviewRateRange(items, func(item query.OverviewRecentSessionItem) float64 { return item.AttendanceRate })
	case "recent_abnormal_students":
		items, err := s.attendance.OverviewRecentAbnormalStudents(term.Name)
		if err != nil {
			return query.AdminOverviewData{}, err
		}
		result.RecentAbnormalStudents, result.RecentAbnormalTotal, result.RecentAbnormalHasMore = paginateOverviewItems(items, offset, limit)
	default:
		return s.AdminOverview()
	}
	return result, nil
}

func (s *AttendanceService) DashboardSummary(weekNo, term, courseID string) (query.AttendanceDashboardSummary, error) {
	return s.attendance.DashboardSummary(weekNo, term, courseID)
}

func (s *AttendanceService) AttendanceResults(weekNo, courseID, status string, page, pageSize int) ([]query.AttendanceResultItem, int64, error) {
	return s.attendance.AttendanceResults(weekNo, courseID, status, page, pageSize)
}

func (s *AttendanceService) AttendanceSessionSummaries(input query.AttendanceSessionSummaryListInput) ([]query.AttendanceSessionSummaryItem, int64, error) {
	input.Term = strings.TrimSpace(input.Term)
	input.Keyword = strings.TrimSpace(input.Keyword)
	input.LessonDate = strings.TrimSpace(input.LessonDate)
	input.LessonDateFrom = strings.TrimSpace(input.LessonDateFrom)
	input.LessonDateTo = strings.TrimSpace(input.LessonDateTo)
	input.CourseName = strings.TrimSpace(input.CourseName)
	input.TeacherName = strings.TrimSpace(input.TeacherName)
	input.WeekNo = strings.TrimSpace(input.WeekNo)
	input.Weekday = strings.TrimSpace(input.Weekday)
	input.Section = strings.TrimSpace(input.Section)
	input.ClassID = strings.TrimSpace(input.ClassID)
	input.ClassName = strings.TrimSpace(input.ClassName)
	input.Status = strings.TrimSpace(input.Status)
	return s.attendance.AttendanceSessionSummaries(input)
}

func (s *AttendanceService) resolveActiveTerm(now time.Time) (model.Term, error) {
	var term model.Term
	today := now.Format("2006-01-02")
	err := s.db.
		Where("term_start_date <= ?", today).
		Order("term_start_date DESC, id DESC").
		First(&term).Error
	switch {
	case err == nil:
		return term, nil
	case !errors.Is(err, gorm.ErrRecordNotFound):
		return model.Term{}, err
	}

	err = s.db.
		Order("term_start_date ASC, id ASC").
		First(&term).Error
	if err != nil {
		return model.Term{}, err
	}
	return term, nil
}

func (s *AttendanceService) AvailableCourseGroupLessons(termID uint64, weekday, weekNo int) ([]query.SessionWithCourse, error) {
	return s.attendance.AvailableCourseGroupLessons(termID, weekday, weekNo)
}

func (s *AttendanceService) AvailableCourseGroupLessonsForClass(termID uint64, weekday, weekNo int, classID uint64) ([]query.SessionWithCourse, error) {
	return s.attendance.AvailableCourseGroupLessonsForClass(termID, weekday, weekNo, classID)
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
	return courseGroupLessonID, nil
}

func (s *AttendanceService) UpdateAttendanceStatus(detailID uint64, status int, operatorUserID uint64, allowOverwrite bool) error {
	var record model.AttendanceRecord
	if err := s.db.First(&record, detailID).Error; err != nil {
		return ErrAttendanceRecordNotFound
	}
	if !allowOverwrite {
		return ErrAttendanceRecordLocked
	}
	if record.AttendanceStatus == status {
		return nil
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
		return s.audit.logAttendanceStatusChange(tx, record, operatorUserID, &oldStatus, status, now)
	})
}

func (s *AttendanceService) UpsertAttendanceStatusForStudent(sessionID, studentID uint64, status int, operatorUserID uint64, allowOverwrite bool) error {
	now := time.Now()
	return s.db.Transaction(func(tx *gorm.DB) error {
		lesson, group, relation, err := s.loadAttendanceStudentContext(tx, sessionID, studentID)
		if err != nil {
			return err
		}

		var record model.AttendanceRecord
		err = tx.Where("course_group_lesson_id = ? AND student_id = ?", lesson.ID, studentID).First(&record).Error
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			record = model.AttendanceRecord{
				TermID:              group.TermID,
				CourseID:            group.CourseID,
				CourseGroupLessonID: lesson.ID,
				StudentID:           studentID,
				ClassID:             relation.ClassID,
				AttendanceStatus:    status,
				UpdatedByUserID:     &operatorUserID,
				CreatedAt:           now,
				UpdatedAt:           now,
			}
			if err := tx.Create(&record).Error; err != nil {
				return err
			}
			return s.audit.logAttendanceStatusCreate(tx, record, operatorUserID, status, now)
		}
		if !allowOverwrite {
			return ErrAttendanceRecordLocked
		}
		if record.AttendanceStatus == status {
			return nil
		}
		oldStatus := record.AttendanceStatus
		if err := tx.Model(&record).Updates(map[string]interface{}{
			"attendance_status":  status,
			"updated_by_user_id": operatorUserID,
			"updated_at":         now,
		}).Error; err != nil {
			return err
		}
		return s.audit.logAttendanceStatusChange(tx, record, operatorUserID, &oldStatus, status, now)
	})
}

func (s *AttendanceService) BulkUpsertAttendanceStatusesForStudents(sessionID uint64, studentIDs []uint64, status int, operatorUserID uint64) (AdminBulkUpdateAttendanceStatusesResult, error) {
	uniqueStudentIDs := uniqueUint64s(studentIDs)
	if len(uniqueStudentIDs) == 0 {
		return AdminBulkUpdateAttendanceStatusesResult{}, nil
	}

	var result AdminBulkUpdateAttendanceStatusesResult
	now := time.Now()
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var lesson model.CourseGroupLesson
		if err := tx.First(&lesson, sessionID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrCourseGroupLessonNotFound
			}
			return err
		}

		var group model.CourseGroup
		if err := tx.First(&group, lesson.CourseGroupID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrCourseGroupNotFound
			}
			return err
		}

		var relations []model.CourseGroupStudent
		if err := tx.
			Where("course_group_id = ? AND student_id IN ? AND status = 1", lesson.CourseGroupID, uniqueStudentIDs).
			Find(&relations).Error; err != nil {
			return err
		}
		relationByStudentID := make(map[uint64]model.CourseGroupStudent, len(relations))
		for _, relation := range relations {
			relationByStudentID[relation.StudentID] = relation
		}

		var records []model.AttendanceRecord
		if err := tx.
			Where("course_group_lesson_id = ? AND student_id IN ?", lesson.ID, uniqueStudentIDs).
			Find(&records).Error; err != nil {
			return err
		}
		recordByStudentID := make(map[uint64]model.AttendanceRecord, len(records))
		for _, record := range records {
			recordByStudentID[record.StudentID] = record
		}

		for _, studentID := range uniqueStudentIDs {
			relation, ok := relationByStudentID[studentID]
			if !ok {
				result.FailedItems = append(result.FailedItems, studentID)
				result.FailedCount++
				continue
			}

			record, hasRecord := recordByStudentID[studentID]
			if !hasRecord {
				record = model.AttendanceRecord{
					TermID:              group.TermID,
					CourseID:            group.CourseID,
					CourseGroupLessonID: lesson.ID,
					StudentID:           studentID,
					ClassID:             relation.ClassID,
					AttendanceStatus:    status,
					UpdatedByUserID:     &operatorUserID,
					CreatedAt:           now,
					UpdatedAt:           now,
				}
				if err := tx.Create(&record).Error; err != nil {
					return err
				}
				if err := s.audit.logAttendanceStatusCreate(tx, record, operatorUserID, status, now); err != nil {
					return err
				}
				result.AppliedItems = append(result.AppliedItems, studentID)
				result.AppliedCount++
				continue
			}

			if record.AttendanceStatus != status {
				oldStatus := record.AttendanceStatus
				if err := tx.Model(&record).Updates(map[string]interface{}{
					"attendance_status":  status,
					"updated_by_user_id": operatorUserID,
					"updated_at":         now,
				}).Error; err != nil {
					return err
				}
				record.AttendanceStatus = status
				record.UpdatedByUserID = &operatorUserID
				record.UpdatedAt = now
				if err := s.audit.logAttendanceStatusChange(tx, record, operatorUserID, &oldStatus, status, now); err != nil {
					return err
				}
			}

			result.AppliedItems = append(result.AppliedItems, studentID)
			result.AppliedCount++
		}

		return nil
	})
	if err != nil {
		return AdminBulkUpdateAttendanceStatusesResult{}, err
	}
	return result, nil
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
			_, group, relation, err := s.loadAttendanceStudentContext(tx, lesson.ID, item.StudentRefID)
			if err != nil {
				if IsServiceError(err, ErrAttendanceRecordNotFound) || IsServiceError(err, ErrCourseGroupLessonNotFound) {
					result.IgnoredItems = append(result.IgnoredItems, item.StudentRefID)
					result.IgnoredCount++
					continue
				}
				return err
			}
			if classID != nil && (relation.ClassID == nil || *relation.ClassID != *classID) {
				result.IgnoredItems = append(result.IgnoredItems, item.StudentRefID)
				result.IgnoredCount++
				continue
			}

			var record model.AttendanceRecord
			err = tx.Where("course_group_lesson_id = ? AND student_id = ?", lesson.ID, item.StudentRefID).First(&record).Error
			if err == nil {
				result.IgnoredItems = append(result.IgnoredItems, item.StudentRefID)
				result.IgnoredCount++
				continue
			}
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}

			record = model.AttendanceRecord{
				TermID:              group.TermID,
				CourseID:            group.CourseID,
				CourseGroupLessonID: lesson.ID,
				StudentID:           item.StudentRefID,
				ClassID:             relation.ClassID,
				AttendanceStatus:    item.Status,
				UpdatedByUserID:     &operatorUserID,
				CreatedAt:           now,
				UpdatedAt:           now,
			}
			if err := tx.Create(&record).Error; err != nil {
				return err
			}
			if err := s.audit.logAttendanceStatusCreate(tx, record, operatorUserID, item.Status, now); err != nil {
				return err
			}
			result.AcceptedItems = append(result.AcceptedItems, item.StudentRefID)
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

func (s *AttendanceService) GetAttendanceSessionPage(sessionID uint64, studentID, realName, className, status, operatorName, operatedDate string, page, pageSize int) (model.CourseGroupLesson, model.Course, []query.AttendanceRecordItem, int64, error) {
	var lesson model.CourseGroupLesson
	if err := s.db.First(&lesson, sessionID).Error; err != nil {
		return model.CourseGroupLesson{}, model.Course{}, nil, 0, ErrCourseGroupLessonNotFound
	}

	var group model.CourseGroup
	if err := s.db.First(&group, lesson.CourseGroupID).Error; err != nil {
		return model.CourseGroupLesson{}, model.Course{}, nil, 0, ErrCourseGroupNotFound
	}

	var course model.Course
	if err := s.db.First(&course, group.CourseID).Error; err != nil {
		return model.CourseGroupLesson{}, model.Course{}, nil, 0, ErrCourseNotFound
	}

	records, total, err := s.attendance.AttendanceSessionRecordPage(lesson.ID, studentID, realName, className, status, operatorName, operatedDate, page, pageSize)
	if err != nil {
		return model.CourseGroupLesson{}, model.Course{}, nil, 0, err
	}

	return lesson, course, records, total, nil
}

func (s *AttendanceService) LocateAttendanceSessionRecordPage(sessionID uint64, studentID, realName, className, status, operatorName, operatedDate string, focusStudentID uint64, pageSize int) (query.FocusPageResult, error) {
	return s.attendance.LocateAttendanceSessionRecordPage(sessionID, studentID, realName, className, status, operatorName, operatedDate, focusStudentID, pageSize)
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

func (s *AttendanceService) loadAttendanceStudentContext(tx *gorm.DB, sessionID, studentID uint64) (model.CourseGroupLesson, model.CourseGroup, model.CourseGroupStudent, error) {
	var lesson model.CourseGroupLesson
	if err := tx.First(&lesson, sessionID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.CourseGroupLesson{}, model.CourseGroup{}, model.CourseGroupStudent{}, ErrCourseGroupLessonNotFound
		}
		return model.CourseGroupLesson{}, model.CourseGroup{}, model.CourseGroupStudent{}, err
	}
	var group model.CourseGroup
	if err := tx.First(&group, lesson.CourseGroupID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.CourseGroupLesson{}, model.CourseGroup{}, model.CourseGroupStudent{}, ErrCourseGroupNotFound
		}
		return model.CourseGroupLesson{}, model.CourseGroup{}, model.CourseGroupStudent{}, err
	}
	var relation model.CourseGroupStudent
	if err := tx.Where("course_group_id = ? AND student_id = ? AND status = 1", lesson.CourseGroupID, studentID).First(&relation).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.CourseGroupLesson{}, model.CourseGroup{}, model.CourseGroupStudent{}, ErrAttendanceRecordNotFound
		}
		return model.CourseGroupLesson{}, model.CourseGroup{}, model.CourseGroupStudent{}, err
	}
	return lesson, group, relation, nil
}

func (s *AttendanceService) AbandonAttendanceSession(checkID uint64, withinDeadline func(model.CourseGroupLesson, time.Time) bool) error {
	var lesson model.CourseGroupLesson
	if err := s.db.First(&lesson, checkID).Error; err != nil {
		return ErrCourseGroupLessonNotFound
	}
	if !withinDeadline(lesson, time.Now()) {
		return ErrAttendanceDeadlinePassed
	}
	return nil
}
