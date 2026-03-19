package httpserver

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"

	"wack-backend/internal/config"
	"wack-backend/internal/httpserver/dto"
	"wack-backend/internal/service"
)

type authHandler struct {
	cfg      config.Config
	auth     *service.AuthService
	sessions *service.SessionService
}

func newAuthHandler(cfg config.Config, db *gorm.DB, sessions *service.SessionService) *authHandler {
	return &authHandler{
		cfg:      cfg,
		auth:     service.NewAuthService(db),
		sessions: sessions,
	}
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

	user, err := h.auth.Authenticate(req.LoginID, req.Password)
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

	token, tokenID, expiresAt, err := h.signToken(user.ID, user.Role)
	if err != nil {
		fail(c, 500, "sign token failed")
		return
	}

	if err := h.sessions.CreateSession(c.Request.Context(), service.SessionPayload{
		TokenID:    tokenID,
		UserID:     user.ID,
		Account:    user.LoginID,
		Role:       user.Role,
		Status:     user.Status,
		DeviceType: service.DeviceTypeForRole(user.Role),
		IssuedAt:   time.Now(),
		ExpiresAt:  expiresAt,
	}); err != nil {
		fail(c, 500, "create session failed")
		return
	}

	ok(c, gin.H{
		"token": token,
		"user": gin.H{
			"id":               user.ID,
			"login_id":         user.LoginID,
			"real_name":        user.RealName,
			"role":             user.Role,
			"status":           user.Status,
			"managed_class_id": user.ManagedClassID,
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

func (h *authHandler) logout(c *gin.Context) {
	const bearerPrefix = "Bearer "
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, bearerPrefix) {
		tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, bearerPrefix))
		if tokenString != "" {
			token, err := jwt.ParseWithClaims(tokenString, &jwtClaims{}, func(token *jwt.Token) (interface{}, error) {
				return []byte(h.cfg.JWTSecret), nil
			})
			if err == nil && token != nil && token.Valid {
				if claims, ok := token.Claims.(*jwtClaims); ok && claims.ID != "" {
					if err := h.sessions.DeleteSession(c.Request.Context(), claims.ID, claims.UserID); err != nil {
						fail(c, 500, "logout failed")
						return
					}
				}
			}
		}
	}
	ok(c, gin.H{})
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

	if err := h.sessions.DeleteAllUserSessions(c.Request.Context(), user.ID); err != nil {
		fail(c, 500, "clear sessions failed")
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

	updatedUser, err := h.auth.UpdateProfile(user.ID, req.LoginID, req.RealName)
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

func (h *authHandler) signToken(userID uint64, role int) (string, string, time.Time, error) {
	tokenID, err := newTokenID()
	if err != nil {
		return "", "", time.Time{}, err
	}
	now := time.Now()
	expiresAt := now.Add(service.SessionTTLForRole(role))
	claims := jwtClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenID,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(h.cfg.JWTSecret))
	if err != nil {
		return "", "", time.Time{}, err
	}
	return signedToken, tokenID, expiresAt, nil
}

func newTokenID() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
