package httpserver

import (
	"strings"

	"github.com/gin-gonic/gin"

	"wack-backend/internal/config"
)

func corsMiddleware(cfg config.Config) gin.HandlerFunc {
	allowedOrigins := splitOrigins(cfg.CORSAllowOrigin)

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" {
			if len(allowedOrigins) == 0 {
				c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			} else {
				for _, allowedOrigin := range allowedOrigins {
					if origin == allowedOrigin {
						c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
						break
					}
				}
			}
		}

		c.Writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Vary", "Origin")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func splitOrigins(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}

	parts := strings.Split(value, ",")
	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			origins = append(origins, trimmed)
		}
	}
	return origins
}
