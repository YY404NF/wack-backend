package query

import (
	"time"

	"gorm.io/gorm"
)

type FreeTimeItem struct {
	ID        uint64    `json:"id"`
	Term      string    `json:"term"`
	UserID    uint64    `json:"user_id"`
	StudentID string    `json:"student_id"`
	Weekday   int       `json:"weekday"`
	Section   int       `json:"section"`
	FreeWeeks string    `json:"free_weeks"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type FreeTimeQuery struct {
	db *gorm.DB
}

func NewFreeTimeQuery(db *gorm.DB) *FreeTimeQuery {
	return &FreeTimeQuery{db: db}
}

func (q *FreeTimeQuery) List(term, studentID string, userID uint64, restrictToUser bool, page, pageSize int) ([]FreeTimeItem, int64, error) {
	query := q.db.Table("student_free_time").
		Select("student_free_time.id, student_free_time.term, student_free_time.user_id, user.student_id, student_free_time.weekday, student_free_time.section, student_free_time.free_weeks, student_free_time.created_at, student_free_time.updated_at").
		Joins("JOIN user ON user.id = student_free_time.user_id")
	if term != "" {
		query = query.Where("student_free_time.term = ?", term)
	}
	if studentID != "" {
		query = query.Where("user.student_id = ?", studentID)
	} else if restrictToUser {
		query = query.Where("student_free_time.user_id = ?", userID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []FreeTimeItem
	if err := query.Order("student_free_time.id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Scan(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (q *FreeTimeQuery) Calendar(term string) ([]FreeTimeItem, error) {
	query := q.db.Table("student_free_time").
		Select("student_free_time.id, student_free_time.term, student_free_time.user_id, user.student_id, student_free_time.weekday, student_free_time.section, student_free_time.free_weeks, student_free_time.created_at, student_free_time.updated_at").
		Joins("JOIN user ON user.id = student_free_time.user_id")
	if term != "" {
		query = query.Where("student_free_time.term = ?", term)
	}
	var items []FreeTimeItem
	if err := query.Order("weekday, section, user.student_id").Scan(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}
