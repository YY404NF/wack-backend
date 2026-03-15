package query

import (
	"time"

	"gorm.io/gorm"

	"wack-backend/internal/model"
)

type CourseStudentItem struct {
	ID        uint64    `json:"id"`
	CourseID  uint64    `json:"course_id"`
	UserID    uint64    `json:"user_id"`
	StudentID string    `json:"student_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CourseCalendarItem struct {
	model.CourseSession
	CourseName  string `json:"course_name"`
	TeacherName string `json:"teacher_name"`
}

type CourseQuery struct {
	db *gorm.DB
}

func NewCourseQuery(db *gorm.DB) *CourseQuery {
	return &CourseQuery{db: db}
}

func (q *CourseQuery) CourseStudents(courseID uint64) ([]CourseStudentItem, error) {
	var students []CourseStudentItem
	err := q.db.Table("course_student").
		Select("course_student.id, course_student.course_id, course_student.user_id, user.student_id, course_student.created_at, course_student.updated_at").
		Joins("JOIN user ON user.id = course_student.user_id").
		Where("course_student.course_id = ?", courseID).
		Find(&students).Error
	return students, err
}

func (q *CourseQuery) CourseCalendar(weekNo, term string) ([]CourseCalendarItem, error) {
	query := q.db.Model(&model.CourseSession{}).
		Joins("JOIN course ON course.id = course_session.course_id")
	if weekNo != "" {
		query = query.Where("week_no = ?", weekNo)
	}
	if term != "" {
		query = query.Where("course.term = ?", term)
	}
	var items []CourseCalendarItem
	err := query.Select("course_session.*, course.course_name, course.teacher_name").
		Order("week_no, weekday, section, session_no").
		Scan(&items).Error
	return items, err
}
