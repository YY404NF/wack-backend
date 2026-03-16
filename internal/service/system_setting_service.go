package service

import (
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

func (s *SystemSettingService) GetSystemSetting() (model.SystemSetting, error) {
	var setting model.SystemSetting
	if err := s.db.Order("id ASC").First(&setting).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			setting = model.SystemSetting{
				CurrentTermStartDate: "",
				CurrentSchedule:      inferScheduleByDate(time.Now()),
			}
			if err := s.db.Create(&setting).Error; err != nil {
				return model.SystemSetting{}, err
			}
			return setting, nil
		}
		return model.SystemSetting{}, err
	}
	expectedSchedule := inferScheduleByDate(time.Now())
	normalizedSchedule := normalizeSchedule(setting.CurrentSchedule)
	if normalizedSchedule != expectedSchedule {
		if err := s.db.Model(&setting).Update("current_schedule", expectedSchedule).Error; err != nil {
			return model.SystemSetting{}, err
		}
		setting.CurrentSchedule = expectedSchedule
		return setting, nil
	}
	if normalizedSchedule != setting.CurrentSchedule {
		if err := s.db.Model(&setting).Update("current_schedule", normalizedSchedule).Error; err != nil {
			return model.SystemSetting{}, err
		}
		setting.CurrentSchedule = normalizedSchedule
	}
	return setting, nil
}

func (s *SystemSettingService) UpdateSystemSetting(startDate string) (model.SystemSetting, error) {
	setting, err := s.GetSystemSetting()
	if err != nil {
		return model.SystemSetting{}, err
	}
	if err := s.db.Model(&setting).Updates(map[string]any{
		"current_term_start_date": startDate,
	}).Error; err != nil {
		return model.SystemSetting{}, err
	}
	if err := s.db.First(&setting, setting.ID).Error; err != nil {
		return model.SystemSetting{}, err
	}
	return setting, nil
}
