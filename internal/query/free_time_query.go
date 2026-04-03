package query

import (
	"strconv"
	"strings"
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

type FreeTimeEditorItem struct {
	ID        uint64 `json:"id"`
	Term      string `json:"term"`
	Weekday   int    `json:"weekday"`
	Section   int    `json:"section"`
	FreeWeeks string `json:"free_weeks"`
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

func (q *FreeTimeQuery) Calendar(term string, weekNo int) ([]FreeTimeItem, error) {
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
	if weekNo <= 0 {
		return items, nil
	}
	filtered := items[:0]
	for _, item := range items {
		if freeTimeContainsWeek(item.FreeWeeks, weekNo) {
			filtered = append(filtered, item)
		}
	}
	return filtered, nil
}

func freeTimeContainsWeek(value string, weekNo int) bool {
	if weekNo <= 0 {
		return true
	}
	for _, chunk := range strings.Split(value, ",") {
		part := strings.TrimSpace(chunk)
		if part == "" {
			continue
		}
		bounds := strings.SplitN(part, "-", 2)
		start, err := strconv.Atoi(strings.TrimSpace(bounds[0]))
		if err != nil {
			continue
		}
		end := start
		if len(bounds) == 2 {
			parsedEnd, err := strconv.Atoi(strings.TrimSpace(bounds[1]))
			if err != nil {
				continue
			}
			end = parsedEnd
		}
		if start > end {
			start, end = end, start
		}
		if weekNo >= start && weekNo <= end {
			return true
		}
	}
	return false
}

func (q *FreeTimeQuery) Editor(term, loginID string, userID uint64, restrictToUser bool) ([]FreeTimeEditorItem, error) {
	query := q.db.Table("user_free_time").
		Select("user_free_time.id, term.name AS term, user_free_time.weekday, user_free_time.section, user_free_time.free_weeks").
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
	var items []FreeTimeEditorItem
	if err := query.Order("user_free_time.id DESC").Scan(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}
