package service

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"wack-backend/internal/model"
)

type AuthService struct {
	db *gorm.DB
}

func NewAuthService(db *gorm.DB) *AuthService {
	return &AuthService{db: db}
}

func (s *AuthService) HasAnyAdmin() (bool, error) {
	var count int64
	if err := s.db.Model(&model.User{}).Where("role = ?", model.RoleAdmin).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *AuthService) InitializeSystem(loginID, realName, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	admin := model.User{
		LoginID:      loginID,
		PasswordHash: string(hash),
		RealName:     realName,
		Role:         model.RoleAdmin,
		Status:       model.UserStatusActive,
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		var total int64
		if err := tx.Model(&model.User{}).Count(&total).Error; err != nil {
			return err
		}
		if total > 0 {
			return ErrSystemAlreadyInitialized
		}
		return tx.Create(&admin).Error
	})
}

func (s *AuthService) Authenticate(loginID, password string) (model.User, error) {
	var user model.User
	if err := s.db.First(&user, "login_id = ?", loginID).Error; err != nil {
		return model.User{}, ErrInvalidCredentials
	}
	if user.Status != model.UserStatusActive {
		return model.User{}, ErrUserFrozen
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return model.User{}, ErrInvalidCredentials
	}
	now := time.Now()
	if err := s.db.Model(&user).Update("last_login_at", now).Error; err != nil {
		return model.User{}, err
	}
	user.LastLoginAt = &now
	return user, nil
}

func (s *AuthService) ChangePassword(userID uint64, oldPassword, newPassword string) error {
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return ErrUserNotFound
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return ErrOldPasswordIncorrect
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.db.Model(&user).Update("password_hash", string(hash)).Error
}

func (s *AuthService) UpdateProfile(userID uint64, loginID, realName string) (model.User, error) {
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return model.User{}, ErrUserNotFound
	}
	if err := s.db.Model(&user).Updates(map[string]interface{}{
		"login_id":  loginID,
		"real_name": realName,
	}).Error; err != nil {
		return model.User{}, err
	}
	if err := s.db.First(&user, userID).Error; err != nil {
		return model.User{}, err
	}
	return user, nil
}

func IsServiceError(err, target error) bool {
	return errors.Is(err, target)
}
