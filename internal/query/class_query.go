package query

import "gorm.io/gorm"

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

func (q *ClassQuery) StudentCandidates() ([]ClassStudentCandidateItem, error) {
	var items []ClassStudentCandidateItem
	err := q.db.Table("class_student").
		Select("class_student.id, class_student.class_id, class_student.student_id, class_student.real_name, class.class_name, class.grade, class.major_name").
		Joins("JOIN class ON class.id = class_student.class_id").
		Order("class_student.student_id ASC").
		Scan(&items).Error
	return items, err
}
