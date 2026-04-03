package service

import (
	"strings"

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

func (s *FreeTimeService) ListFreeTimeEditor(term, loginID string, currentUser model.User) ([]query.FreeTimeEditorItem, error) {
	return s.freeTimes.Editor(strings.TrimSpace(term), strings.TrimSpace(loginID), currentUser.ID, currentUser.Role == model.RoleStudent)
}

func (s *FreeTimeService) FreeTimeCalendar(term string, weekNo int) ([]query.FreeTimeItem, error) {
	return s.freeTimes.Calendar(term, weekNo)
}

func (s *FreeTimeService) CreateFreeTime(termName string, userID uint64, weekday, section int, freeWeeks string) (model.UserFreeTime, error) {
	termName = strings.TrimSpace(termName)
	freeWeeks = strings.TrimSpace(freeWeeks)
	term, err := s.resolveTerm(termName)
	if err != nil {
		return model.UserFreeTime{}, err
	}
	item := model.UserFreeTime{
		TermID:    term.ID,
		UserID:    userID,
		Weekday:   weekday,
		Section:   section,
		FreeWeeks: strings.TrimSpace(freeWeeks),
	}
	if err := s.validateFreeTimeInput(term.ID, userID, weekday, section, item.FreeWeeks); err != nil {
		return model.UserFreeTime{}, err
	}
	return item, s.db.Create(&item).Error
}

func (s *FreeTimeService) GetFreeTime(id uint64) (model.UserFreeTime, error) {
	var item model.UserFreeTime
	if err := s.db.First(&item, id).Error; err != nil {
		return model.UserFreeTime{}, ErrFreeTimeNotFound
	}
	return item, nil
}

func (s *FreeTimeService) UpdateFreeTime(id uint64, termName string, userID uint64, weekday, section int, freeWeeks string) error {
	var item model.UserFreeTime
	if err := s.db.First(&item, id).Error; err != nil {
		return ErrFreeTimeNotFound
	}
	termName = strings.TrimSpace(termName)
	freeWeeks = strings.TrimSpace(freeWeeks)
	term, err := s.resolveTerm(termName)
	if err != nil {
		return err
	}
	if err := s.validateFreeTimeInput(term.ID, userID, weekday, section, freeWeeks); err != nil {
		return err
	}
	return s.db.Model(&item).Updates(map[string]interface{}{
		"term_id":    term.ID,
		"user_id":    userID,
		"weekday":    weekday,
		"section":    section,
		"free_weeks": freeWeeks,
	}).Error
}

func (s *FreeTimeService) DeleteFreeTime(id uint64) error {
	var item model.UserFreeTime
	if err := s.db.First(&item, id).Error; err != nil {
		return ErrFreeTimeNotFound
	}
	return s.db.Delete(&item).Error
}

func (s *FreeTimeService) validateFreeTimeInput(termID uint64, userID uint64, weekday, section int, freeWeeks string) error {
	if termID == 0 || freeWeeks == "" {
		return ErrInvalidInput
	}
	if len(freeWeeks) > 100 {
		return ErrInvalidInput
	}
	if weekday < 1 || weekday > 7 || section < 1 || section > 5 {
		return ErrInvalidInput
	}

	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return ErrUserNotFound
	}
	if user.Role != model.RoleStudent {
		return ErrInvalidInput
	}

	return nil
}

func (s *FreeTimeService) resolveTerm(name string) (model.Term, error) {
	if name == "" {
		return model.Term{}, ErrInvalidInput
	}
	var term model.Term
	if err := s.db.First(&term, "name = ?", name).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return model.Term{}, ErrInvalidInput
		}
		return model.Term{}, err
	}
	return term, nil
}
