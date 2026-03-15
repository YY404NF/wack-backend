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
	StudentID string
	RealName  string
	Password  string
	Role      int
	Status    int
	ClassIDs  []uint64
}

type UpdateUserInput struct {
	StudentID string
	RealName  string
	Role      int
	Status    int
	ClassIDs  []uint64
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
		query = query.Where("student_id LIKE ? OR real_name LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
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
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return model.User{}, err
	}

	user := model.User{
		StudentID:    input.StudentID,
		PasswordHash: string(hash),
		RealName:     input.RealName,
		Role:         input.Role,
		Status:       input.Status,
	}
	if user.Status == 0 {
		user.Status = model.UserStatusActive
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&user).Error; err != nil {
			return err
		}
		if len(input.ClassIDs) == 0 {
			return nil
		}
		relations := make([]model.UserClass, 0, len(input.ClassIDs))
		for _, classID := range input.ClassIDs {
			relations = append(relations, model.UserClass{UserID: user.ID, ClassID: classID})
		}
		return tx.Create(&relations).Error
	})
	return user, err
}

func (s *UserService) GetUserWithClasses(studentID string) (model.User, []model.UserClass, error) {
	var user model.User
	if err := s.db.First(&user, "student_id = ?", studentID).Error; err != nil {
		return model.User{}, nil, ErrUserNotFound
	}
	var classes []model.UserClass
	if err := s.db.Where("user_id = ?", user.ID).Find(&classes).Error; err != nil {
		return model.User{}, nil, err
	}
	return user, classes, nil
}

func (s *UserService) UpdateUser(currentUserID uint64, targetStudentID string, input UpdateUserInput) (model.User, []model.UserClass, error) {
	var user model.User
	if err := s.db.First(&user, "student_id = ?", targetStudentID).Error; err != nil {
		return model.User{}, nil, ErrUserNotFound
	}
	if currentUserID == user.ID && input.Role != model.RoleAdmin {
		return model.User{}, nil, ErrAdminRemoveOwnRole
	}
	if currentUserID == user.ID && input.Status != model.UserStatusActive {
		return model.User{}, nil, ErrAdminFreezeSelf
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&user).Updates(map[string]interface{}{
			"student_id": input.StudentID,
			"real_name":  input.RealName,
			"role":       input.Role,
			"status":     input.Status,
		}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", user.ID).Delete(&model.UserClass{}).Error; err != nil {
			return err
		}
		if len(input.ClassIDs) == 0 {
			return nil
		}
		relations := make([]model.UserClass, 0, len(input.ClassIDs))
		for _, classID := range input.ClassIDs {
			relations = append(relations, model.UserClass{UserID: user.ID, ClassID: classID})
		}
		return tx.Create(&relations).Error
	})
	if err != nil {
		return model.User{}, nil, err
	}

	return s.GetUserWithClasses(input.StudentID)
}

func (s *UserService) ResetUserPassword(currentUserID uint64, targetStudentID, newPassword string) error {
	var user model.User
	if err := s.db.First(&user, "student_id = ?", targetStudentID).Error; err != nil {
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

func (s *UserService) UpdateUserStatus(currentUserID uint64, targetStudentID string, status int) error {
	var user model.User
	if err := s.db.First(&user, "student_id = ?", targetStudentID).Error; err != nil {
		return ErrUserNotFound
	}
	if currentUserID == user.ID {
		return ErrAdminFreezeSelf
	}
	return s.db.Model(&model.User{}).Where("id = ?", user.ID).Update("status", status).Error
}

func (s *UserService) UpdateUserRole(currentUserID uint64, targetStudentID string, role int) error {
	var user model.User
	if err := s.db.First(&user, "student_id = ?", targetStudentID).Error; err != nil {
		return ErrUserNotFound
	}
	if currentUserID == user.ID {
		return ErrAdminRemoveOwnRole
	}
	return s.db.Model(&model.User{}).Where("id = ?", user.ID).Update("role", role).Error
}
