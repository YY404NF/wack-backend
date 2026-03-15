package service

import (
	"fmt"
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

func (s *CourseService) ListCourses(term, teacher, keyword string, page, pageSize int) ([]model.Course, int64, error) {
	queryDB := s.db.Model(&model.Course{})
	if term != "" {
		queryDB = queryDB.Where("term = ?", term)
	}
	if teacher != "" {
		queryDB = queryDB.Where("teacher_name LIKE ?", "%"+teacher+"%")
	}
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		queryDB = queryDB.Where("course_name LIKE ?", "%"+keyword+"%")
	}
	var total int64
	if err := queryDB.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []model.Course
	if err := queryDB.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
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

func (s *CourseService) ReplaceCourseStudents(id uint64, studentIDs []string, users []model.User) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("course_id = ?", id).Delete(&model.CourseStudent{}).Error; err != nil {
			return err
		}
		userIDByStudentID := make(map[string]uint64, len(users))
		for _, user := range users {
			userIDByStudentID[user.StudentID] = user.ID
		}
		relations := make([]model.CourseStudent, 0, len(studentIDs))
		for _, studentID := range studentIDs {
			userID, ok := userIDByStudentID[studentID]
			if !ok {
				return fmt.Errorf("student %s not found", studentID)
			}
			relations = append(relations, model.CourseStudent{CourseID: id, UserID: userID})
		}
		if len(relations) > 0 {
			if err := tx.Create(&relations).Error; err != nil {
				return err
			}
		}
		return tx.Model(&model.Course{}).Where("id = ?", id).Update("attendance_student_count", len(studentIDs)).Error
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
