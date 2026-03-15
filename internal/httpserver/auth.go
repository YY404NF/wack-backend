package httpserver

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"wack-backend/internal/config"
	"wack-backend/internal/model"
)

type authHandler struct {
	cfg config.Config
	db  *gorm.DB
}

func newAuthHandler(cfg config.Config, db *gorm.DB) *authHandler {
	return &authHandler{cfg: cfg, db: db}
}

func (h *authHandler) login(c *gin.Context) {
	initialized, err := h.hasAnyAdmin()
	if err != nil {
		fail(c, 500, "load setup status failed")
		return
	}
	if !initialized {
		fail(c, 403, "system is not initialized")
		return
	}

	var req struct {
		StudentID string `json:"student_id" binding:"required"`
		Password  string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "invalid request")
		return
	}

	var user model.User
	if err := h.db.First(&user, "student_id = ?", req.StudentID).Error; err != nil {
		fail(c, 401, "invalid credentials")
		return
	}
	if user.Status != model.UserStatusActive {
		fail(c, 403, "user is frozen")
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		fail(c, 401, "invalid credentials")
		return
	}

	token, err := h.signToken(user.ID, user.Role)
	if err != nil {
		fail(c, 500, "sign token failed")
		return
	}

	ok(c, gin.H{
		"token": token,
		"user": gin.H{
			"id":         user.ID,
			"student_id": user.StudentID,
			"real_name":  user.RealName,
			"role":       user.Role,
			"status":     user.Status,
		},
	})
}

func (h *authHandler) me(c *gin.Context) {
	user, exists := currentUser(c)
	if !exists {
		fail(c, 401, "unauthorized")
		return
	}
	ok(c, user)
}

func (h *authHandler) changePassword(c *gin.Context) {
	user, exists := currentUser(c)
	if !exists {
		fail(c, 401, "unauthorized")
		return
	}

	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "invalid request")
		return
	}

	var dbUser model.User
	if err := h.db.First(&dbUser, user.ID).Error; err != nil {
		fail(c, 404, "user not found")
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(dbUser.PasswordHash), []byte(req.OldPassword)); err != nil {
		fail(c, 400, "old password is incorrect")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		fail(c, 500, "hash password failed")
		return
	}

	if err := h.db.Model(&dbUser).Update("password_hash", string(hash)).Error; err != nil {
		fail(c, 500, "update password failed")
		return
	}
	ok(c, gin.H{})
}

func (h *authHandler) updateProfile(c *gin.Context) {
	user, exists := currentUser(c)
	if !exists {
		fail(c, 401, "unauthorized")
		return
	}

	var req struct {
		StudentID string `json:"student_id" binding:"required"`
		RealName  string `json:"real_name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "invalid request")
		return
	}

	var dbUser model.User
	if err := h.db.First(&dbUser, user.ID).Error; err != nil {
		fail(c, 404, "user not found")
		return
	}

	if err := h.db.Model(&dbUser).Updates(map[string]interface{}{
		"student_id": req.StudentID,
		"real_name":  req.RealName,
	}).Error; err != nil {
		fail(c, 400, "update profile failed")
		return
	}

	if err := h.db.First(&dbUser, user.ID).Error; err != nil {
		fail(c, 500, "reload profile failed")
		return
	}
	ok(c, dbUser)
}

func (h *authHandler) signToken(userID uint64, role int) (string, error) {
	claims := jwtClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.cfg.JWTSecret))
}
