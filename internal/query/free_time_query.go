package query

import (
	"time"

	"gorm.io/gorm"

	"wack-backend/internal/model"
)

type FreeTimeItem struct {
	ID        uint64    `json:"id"`
	Term      string    `json:"term"`
	UserID    uint64    `json:"user_id"`
	LoginID   string    `json:"login_id"`
	RealName  string    `json:"real_name"`
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

func (q *FreeTimeQuery) List(term, loginID string, userID uint64, restrictToUser bool, page, pageSize int) ([]FreeTimeItem, int64, error) {
	query := q.db.Table("user_free_time").
		Select("user_free_time.id, term.name AS term, user_free_time.user_id, user.login_id, user.real_name, user_free_time.weekday, user_free_time.section, user_free_time.free_weeks, user_free_time.created_at, user_free_time.updated_at").
		Joins("JOIN user ON user.id = user_free_time.user_id").
		Joins("JOIN term ON term.id = user_free_time.term_id").
		Where("user.role = ?", model.RoleStudent)
	if term != "" {
		query = query.Where("term.name = ?", term)
	}
	if loginID != "" {
		query = query.Where("user.login_id = ?", loginID)
	} else if restrictToUser {
		query = query.Where("user_free_time.user_id = ?", userID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []FreeTimeItem
	if err := query.Order("user_free_time.id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Scan(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (q *FreeTimeQuery) Calendar(term string) ([]FreeTimeItem, error) {
	query := q.db.Table("user_free_time").
		Select("user_free_time.id, term.name AS term, user_free_time.user_id, user.login_id, user.real_name, user_free_time.weekday, user_free_time.section, user_free_time.free_weeks, user_free_time.created_at, user_free_time.updated_at").
		Joins("JOIN user ON user.id = user_free_time.user_id").
		Joins("JOIN term ON term.id = user_free_time.term_id").
		Where("user.role = ?", model.RoleStudent)
	if term != "" {
		query = query.Where("term.name = ?", term)
	}
	var items []FreeTimeItem
	if err := query.Order("weekday, section, user.login_id").Scan(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}
