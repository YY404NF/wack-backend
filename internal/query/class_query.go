package query

import (
	"strings"
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

type ClassOptionItem struct {
	ID        uint64 `json:"id"`
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

func (q *ClassQuery) classListFilterBase(grade, majorName, className string) *gorm.DB {
	base := q.db.Table("class").Where("class.status = 1")

	if grade = strings.TrimSpace(grade); grade != "" {
		base = base.Where("CAST(class.grade AS TEXT) = ?", grade)
	}
	if majorName = strings.TrimSpace(majorName); majorName != "" {
		base = base.Where("class.major_name LIKE ?", "%"+majorName+"%")
	}
	if className = strings.TrimSpace(className); className != "" {
		base = base.Where("class.class_name LIKE ?", "%"+className+"%")
	}

	return base
}

func (q *ClassQuery) ListClasses(grade, majorName, className string, page, pageSize int) ([]model.Class, int64, error) {
	filterBase := q.classListFilterBase(grade, majorName, className)
	base := filterBase.
		Select("class.id, class.class_name, class.grade, class.major_name, class.status, class.created_at, class.updated_at, COUNT(student.id) AS student_count").
		Joins("LEFT JOIN student ON student.class_id = class.id AND student.status = 1").
		Group("class.id")

	var total int64
	if err := filterBase.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []model.Class
	err := base.
		Order("class.grade DESC, class.major_name ASC, class.class_name ASC, class.id ASC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&items).Error
	return items, total, err
}

func (q *ClassQuery) LocateClassPage(grade, majorName, className string, focusClassID uint64, pageSize int) (FocusPageResult, error) {
	base := q.classListFilterBase(grade, majorName, className)

	var target struct {
		ID        uint64 `gorm:"column:id"`
		Grade     int    `gorm:"column:grade"`
		MajorName string `gorm:"column:major_name"`
		ClassName string `gorm:"column:class_name"`
	}
	if err := base.Select("class.id, class.grade, class.major_name, class.class_name").
		Where("class.id = ?", focusClassID).
		Take(&target).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return FocusPageResult{}, nil
		}
		return FocusPageResult{}, err
	}

	var rowNo int64
	if err := q.classListFilterBase(grade, majorName, className).
		Where(`
			class.grade > ?
			OR (class.grade = ? AND class.major_name < ?)
			OR (class.grade = ? AND class.major_name = ? AND class.class_name < ?)
			OR (class.grade = ? AND class.major_name = ? AND class.class_name = ? AND class.id <= ?)
		`,
			target.Grade,
			target.Grade, target.MajorName,
			target.Grade, target.MajorName, target.ClassName,
			target.Grade, target.MajorName, target.ClassName, target.ID,
		).
		Count(&rowNo).Error; err != nil {
		return FocusPageResult{}, err
	}
	if rowNo <= 0 {
		return FocusPageResult{}, nil
	}

	return FocusPageResult{
		Found:  true,
		Page:   int((rowNo-1)/int64(pageSize)) + 1,
		RowKey: target.ID,
	}, nil
}

func (q *ClassQuery) ClassByID(classID uint64) (model.Class, error) {
	var item model.Class
	err := q.db.Table("class").
		Select("class.id, class.class_name, class.grade, class.major_name, class.status, class.created_at, class.updated_at, COUNT(student.id) AS student_count").
		Joins("LEFT JOIN student ON student.class_id = class.id AND student.status = 1").
		Where("class.id = ?", classID).
		Group("class.id").
		Scan(&item).Error
	if err != nil {
		return model.Class{}, err
	}
	if item.ID == 0 {
		return model.Class{}, gorm.ErrRecordNotFound
	}
	return item, nil
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

func (q *ClassQuery) ClassOptions(keyword string) ([]ClassOptionItem, error) {
	query := q.db.Table("class").
		Select("class.id, class.class_name, class.grade, class.major_name").
		Where("class.status = 1")
	if keyword != "" {
		query = query.Where("class.class_name LIKE ? OR class.major_name LIKE ? OR CAST(class.grade AS TEXT) LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	var items []ClassOptionItem
	err := query.Order("class.grade DESC, class.major_name ASC, class.class_name ASC").Scan(&items).Error
	return items, err
}
