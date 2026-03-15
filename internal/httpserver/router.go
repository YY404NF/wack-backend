package httpserver

import (
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"wack-backend/internal/config"
)

func NewRouter(cfg config.Config, db *gorm.DB) (*gin.Engine, error) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery(), corsMiddleware(cfg))

	authHandler := newAuthHandler(cfg, db)
	apiHandler := newAPIHandler(db)
	authMW := authMiddleware(cfg, db)

	api := router.Group("/api")
	{
		mountSetupRoutes(api, authHandler)
		mountAuthRoutes(api, authMW, authHandler)

		protected := api.Group("")
		protected.Use(authMW)
		mountFreeTimeRoutes(protected, apiHandler)

		admin := protected.Group("")
		admin.Use(requireRole(1))
		mountUserRoutes(admin, apiHandler)
		mountClassRoutes(admin, apiHandler)
		mountCourseRoutes(admin, apiHandler)

		student := protected.Group("")
		student.Use(requireRole(2))
		mountAttendanceRoutes(admin, student, apiHandler)
	}

	router.GET("/healthz", func(c *gin.Context) {
		ok(c, gin.H{
			"time": time.Now().Format("2006-01-02 15:04:05"),
		})
	})

	return router, nil
}
