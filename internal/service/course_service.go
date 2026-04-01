package service

import (
	"strings"

	"gorm.io/gorm"

	"wack-backend/internal/model"
	"wack-backend/internal/query"
)

type CourseService struct {
	db      *gorm.DB
	courses *query.CourseQuery
}

func NewCourseService(db *gorm.DB) *CourseService {
	return &CourseService{db: db, courses: query.NewCourseQuery(db)}
}

func (s *CourseService) ListCourses(term, grade, teacher, keyword, className, studentCount string, page, pageSize int) ([]query.CourseListItem, int64, error) {
	return s.courses.ListCourses(
		strings.TrimSpace(term),
		strings.TrimSpace(grade),
		strings.TrimSpace(teacher),
		strings.TrimSpace(keyword),
		strings.TrimSpace(className),
		strings.TrimSpace(studentCount),
		page,
		pageSize,
	)
}

func (s *CourseService) LocateCoursePage(term, grade, teacher, keyword, className, studentCount string, focusCourseID uint64, pageSize int) (query.FocusPageResult, error) {
	return s.courses.LocateCoursePage(
		strings.TrimSpace(term),
		strings.TrimSpace(grade),
		strings.TrimSpace(teacher),
		strings.TrimSpace(keyword),
		strings.TrimSpace(className),
		strings.TrimSpace(studentCount),
		focusCourseID,
		pageSize,
	)
}

func (s *CourseService) CreateCourse(course model.Course) (model.Course, error) {
	course.CourseName = strings.TrimSpace(course.CourseName)
	course.TeacherName = strings.TrimSpace(course.TeacherName)
	course.Status = 1
	if err := validateCourse(course); err != nil {
		return model.Course{}, err
	}
	if err := ensureTermExists(s.db, course.TermID); err != nil {
		return model.Course{}, err
	}
	if err := s.db.Create(&course).Error; err != nil {
		return model.Course{}, err
	}
	return s.loadCourseView(course.ID)
}

func (s *CourseService) ListCourseGroups(courseID uint64) ([]query.CourseGroupListItem, error) {
	if err := ensureCourseExists(s.db, courseID); err != nil {
		return nil, err
	}
	return s.courses.CourseGroups(courseID)
}

func (s *CourseService) GetCourseSummary(courseID uint64) (model.Course, error) {
	return s.loadCourseView(courseID)
}

func (s *CourseService) GetCourseGroup(courseID, groupID uint64) (model.CourseGroup, []query.CourseGroupStudentItem, []model.CourseGroupLesson, error) {
	group, err := s.courses.CourseGroup(courseID, groupID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return model.CourseGroup{}, nil, nil, ErrCourseGroupNotFound
		}
		return model.CourseGroup{}, nil, nil, err
	}
	students, err := s.courses.CourseGroupStudents(groupID)
	if err != nil {
		return model.CourseGroup{}, nil, nil, err
	}
	lessons, err := s.courses.CourseGroupLessons(groupID)
	if err != nil {
		return model.CourseGroup{}, nil, nil, err
	}
	return group, students, lessons, nil
}

func (s *CourseService) GetCourseGroupLessons(courseID, groupID uint64) ([]model.CourseGroupLesson, error) {
	if _, err := s.courses.CourseGroup(courseID, groupID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrCourseGroupNotFound
		}
		return nil, err
	}
	return s.courses.CourseGroupLessons(groupID)
}

func (s *CourseService) CreateCourseGroupLesson(courseID, groupID uint64, lesson model.CourseGroupLesson) (model.CourseGroupLesson, error) {
	group, err := ensureCourseGroupExists(s.db, courseID, groupID)
	if err != nil {
		return model.CourseGroupLesson{}, err
	}
	lesson.ID = 0
	lesson.TermID = group.TermID
	lesson.CourseGroupID = group.ID
	lesson.Status = 1
	lesson.BuildingName = strings.TrimSpace(lesson.BuildingName)
	lesson.RoomName = strings.TrimSpace(lesson.RoomName)
	if err := validateCourseGroupLesson(lesson); err != nil {
		return model.CourseGroupLesson{}, err
	}
	return lesson, s.db.Create(&lesson).Error
}

func (s *CourseService) UpdateCourseGroupLesson(courseID, groupID, lessonID uint64, input model.CourseGroupLesson) (model.CourseGroupLesson, error) {
	returnLesson := model.CourseGroupLesson{}
	err := s.db.Transaction(func(tx *gorm.DB) error {
		group, err := ensureCourseGroupExists(tx, courseID, groupID)
		if err != nil {
			return err
		}
		var lesson model.CourseGroupLesson
		if err := tx.Where("id = ? AND course_group_id = ? AND status = 1", lessonID, group.ID).First(&lesson).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return ErrCourseGroupLessonNotFound
			}
			return err
		}
		lesson.WeekNo = input.WeekNo
		lesson.Weekday = input.Weekday
		lesson.Section = input.Section
		lesson.BuildingName = strings.TrimSpace(input.BuildingName)
		lesson.RoomName = strings.TrimSpace(input.RoomName)
		lesson.TermID = group.TermID
		lesson.CourseGroupID = group.ID
		if err := validateCourseGroupLesson(lesson); err != nil {
			return err
		}
		if err := tx.Model(&lesson).Updates(map[string]interface{}{
			"week_no":       lesson.WeekNo,
			"weekday":       lesson.Weekday,
			"section":       lesson.Section,
			"building_name": lesson.BuildingName,
			"room_name":     lesson.RoomName,
		}).Error; err != nil {
			return err
		}
		returnLesson = lesson
		return nil
	})
	return returnLesson, err
}

func (s *CourseService) DeleteCourseGroupLesson(courseID, groupID, lessonID uint64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		group, err := ensureCourseGroupExists(tx, courseID, groupID)
		if err != nil {
			return err
		}
		result := tx.Where("id = ? AND course_group_id = ?", lessonID, group.ID).Delete(&model.CourseGroupLesson{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrCourseGroupLessonNotFound
		}
		return nil
	})
}

func (s *CourseService) GetCourseGroupStudents(courseID, groupID uint64) ([]query.CourseGroupStudentItem, error) {
	if _, err := s.courses.CourseGroup(courseID, groupID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrCourseGroupNotFound
		}
		return nil, err
	}
	return s.courses.CourseGroupStudents(groupID)
}

func (s *CourseService) ListAvailableCourseGroupClasses(courseID, groupID uint64, className string, page, pageSize int) ([]query.AvailableCourseGroupClassItem, int64, error) {
	if _, err := s.courses.CourseGroup(courseID, groupID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, 0, ErrCourseGroupNotFound
		}
		return nil, 0, err
	}
	return s.courses.AvailableCourseGroupClasses(groupID, strings.TrimSpace(className), page, pageSize)
}

func (s *CourseService) ListAvailableCourseGroupStudents(courseID, groupID uint64, studentNo, studentName, className string, page, pageSize int) ([]query.AvailableCourseGroupStudentItem, int64, error) {
	if _, err := s.courses.CourseGroup(courseID, groupID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, 0, ErrCourseGroupNotFound
		}
		return nil, 0, err
	}
	return s.courses.AvailableCourseGroupStudents(
		groupID,
		strings.TrimSpace(studentNo),
		strings.TrimSpace(studentName),
		strings.TrimSpace(className),
		page,
		pageSize,
	)
}

func (s *CourseService) CreateCourseGroup(courseID uint64) (model.CourseGroup, error) {
	termID, err := courseTermID(s.db, courseID)
	if err != nil {
		return model.CourseGroup{}, err
	}
	group := model.CourseGroup{
		TermID:   termID,
		CourseID: courseID,
		Status:   1,
	}
	return group, s.db.Create(&group).Error
}

func (s *CourseService) DeleteCourseGroup(courseID, groupID uint64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		group, err := ensureCourseGroupExists(tx, courseID, groupID)
		if err != nil {
			return err
		}
		var lessonIDs []uint64
		if err := tx.Model(&model.CourseGroupLesson{}).
			Where("course_group_id = ? AND status = 1", groupID).
			Pluck("id", &lessonIDs).Error; err != nil {
			return err
		}
		hasHistory := false
		if len(lessonIDs) > 0 {
			var recordCount int64
			if err := tx.Model(&model.AttendanceRecord{}).
				Where("course_group_lesson_id IN ?", lessonIDs).
				Count(&recordCount).Error; err != nil {
				return err
			}
			hasHistory = recordCount > 0
		}
		if hasHistory {
			if err := tx.Model(&model.CourseGroupLesson{}).
				Where("course_group_id = ? AND status = 1", groupID).
				Update("status", 0).Error; err != nil {
				return err
			}
			if err := tx.Model(&model.CourseGroupStudent{}).
				Where("course_group_id = ? AND status = 1", groupID).
				Update("status", 0).Error; err != nil {
				return err
			}
			return tx.Model(&model.CourseGroup{}).
				Where("id = ? AND course_id = ? AND status = 1", group.ID, courseID).
				Update("status", 0).Error
		}
		if err := tx.Where("course_group_id = ?", groupID).Delete(&model.CourseGroupLesson{}).Error; err != nil {
			return err
		}
		if err := tx.Where("course_group_id = ?", groupID).Delete(&model.CourseGroupStudent{}).Error; err != nil {
			return err
		}
		return tx.Where("id = ? AND course_id = ?", groupID, courseID).Delete(&model.CourseGroup{}).Error
	})
}

func (s *CourseService) AddCourseGroupClasses(courseID, groupID uint64, classIDs []uint64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		group, err := ensureCourseGroupExists(tx, courseID, groupID)
		if err != nil {
			return err
		}
		classIDs = normalizeUint64s(classIDs)
		if len(classIDs) == 0 {
			return ErrInvalidInput
		}
		if err := ensureClassesExist(tx, classIDs); err != nil {
			return err
		}
		var students []model.Student
		if err := tx.Where("class_id IN ? AND status = 1", classIDs).Find(&students).Error; err != nil {
			return err
		}
		if len(students) == 0 {
			return nil
		}
		for _, student := range students {
			if student.ClassID == nil {
				continue
			}
			classID := new(uint64)
			*classID = *student.ClassID
			if err := upsertCourseGroupStudent(tx, model.CourseGroupStudent{
				TermID:        group.TermID,
				CourseGroupID: group.ID,
				StudentID:     student.ID,
				ClassID:       classID,
				Status:        1,
			}); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *CourseService) RemoveCourseGroupClass(courseID, groupID, classID uint64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if _, err := ensureCourseGroupExists(tx, courseID, groupID); err != nil {
			return err
		}
		var relationCount int64
		if err := tx.Model(&model.CourseGroupStudent{}).
			Where("course_group_id = ? AND class_id = ? AND status = 1", groupID, classID).
			Count(&relationCount).Error; err != nil {
			return err
		}
		if relationCount == 0 {
			return ErrClassNotFound
		}
		hasHistory, err := courseGroupMemberHasHistory(tx, groupID, func(db *gorm.DB) *gorm.DB {
			return db.Where("class_id = ?", classID)
		})
		if err != nil {
			return err
		}
		if hasHistory {
			return tx.Model(&model.CourseGroupStudent{}).
				Where("course_group_id = ? AND class_id = ? AND status = 1", groupID, classID).
				Update("status", 0).Error
		}
		return tx.Where("course_group_id = ? AND class_id = ? AND status = 1", groupID, classID).Delete(&model.CourseGroupStudent{}).Error
	})
}

func (s *CourseService) AddCourseGroupStudents(courseID, groupID uint64, studentIDs []uint64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		group, err := ensureCourseGroupExists(tx, courseID, groupID)
		if err != nil {
			return err
		}
		studentIDs = normalizeUint64s(studentIDs)
		if len(studentIDs) == 0 {
			return ErrInvalidInput
		}
		var students []model.Student
		if err := tx.Where("id IN ? AND status = 1", studentIDs).Find(&students).Error; err != nil {
			return err
		}
		if len(students) != len(studentIDs) {
			return ErrStudentNotFound
		}
		for _, student := range students {
			classID := (*uint64)(nil)
			if student.ClassID != nil {
				var classRelationCount int64
				if err := tx.Model(&model.CourseGroupStudent{}).
					Where("course_group_id = ? AND class_id = ? AND status = 1", groupID, *student.ClassID).
					Count(&classRelationCount).Error; err != nil {
					return err
				}
				if classRelationCount > 0 {
					classID = new(uint64)
					*classID = *student.ClassID
				}
			}
			if err := upsertCourseGroupStudent(tx, model.CourseGroupStudent{
				TermID:        group.TermID,
				CourseGroupID: group.ID,
				StudentID:     student.ID,
				ClassID:       classID,
				Status:        1,
			}); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *CourseService) RemoveCourseGroupStudent(courseID, groupID, studentID uint64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if _, err := ensureCourseGroupExists(tx, courseID, groupID); err != nil {
			return err
		}
		var relationCount int64
		if err := tx.Model(&model.CourseGroupStudent{}).
			Where("course_group_id = ? AND student_id = ? AND status = 1", groupID, studentID).
			Count(&relationCount).Error; err != nil {
			return err
		}
		if relationCount == 0 {
			return ErrStudentNotFound
		}
		hasHistory, err := courseGroupMemberHasHistory(tx, groupID, func(db *gorm.DB) *gorm.DB {
			return db.Where("student_id = ?", studentID)
		})
		if err != nil {
			return err
		}
		if hasHistory {
			return tx.Model(&model.CourseGroupStudent{}).
				Where("course_group_id = ? AND student_id = ? AND status = 1", groupID, studentID).
				Update("status", 0).Error
		}
		return tx.Where("course_group_id = ? AND student_id = ? AND status = 1", groupID, studentID).Delete(&model.CourseGroupStudent{}).Error
	})
}

func (s *CourseService) UpdateCourse(id uint64, req model.Course) (model.Course, error) {
	req.CourseName = strings.TrimSpace(req.CourseName)
	req.TeacherName = strings.TrimSpace(req.TeacherName)
	req.Status = 1
	if err := validateCourse(req); err != nil {
		return model.Course{}, err
	}
	if err := ensureTermExists(s.db, req.TermID); err != nil {
		return model.Course{}, err
	}
	var course model.Course
	if err := s.db.First(&course, id).Error; err != nil {
		return model.Course{}, ErrCourseNotFound
	}
	if err := s.db.Model(&course).Updates(map[string]interface{}{
		"term_id":      req.TermID,
		"grade":        req.Grade,
		"course_name":  req.CourseName,
		"teacher_name": req.TeacherName,
		"status":       req.Status,
	}).Error; err != nil {
		return model.Course{}, err
	}
	return s.loadCourseView(id)
}

func (s *CourseService) DeleteCourse(id uint64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var groupIDs []uint64
		if err := tx.Model(&model.CourseGroup{}).Where("course_id = ?", id).Pluck("id", &groupIDs).Error; err != nil {
			return err
		}
		if len(groupIDs) > 0 {
			var lessonIDs []uint64
			if err := tx.Model(&model.CourseGroupLesson{}).Where("course_group_id IN ?", groupIDs).Pluck("id", &lessonIDs).Error; err != nil {
				return err
			}
			if len(lessonIDs) > 0 {
				var recordIDs []uint64
				if err := tx.Model(&model.AttendanceRecord{}).Where("course_group_lesson_id IN ?", lessonIDs).Pluck("id", &recordIDs).Error; err != nil {
					return err
				}
				if len(recordIDs) > 0 {
					if err := tx.Where("attendance_record_id IN ?", recordIDs).Delete(&model.AttendanceRecordLog{}).Error; err != nil {
						return err
					}
				}
				if len(lessonIDs) > 0 {
					if err := tx.Where("course_group_lesson_id IN ?", lessonIDs).Delete(&model.AttendanceRecord{}).Error; err != nil {
						return err
					}
				}
				if err := tx.Where("id IN ?", lessonIDs).Delete(&model.CourseGroupLesson{}).Error; err != nil {
					return err
				}
			}
			if err := tx.Where("course_group_id IN ?", groupIDs).Delete(&model.CourseGroupStudent{}).Error; err != nil {
				return err
			}
			if err := tx.Where("id IN ?", groupIDs).Delete(&model.CourseGroup{}).Error; err != nil {
				return err
			}
		}
		return tx.Delete(&model.Course{}, id).Error
	})
}

func (s *CourseService) CourseCalendar(weekNo, term string) ([]query.CourseCalendarItem, error) {
	return s.courses.CourseCalendar(weekNo, term)
}

func ensureCourseExists(db *gorm.DB, courseID uint64) error {
	var count int64
	if err := db.Model(&model.Course{}).Where("id = ?", courseID).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return ErrCourseNotFound
	}
	return nil
}

func ensureCourseGroupExists(db *gorm.DB, courseID, groupID uint64) (model.CourseGroup, error) {
	var group model.CourseGroup
	if err := db.Where("id = ? AND course_id = ? AND status = 1", groupID, courseID).First(&group).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return model.CourseGroup{}, ErrCourseGroupNotFound
		}
		return model.CourseGroup{}, err
	}
	return group, nil
}

func courseTermID(db *gorm.DB, courseID uint64) (uint64, error) {
	type courseTermRow struct {
		TermID uint64 `gorm:"column:term_id"`
	}
	var row courseTermRow
	if err := db.Table("course").Select("term_id").Where("id = ?", courseID).First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, ErrCourseNotFound
		}
		return 0, err
	}
	if row.TermID == 0 {
		return 0, ErrCourseNotFound
	}
	return row.TermID, nil
}

func ensureTermExists(db *gorm.DB, termID uint64) error {
	if termID == 0 {
		return ErrInvalidInput
	}
	var count int64
	if err := db.Model(&model.Term{}).Where("id = ?", termID).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return ErrTermNotFound
	}
	return nil
}

func (s *CourseService) loadCourseView(courseID uint64) (model.Course, error) {
	var course model.Course
	err := s.db.Table("course").
		Select("course.id, course.term_id, course.grade, term.name AS term, course.course_name, course.teacher_name, course.status, 0 AS student_count, course.created_at, course.updated_at").
		Joins("JOIN term ON term.id = course.term_id").
		Where("course.id = ?", courseID).
		Scan(&course).Error
	if err != nil {
		return model.Course{}, err
	}
	if course.ID == 0 {
		return model.Course{}, ErrCourseNotFound
	}
	return course, nil
}

func normalizeUint64s(values []uint64) []uint64 {
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

func upsertCourseGroupStudent(tx *gorm.DB, relation model.CourseGroupStudent) error {
	var existing model.CourseGroupStudent
	err := tx.Where(
		"term_id = ? AND course_group_id = ? AND student_id = ?",
		relation.TermID,
		relation.CourseGroupID,
		relation.StudentID,
	).First(&existing).Error
	if err == nil {
		return tx.Model(&existing).Updates(map[string]interface{}{
			"class_id": relation.ClassID,
			"status":   relation.Status,
		}).Error
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	return tx.Create(&relation).Error
}

func courseGroupMemberHasHistory(tx *gorm.DB, groupID uint64, scope func(*gorm.DB) *gorm.DB) (bool, error) {
	lessonQuery := tx.Model(&model.CourseGroupLesson{}).
		Select("id").
		Where("course_group_id = ? AND status = 1", groupID)
	recordQuery := tx.Model(&model.AttendanceRecord{}).
		Where("course_group_lesson_id IN (?)", lessonQuery)
	if scope != nil {
		recordQuery = scope(recordQuery)
	}
	var recordCount int64
	if err := recordQuery.Count(&recordCount).Error; err != nil {
		return false, err
	}
	return recordCount > 0, nil
}

func ensureClassesExist(db *gorm.DB, classIDs []uint64) error {
	if len(classIDs) == 0 {
		return nil
	}
	var count int64
	if err := db.Model(&model.Class{}).Where("id IN ?", classIDs).Count(&count).Error; err != nil {
		return err
	}
	if count != int64(len(classIDs)) {
		return ErrClassNotFound
	}
	return nil
}

func validateCourse(course model.Course) error {
	if course.TermID == 0 || course.Grade <= 0 || course.CourseName == "" || course.TeacherName == "" {
		return ErrInvalidInput
	}
	if len(course.CourseName) > 100 || len(course.TeacherName) > 50 {
		return ErrInvalidInput
	}
	return nil
}

func validateCourseGroupLesson(lesson model.CourseGroupLesson) error {
	if lesson.WeekNo < 1 || lesson.Weekday < 1 || lesson.Weekday > 7 || lesson.Section < 1 || lesson.Section > 5 {
		return ErrInvalidInput
	}
	if lesson.BuildingName == "" || lesson.RoomName == "" {
		return ErrInvalidInput
	}
	if len(lesson.BuildingName) > 50 || len(lesson.RoomName) > 50 {
		return ErrInvalidInput
	}
	return nil
}
