package service

import (
	"gorm.io/gorm"

	"wack-backend/internal/model"
)

type ClassService struct {
	db *gorm.DB
}

func NewClassService(db *gorm.DB) *ClassService {
	return &ClassService{db: db}
}

func (s *ClassService) ListClasses(page, pageSize int) ([]model.Class, int64, error) {
	query := s.db.Model(&model.Class{})
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var classes []model.Class
	if err := s.db.Model(&model.Class{}).
		Select("class.id, class.class_name, class.grade, class.major_name, class.created_at, class.updated_at, COUNT(user_class.id) AS student_count").
		Joins("LEFT JOIN user_class ON user_class.class_id = class.id").
		Group("class.id").
		Order("class.grade DESC, class.major_name ASC, class.class_name ASC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&classes).Error; err != nil {
		return nil, 0, err
	}
	return classes, total, nil
}

func (s *ClassService) CreateClass(class model.Class) (model.Class, error) {
	return class, s.db.Create(&class).Error
}

func (s *ClassService) GetClass(id uint64) (model.Class, error) {
	var class model.Class
	if err := s.db.First(&class, id).Error; err != nil {
		return model.Class{}, ErrClassNotFound
	}
	return class, nil
}

func (s *ClassService) UpdateClass(id uint64, req model.Class) (model.Class, error) {
	var class model.Class
	if err := s.db.First(&class, id).Error; err != nil {
		return model.Class{}, ErrClassNotFound
	}
	if err := s.db.Model(&class).Updates(map[string]interface{}{
		"class_name": req.ClassName,
		"grade":      req.Grade,
		"major_name": req.MajorName,
	}).Error; err != nil {
		return model.Class{}, err
	}
	if err := s.db.First(&class, id).Error; err != nil {
		return model.Class{}, err
	}
	return class, nil
}

func (s *ClassService) DeleteClass(id uint64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("class_id = ?", id).Delete(&model.UserClass{}).Error; err != nil {
			return err
		}
		if err := tx.Where("class_id = ?", id).Delete(&model.CourseClass{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.Class{}, id).Error
	})
}

func (s *ClassService) GetClassStudents(id uint64) ([]model.User, error) {
	var users []model.User
	err := s.db.Table("user").
		Joins("JOIN user_class ON user_class.user_id = user.id").
		Where("user_class.class_id = ?", id).
		Find(&users).Error
	return users, err
}
