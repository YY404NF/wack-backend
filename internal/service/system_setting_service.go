package service

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"wack-backend/internal/model"
)

type SystemSettingService struct {
	db *gorm.DB
}

func NewSystemSettingService(db *gorm.DB) *SystemSettingService {
	return &SystemSettingService{db: db}
}

func normalizeSchedule(schedule string) string {
	switch schedule {
	case "winter":
		return "autumn"
	case "summer", "autumn":
		return schedule
	default:
		return inferScheduleByDate(time.Now())
	}
}

func inferScheduleByDate(now time.Time) string {
	month := now.Month()
	day := now.Day()
	if month > time.May && month < time.October {
		return "summer"
	}
	if month == time.May && day >= 1 {
		return "summer"
	}
	if month == time.October && day >= 1 {
		return "autumn"
	}
	if month > time.October || month < time.May {
		return "autumn"
	}
	return "autumn"
}

func (s *SystemSettingService) resolveActiveTerm(now time.Time) (model.Term, error) {
	var term model.Term
	today := now.Format("2006-01-02")
	err := s.db.
		Where("term_start_date <= ?", today).
		Order("term_start_date DESC, id DESC").
		First(&term).Error
	switch {
	case err == nil:
		return term, nil
	case err != gorm.ErrRecordNotFound:
		return model.Term{}, err
	}

	err = s.db.
		Order("term_start_date ASC, id ASC").
		First(&term).Error
	if err != nil {
		return model.Term{}, err
	}
	return term, nil
}

func (s *SystemSettingService) GetSystemSetting() (model.SystemSetting, error) {
	term, err := s.resolveActiveTerm(time.Now())
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return model.SystemSetting{
				CurrentTermStartDate: "",
				CurrentSchedule:      inferScheduleByDate(time.Now()),
			}, nil
		}
		return model.SystemSetting{}, err
	}

	return model.SystemSetting{
		ID:                   term.ID,
		CurrentTermStartDate: term.TermStartDate,
		CurrentSchedule:      normalizeSchedule(inferScheduleByDate(time.Now())),
		CreatedAt:            term.CreatedAt,
		UpdatedAt:            term.UpdatedAt,
	}, nil
}

func (s *SystemSettingService) UpdateSystemSetting(startDate string) (model.SystemSetting, error) {
	startDate = strings.TrimSpace(startDate)
	if startDate == "" {
		return model.SystemSetting{}, ErrInvalidInput
	}
	parsedDate, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return model.SystemSetting{}, ErrInvalidInput
	}
	if parsedDate.Weekday() != time.Monday {
		return model.SystemSetting{}, ErrInvalidInput
	}

	term, err := s.resolveActiveTerm(time.Now())
	switch {
	case err == nil:
		if err := s.db.Model(&term).Update("term_start_date", startDate).Error; err != nil {
			return model.SystemSetting{}, err
		}
		if err := s.db.First(&term, term.ID).Error; err != nil {
			return model.SystemSetting{}, err
		}
	case err == gorm.ErrRecordNotFound:
		term = model.Term{
			Name:          buildTermName(parsedDate),
			TermStartDate: startDate,
		}
		if err := s.db.Create(&term).Error; err != nil {
			return model.SystemSetting{}, err
		}
	default:
		return model.SystemSetting{}, err
	}

	return model.SystemSetting{
		ID:                   term.ID,
		CurrentTermStartDate: term.TermStartDate,
		CurrentSchedule:      inferScheduleByDate(time.Now()),
		CreatedAt:            term.CreatedAt,
		UpdatedAt:            term.UpdatedAt,
	}, nil
}

func buildTermName(start time.Time) string {
	year := start.Year()
	month := start.Month()
	termNo := 1
	if month >= time.July {
		termNo = 1
	} else {
		termNo = 2
		year = year - 1
	}
	return fmt.Sprintf("%d-%d-%d", year, year+1, termNo)
}
