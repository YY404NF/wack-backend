package httpserver

import (
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"wack-backend/internal/config"
	"wack-backend/internal/service"
)

func NewRouter(cfg config.Config, db *gorm.DB, sessions *service.SessionService) (*gin.Engine, error) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery(), corsMiddleware(cfg))

	authHandler := newAuthHandler(cfg, db, sessions)
	apiHandler := newAPIHandler(db, sessions)
	authMW := authMiddleware(cfg, db, sessions)

	api := router.Group("/api")
	{
		mountSetupRoutes(api, authHandler)
		mountAuthRoutes(api, authMW, authHandler)

		protected := api.Group("")
		protected.Use(authMW)
		mountFreeTimeRoutes(protected, apiHandler)
		mountMetaRoutes(protected, apiHandler)

		admin := protected.Group("/admin")
		admin.Use(requireRole(1))
		mountUserRoutes(admin, apiHandler)
		mountClassRoutes(admin, apiHandler)
		mountStudentRoutes(admin, apiHandler)
		mountCourseRoutes(admin, apiHandler)
		mountSystemSettingRoutes(admin, apiHandler)

		student := protected.Group("/student")
		student.Use(requireRole(2, 3))
		mountAttendanceRoutes(admin, student, apiHandler)
	}

	router.GET("/healthz", func(c *gin.Context) {
		ok(c, gin.H{
			"time": time.Now().Format("2006-01-02 15:04:05"),
		})
	})

	return router, nil
}
