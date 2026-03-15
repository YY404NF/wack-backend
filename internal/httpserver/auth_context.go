package httpserver

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"

	"wack-backend/internal/config"
	"wack-backend/internal/model"
)

type jwtClaims struct {
	UserID uint64 `json:"user_id"`
	Role   int    `json:"role"`
	jwt.RegisteredClaims
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
		if err := db.First(&user, claims.UserID).Error; err != nil {
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
