package service

import (
	"gorm.io/gorm"

	"wack-backend/internal/model"
)

type SystemSettingService struct {
	db *gorm.DB
}

func NewSystemSettingService(db *gorm.DB) *SystemSettingService {
	return &SystemSettingService{db: db}
}

func (s *SystemSettingService) GetSystemSetting() (model.SystemSetting, error) {
	var setting model.SystemSetting
	if err := s.db.Order("id ASC").First(&setting).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			setting = model.SystemSetting{
				CurrentTermStartDate: "",
				CurrentSchedule:      "summer",
			}
			if err := s.db.Create(&setting).Error; err != nil {
				return model.SystemSetting{}, err
			}
			return setting, nil
		}
		return model.SystemSetting{}, err
	}
	return setting, nil
}

func (s *SystemSettingService) UpdateSystemSetting(startDate, schedule string) (model.SystemSetting, error) {
	setting, err := s.GetSystemSetting()
	if err != nil {
		return model.SystemSetting{}, err
	}
	if err := s.db.Model(&setting).Updates(map[string]any{
		"current_term_start_date": startDate,
		"current_schedule":        schedule,
	}).Error; err != nil {
		return model.SystemSetting{}, err
	}
	if err := s.db.First(&setting, setting.ID).Error; err != nil {
		return model.SystemSetting{}, err
	}
	return setting, nil
}
