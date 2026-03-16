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

func (s *CourseService) ListCourses(term, teacher, keyword string, page, pageSize int) ([]query.CourseListItem, int64, error) {
	return s.courses.ListCourses(term, teacher, strings.TrimSpace(keyword), page, pageSize)
}

func (s *CourseService) CreateCourse(course model.Course) (model.Course, error) {
	return course, s.db.Create(&course).Error
}

func (s *CourseService) GetCourse(id uint64) (model.Course, []query.CourseStudentItem, []model.CourseClass, []model.CourseSession, error) {
	var course model.Course
	if err := s.db.First(&course, id).Error; err != nil {
		return model.Course{}, nil, nil, nil, ErrCourseNotFound
	}
	students, err := s.courses.CourseStudents(id)
	if err != nil {
		return model.Course{}, nil, nil, nil, err
	}
	var classes []model.CourseClass
	if err := s.db.Where("course_id = ?", id).Find(&classes).Error; err != nil {
		return model.Course{}, nil, nil, nil, err
	}
	var sessions []model.CourseSession
	if err := s.db.Where("course_id = ?", id).Order("session_no ASC").Find(&sessions).Error; err != nil {
		return model.Course{}, nil, nil, nil, err
	}
	return course, students, classes, sessions, nil
}

func (s *CourseService) UpdateCourse(id uint64, req model.Course) (model.Course, error) {
	var course model.Course
	if err := s.db.First(&course, id).Error; err != nil {
		return model.Course{}, ErrCourseNotFound
	}
	if err := s.db.Model(&course).Updates(map[string]interface{}{
		"term":                     req.Term,
		"course_name":              req.CourseName,
		"teacher_name":             req.TeacherName,
		"attendance_student_count": req.AttendanceStudentCount,
	}).Error; err != nil {
		return model.Course{}, err
	}
	if err := s.db.First(&course, id).Error; err != nil {
		return model.Course{}, err
	}
	return course, nil
}

func (s *CourseService) DeleteCourse(id uint64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var sessionIDs []uint64
		if err := tx.Model(&model.CourseSession{}).Where("course_id = ?", id).Pluck("id", &sessionIDs).Error; err != nil {
			return err
		}
		if len(sessionIDs) > 0 {
			var checkIDs []uint64
			if err := tx.Model(&model.AttendanceCheck{}).Where("course_session_id IN ?", sessionIDs).Pluck("id", &checkIDs).Error; err != nil {
				return err
			}
			if len(checkIDs) > 0 {
				if err := tx.Where("attendance_check_id IN ?", checkIDs).Delete(&model.AttendanceDetailLog{}).Error; err != nil {
					return err
				}
				if err := tx.Where("attendance_check_id IN ?", checkIDs).Delete(&model.AttendanceDetail{}).Error; err != nil {
					return err
				}
				if err := tx.Where("id IN ?", checkIDs).Delete(&model.AttendanceCheck{}).Error; err != nil {
					return err
				}
			}
			if err := tx.Where("id IN ?", sessionIDs).Delete(&model.CourseSession{}).Error; err != nil {
				return err
			}
		}
		if err := tx.Where("course_id = ?", id).Delete(&model.CourseStudent{}).Error; err != nil {
			return err
		}
		if err := tx.Where("course_id = ?", id).Delete(&model.CourseClass{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.Course{}, id).Error
	})
}

func (s *CourseService) ReplaceCourseStudents(id uint64, students []model.CourseStudent) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("course_id = ?", id).Delete(&model.CourseStudent{}).Error; err != nil {
			return err
		}
		relations := make([]model.CourseStudent, 0, len(students))
		for _, student := range students {
			relations = append(relations, model.CourseStudent{
				CourseID:  id,
				StudentID: student.StudentID,
				RealName:  student.RealName,
			})
		}
		if len(relations) > 0 {
			if err := tx.Create(&relations).Error; err != nil {
				return err
			}
		}
		return tx.Model(&model.Course{}).Where("id = ?", id).Update("attendance_student_count", len(students)).Error
	})
}

func (s *CourseService) ReplaceCourseClasses(id uint64, classIDs []uint64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("course_id = ?", id).Delete(&model.CourseClass{}).Error; err != nil {
			return err
		}
		relations := make([]model.CourseClass, 0, len(classIDs))
		for _, classID := range classIDs {
			relations = append(relations, model.CourseClass{CourseID: id, ClassID: classID})
		}
		if len(relations) > 0 {
			return tx.Create(&relations).Error
		}
		return nil
	})
}

func (s *CourseService) ReplaceCourseSessions(id uint64, sessions []model.CourseSession) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("course_id = ?", id).Delete(&model.CourseSession{}).Error; err != nil {
			return err
		}
		for i := range sessions {
			sessions[i].ID = 0
			sessions[i].CourseID = id
		}
		if len(sessions) > 0 {
			return tx.Create(&sessions).Error
		}
		return nil
	})
}

func (s *CourseService) CourseCalendar(weekNo, term string) ([]query.CourseCalendarItem, error) {
	return s.courses.CourseCalendar(weekNo, term)
}
