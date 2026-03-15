package httpserver

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"wack-backend/internal/config"
	"wack-backend/internal/model"
)

type jwtClaims struct {
	StudentID string `json:"student_id"`
	Role      int    `json:"role"`
	jwt.RegisteredClaims
}

type authHandler struct {
	cfg config.Config
	db  *gorm.DB
}

func newAuthHandler(cfg config.Config, db *gorm.DB) *authHandler {
	return &authHandler{cfg: cfg, db: db}
}

func (h *authHandler) login(c *gin.Context) {
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

	token, err := h.signToken(user.StudentID, user.Role)
	if err != nil {
		fail(c, 500, "sign token failed")
		return
	}

	ok(c, gin.H{
		"token": token,
		"user": gin.H{
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
	if err := h.db.First(&dbUser, "student_id = ?", user.StudentID).Error; err != nil {
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

func (h *authHandler) signToken(studentID string, role int) (string, error) {
	claims := jwtClaims{
		StudentID: studentID,
		Role:      role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.cfg.JWTSecret))
}

func authMiddleware(cfg config.Config, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			fail(c, 401, "missing authorization header")
			c.Abort()
			return
		}
		tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer"))
		if tokenString == authHeader {
			fail(c, 401, "invalid authorization header")
			c.Abort()
			return
		}

		token, err := jwt.ParseWithClaims(tokenString, &jwtClaims{}, func(token *jwt.Token) (interface{}, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(cfg.JWTSecret), nil
		})
		if err != nil || !token.Valid {
			fail(c, 401, "invalid token")
			c.Abort()
			return
		}

		claims, okCast := token.Claims.(*jwtClaims)
		if !okCast {
			fail(c, 401, "invalid token claims")
			c.Abort()
			return
		}

		var user model.User
		if err := db.First(&user, "student_id = ?", claims.StudentID).Error; err != nil {
			fail(c, 401, "user not found")
			c.Abort()
			return
		}
		if user.Status != model.UserStatusActive {
			fail(c, 403, "user is frozen")
			c.Abort()
			return
		}

		c.Set("currentUser", user)
		c.Next()
	}
}

func requireRole(roles ...int) gin.HandlerFunc {
	allowed := make(map[int]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}
	return func(c *gin.Context) {
		user, exists := currentUser(c)
		if !exists {
			fail(c, 401, "unauthorized")
			c.Abort()
			return
		}
		if _, ok := allowed[user.Role]; !ok {
			fail(c, 403, fmt.Sprintf("role %d is not allowed", user.Role))
			c.Abort()
			return
		}
		c.Next()
	}
}

func currentUser(c *gin.Context) (model.User, bool) {
	value, exists := c.Get("currentUser")
	if !exists {
		return model.User{}, false
	}
	user, ok := value.(model.User)
	return user, ok
}
