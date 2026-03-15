package httpserver

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"

	"wack-backend/internal/config"
	"wack-backend/internal/httpserver/dto"
	"wack-backend/internal/service"
)

type authHandler struct {
	cfg  config.Config
	auth *service.AuthService
}

func newAuthHandler(cfg config.Config, db *gorm.DB) *authHandler {
	return &authHandler{cfg: cfg, auth: service.NewAuthService(db)}
}

func (h *authHandler) login(c *gin.Context) {
	initialized, err := h.auth.HasAnyAdmin()
	if err != nil {
		fail(c, 500, "load setup status failed")
		return
	}
	if !initialized {
		fail(c, 403, "system is not initialized")
		return
	}

	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "invalid request")
		return
	}

	user, err := h.auth.Authenticate(req.StudentID, req.Password)
	if err != nil {
		switch {
		case service.IsServiceError(err, service.ErrInvalidCredentials):
			fail(c, 401, "invalid credentials")
		case service.IsServiceError(err, service.ErrUserFrozen):
			fail(c, 403, "user is frozen")
		default:
			fail(c, 500, "authenticate failed")
		}
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

	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "invalid request")
		return
	}

	if err := h.auth.ChangePassword(user.ID, req.OldPassword, req.NewPassword); err != nil {
		switch {
		case service.IsServiceError(err, service.ErrUserNotFound):
			fail(c, 404, "user not found")
		case service.IsServiceError(err, service.ErrOldPasswordIncorrect):
			fail(c, 400, "old password is incorrect")
		default:
			fail(c, 500, "update password failed")
		}
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

	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "invalid request")
		return
	}

	updatedUser, err := h.auth.UpdateProfile(user.ID, req.StudentID, req.RealName)
	if err != nil {
		if service.IsServiceError(err, service.ErrUserNotFound) {
			fail(c, 404, "user not found")
			return
		}
		fail(c, 400, "update profile failed")
		return
	}
	ok(c, updatedUser)
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
