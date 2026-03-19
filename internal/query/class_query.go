package query

import (
	"time"

	"gorm.io/gorm"

	"wack-backend/internal/model"
)

type ClassStudentCandidateItem struct {
	ID        uint64 `json:"id"`
	ClassID   uint64 `json:"class_id"`
	StudentID string `json:"student_id"`
	RealName  string `json:"real_name"`
	ClassName string `json:"class_name"`
	Grade     int    `json:"grade"`
	MajorName string `json:"major_name"`
}

type ClassQuery struct {
	db *gorm.DB
}

func NewClassQuery(db *gorm.DB) *ClassQuery {
	return &ClassQuery{db: db}
}

func (q *ClassQuery) ListClasses(page, pageSize int) ([]model.Class, int64, error) {
	base := q.db.Table("class").
		Select("class.id, class.class_name, class.grade, class.major_name, class.status, class.created_at, class.updated_at, COUNT(student.id) AS student_count").
		Joins("LEFT JOIN student ON student.class_id = class.id AND student.status = 1").
		Group("class.id")

	var total int64
	if err := q.db.Table("class").Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []model.Class
	err := base.
		Order("class.grade DESC, class.major_name ASC, class.class_name ASC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&items).Error
	return items, total, err
}

type ClassStudentItem struct {
	ID        uint64    `json:"id"`
	ClassID   uint64    `json:"class_id"`
	StudentID string    `json:"student_id"`
	RealName  string    `json:"real_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (q *ClassQuery) ClassStudents(classID uint64) ([]ClassStudentItem, error) {
	var items []ClassStudentItem
	err := q.db.Table("student").
		Select("student.id, student.class_id, student.student_no AS student_id, student.student_name AS real_name, student.created_at, student.updated_at").
		Where("student.class_id = ? AND student.status = 1", classID).
		Order("student.student_no ASC").
		Scan(&items).Error
	return items, err
}

func (q *ClassQuery) StudentCandidates() ([]ClassStudentCandidateItem, error) {
	var items []ClassStudentCandidateItem
	err := q.db.Table("student").
		Select("student.id, student.class_id, student.student_no AS student_id, student.student_name AS real_name, class.class_name, class.grade, class.major_name").
		Joins("JOIN class ON class.id = student.class_id").
		Where("student.status = 1").
		Order("student.student_no ASC").
		Scan(&items).Error
	return items, err
}
