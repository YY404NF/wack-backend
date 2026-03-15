package service

import (
	"gorm.io/gorm"

	"wack-backend/internal/model"
	"wack-backend/internal/query"
)

type FreeTimeService struct {
	db        *gorm.DB
	freeTimes *query.FreeTimeQuery
}

func NewFreeTimeService(db *gorm.DB) *FreeTimeService {
	return &FreeTimeService{db: db, freeTimes: query.NewFreeTimeQuery(db)}
}

func (s *FreeTimeService) ListFreeTimes(term, studentID string, currentUser model.User, page, pageSize int) ([]query.FreeTimeItem, int64, error) {
	return s.freeTimes.List(term, studentID, currentUser.ID, currentUser.Role == model.RoleStudent, page, pageSize)
}

func (s *FreeTimeService) FreeTimeCalendar(term string) ([]query.FreeTimeItem, error) {
	return s.freeTimes.Calendar(term)
}

func (s *FreeTimeService) CreateFreeTime(item model.StudentFreeTime) (model.StudentFreeTime, error) {
	return item, s.db.Create(&item).Error
}

func (s *FreeTimeService) GetFreeTime(id uint64) (model.StudentFreeTime, error) {
	var item model.StudentFreeTime
	if err := s.db.First(&item, id).Error; err != nil {
		return model.StudentFreeTime{}, ErrFreeTimeNotFound
	}
	return item, nil
}

func (s *FreeTimeService) UpdateFreeTime(id uint64, term string, userID uint64, weekday, section int, freeWeeks string) error {
	var item model.StudentFreeTime
	if err := s.db.First(&item, id).Error; err != nil {
		return ErrFreeTimeNotFound
	}
	return s.db.Model(&item).Updates(map[string]interface{}{
		"term":       term,
		"user_id":    userID,
		"weekday":    weekday,
		"section":    section,
		"free_weeks": freeWeeks,
	}).Error
}

func (s *FreeTimeService) DeleteFreeTime(id uint64) error {
	var item model.StudentFreeTime
	if err := s.db.First(&item, id).Error; err != nil {
		return ErrFreeTimeNotFound
	}
	return s.db.Delete(&item).Error
}
