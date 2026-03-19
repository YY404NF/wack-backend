package query

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

type StudentItem struct {
	ID        uint64    `json:"id"`
	ClassID   *uint64   `json:"class_id"`
	StudentID string    `json:"student_id"`
	RealName  string    `json:"real_name"`
	ClassName *string   `json:"class_name"`
	Grade     *int      `json:"grade"`
	MajorName *string   `json:"major_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ListStudentsInput struct {
	Page     int
	PageSize int
	ClassID  uint64
	Keyword  string
}

type StudentQuery struct {
	db *gorm.DB
}

func NewStudentQuery(db *gorm.DB) *StudentQuery {
	return &StudentQuery{db: db}
}

func (q *StudentQuery) ListStudents(input ListStudentsInput) ([]StudentItem, int64, error) {
	base := q.db.Table("student").
		Select(`
			student.id,
			student.class_id,
			student.student_no AS student_id,
			student.student_name AS real_name,
			class.class_name,
			class.grade,
			class.major_name,
			student.created_at,
			student.updated_at
		`).
		Joins("LEFT JOIN class ON class.id = student.class_id").
		Where("student.status = 1 AND (class.status = 1 OR student.class_id IS NULL)")

	if input.ClassID > 0 {
		base = base.Where("student.class_id = ?", input.ClassID)
	}

	keyword := strings.TrimSpace(input.Keyword)
	if keyword != "" {
		like := "%" + keyword + "%"
		base = base.Where(
			"student.student_no LIKE ? OR student.student_name LIKE ? OR COALESCE(class.class_name, '') LIKE ? OR COALESCE(class.major_name, '') LIKE ?",
			like,
			like,
			like,
			like,
		)
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []StudentItem
	err := base.
		Order("student.student_no ASC, student.id ASC").
		Offset((input.Page - 1) * input.PageSize).
		Limit(input.PageSize).
		Scan(&items).Error
	return items, total, err
}

func (q *StudentQuery) GetStudent(id uint64) (StudentItem, error) {
	var item StudentItem
	err := q.db.Table("student").
		Select(`
			student.id,
			student.class_id,
			student.student_no AS student_id,
			student.student_name AS real_name,
			class.class_name,
			class.grade,
			class.major_name,
			student.created_at,
			student.updated_at
		`).
		Joins("LEFT JOIN class ON class.id = student.class_id").
		Where("student.id = ? AND student.status = 1 AND (class.status = 1 OR student.class_id IS NULL)", id).
		Scan(&item).Error
	return item, err
}
