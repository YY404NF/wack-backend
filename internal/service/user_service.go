package service

import (
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"wack-backend/internal/model"
)

type UserService struct {
	db *gorm.DB
}

type ListUsersInput struct {
	Page     int
	PageSize int
	Role     string
	Status   string
	Keyword  string
}

type CreateUserInput struct {
	LoginID        string
	RealName       string
	Password       string
	Role           int
	Status         int
	ManagedClassID *uint64
}

type UpdateUserInput struct {
	LoginID        string
	RealName       string
	Role           int
	Status         int
	ManagedClassID *uint64
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

func (s *UserService) ListUsers(input ListUsersInput) ([]model.User, int64, error) {
	query := s.db.Model(&model.User{})
	if input.Role != "" {
		query = query.Where("role = ?", input.Role)
	}
	if input.Status != "" {
		query = query.Where("status = ?", input.Status)
	}
	if keyword := strings.TrimSpace(input.Keyword); keyword != "" {
		query = query.Where("login_id LIKE ? OR real_name LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var users []model.User
	if err := query.Order("created_at DESC").Offset((input.Page - 1) * input.PageSize).Limit(input.PageSize).Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

func (s *UserService) CreateUser(input CreateUserInput) (model.User, error) {
	input.LoginID = strings.TrimSpace(input.LoginID)
	input.RealName = strings.TrimSpace(input.RealName)
	if input.LoginID == "" || input.RealName == "" {
		return model.User{}, ErrInvalidInput
	}
	if len(input.LoginID) > 32 || len(input.RealName) > 50 {
		return model.User{}, ErrInvalidInput
	}
	if input.Role != model.RoleAdmin && input.Role != model.RoleStudent && input.Role != model.RoleCommissioner {
		return model.User{}, ErrInvalidInput
	}
	if input.Status != 0 && input.Status != model.UserStatusActive && input.Status != model.UserStatusFrozen {
		return model.User{}, ErrInvalidInput
	}

	managedClassID, err := s.normalizeManagedClassID(input.Role, input.ManagedClassID)
	if err != nil {
		return model.User{}, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return model.User{}, err
	}

	user := model.User{
		LoginID:        input.LoginID,
		PasswordHash:   string(hash),
		RealName:       input.RealName,
		Role:           input.Role,
		Status:         input.Status,
		ManagedClassID: managedClassID,
	}
	if user.Status == 0 {
		user.Status = model.UserStatusActive
	}

	err = s.db.Create(&user).Error
	return user, err
}

func (s *UserService) GetUser(loginID string) (model.User, error) {
	var user model.User
	if err := s.db.First(&user, "login_id = ?", loginID).Error; err != nil {
		return model.User{}, ErrUserNotFound
	}
	return user, nil
}

func (s *UserService) UpdateUser(currentUserID uint64, targetLoginID string, input UpdateUserInput) (model.User, error) {
	input.LoginID = strings.TrimSpace(input.LoginID)
	input.RealName = strings.TrimSpace(input.RealName)
	if input.LoginID == "" || input.RealName == "" {
		return model.User{}, ErrInvalidInput
	}
	if len(input.LoginID) > 32 || len(input.RealName) > 50 {
		return model.User{}, ErrInvalidInput
	}
	if input.Role != model.RoleAdmin && input.Role != model.RoleStudent && input.Role != model.RoleCommissioner {
		return model.User{}, ErrInvalidInput
	}
	if input.Status != model.UserStatusActive && input.Status != model.UserStatusFrozen {
		return model.User{}, ErrInvalidInput
	}

	managedClassID, err := s.normalizeManagedClassID(input.Role, input.ManagedClassID)
	if err != nil {
		return model.User{}, err
	}

	var user model.User
	if err := s.db.First(&user, "login_id = ?", targetLoginID).Error; err != nil {
		return model.User{}, ErrUserNotFound
	}
	if currentUserID == user.ID && input.Role != model.RoleAdmin {
		return model.User{}, ErrAdminRemoveOwnRole
	}
	if currentUserID == user.ID && input.Status != model.UserStatusActive {
		return model.User{}, ErrAdminFreezeSelf
	}

	err = s.db.Model(&user).Updates(map[string]interface{}{
		"login_id":         input.LoginID,
		"real_name":        input.RealName,
		"role":             input.Role,
		"status":           input.Status,
		"managed_class_id": managedClassID,
	}).Error
	if err != nil {
		return model.User{}, err
	}

	return s.GetUser(input.LoginID)
}

func (s *UserService) ResetUserPassword(currentUserID uint64, targetLoginID, newPassword string) error {
	var user model.User
	if err := s.db.First(&user, "login_id = ?", targetLoginID).Error; err != nil {
		return ErrUserNotFound
	}
	if currentUserID == user.ID {
		return ErrAdminResetOwnPassword
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.db.Model(&user).Update("password_hash", string(hash)).Error
}

func (s *UserService) UpdateUserStatus(currentUserID uint64, targetLoginID string, status int) error {
	var user model.User
	if err := s.db.First(&user, "login_id = ?", targetLoginID).Error; err != nil {
		return ErrUserNotFound
	}
	if currentUserID == user.ID {
		return ErrAdminFreezeSelf
	}
	return s.db.Model(&model.User{}).Where("id = ?", user.ID).Update("status", status).Error
}

func (s *UserService) UpdateUserRole(currentUserID uint64, targetLoginID string, role int) error {
	var user model.User
	if err := s.db.First(&user, "login_id = ?", targetLoginID).Error; err != nil {
		return ErrUserNotFound
	}
	if currentUserID == user.ID {
		return ErrAdminRemoveOwnRole
	}
	managedClassID, err := s.normalizeManagedClassID(role, user.ManagedClassID)
	if err != nil {
		return err
	}
	return s.db.Model(&model.User{}).Where("id = ?", user.ID).Updates(map[string]interface{}{
		"role":             role,
		"managed_class_id": managedClassID,
	}).Error
}

func (s *UserService) normalizeManagedClassID(role int, managedClassID *uint64) (*uint64, error) {
	if role != model.RoleCommissioner {
		return nil, nil
	}
	if managedClassID == nil || *managedClassID == 0 {
		return nil, ErrInvalidInput
	}

	var count int64
	if err := s.db.Model(&model.Class{}).Where("id = ?", *managedClassID).Count(&count).Error; err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, ErrInvalidInput
	}
	return managedClassID, nil
}
