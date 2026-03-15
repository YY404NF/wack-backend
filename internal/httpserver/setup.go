package httpserver

import (
	"errors"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"wack-backend/internal/model"
)

func (h *authHandler) setupStatus(c *gin.Context) {
	initialized, err := h.hasAnyAdmin()
	if err != nil {
		fail(c, 500, "load setup status failed")
		return
	}

	ok(c, gin.H{
		"initialized": initialized,
	})
}

func (h *authHandler) initializeSystem(c *gin.Context) {
	var req struct {
		StudentID string `json:"student_id" binding:"required"`
		RealName  string `json:"real_name" binding:"required"`
		Password  string `json:"password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "invalid request")
		return
	}

	initialized, err := h.hasAnyAdmin()
	if err != nil {
		fail(c, 500, "load setup status failed")
		return
	}
	if initialized {
		fail(c, 409, "system already initialized")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		fail(c, 500, "hash password failed")
		return
	}

	admin := model.User{
		StudentID:    req.StudentID,
		PasswordHash: string(hash),
		RealName:     req.RealName,
		Role:         model.RoleAdmin,
		Status:       model.UserStatusActive,
	}

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		var total int64
		if err := tx.Model(&model.User{}).Count(&total).Error; err != nil {
			return err
		}
		if total > 0 {
			return errors.New("system already initialized")
		}
		return tx.Create(&admin).Error
	}); err != nil {
		if err.Error() == "system already initialized" {
			fail(c, 409, "system already initialized")
			return
		}
		fail(c, 500, "initialize system failed")
		return
	}

	ok(c, gin.H{
		"initialized": true,
	})
}

func (h *authHandler) hasAnyAdmin() (bool, error) {
	var count int64
	if err := h.db.Model(&model.User{}).Where("role = ?", model.RoleAdmin).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
