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
	router.Use(gin.Logger(), gin.Recovery())

	authHandler := newAuthHandler(cfg, db)
	apiHandler := newAPIHandler(db)

	api := router.Group("/api")
	{
		authGroup := api.Group("/auth")
		authGroup.POST("/login", authHandler.login)

		protected := api.Group("")
		protected.Use(authMiddleware(cfg, db))
		protected.GET("/auth/me", authHandler.me)
		protected.POST("/auth/change-password", authHandler.changePassword)

		protected.GET("/free-times", apiHandler.listFreeTimes)
		protected.POST("/free-times", apiHandler.createFreeTime)
		protected.PUT("/free-times/:id", apiHandler.updateFreeTime)
		protected.DELETE("/free-times/:id", apiHandler.deleteFreeTime)

		admin := protected.Group("")
		admin.Use(requireRole(1))
		admin.GET("/users", apiHandler.listUsers)
		admin.POST("/users", apiHandler.createUser)
		admin.GET("/users/:student_id", apiHandler.getUser)
		admin.PUT("/users/:student_id", apiHandler.updateUser)
		admin.PATCH("/users/:student_id/status", apiHandler.updateUserStatus)
		admin.PATCH("/users/:student_id/role", apiHandler.updateUserRole)

		admin.GET("/classes", apiHandler.listClasses)
		admin.POST("/classes", apiHandler.createClass)
		admin.GET("/classes/:id", apiHandler.getClass)
		admin.PUT("/classes/:id", apiHandler.updateClass)
		admin.DELETE("/classes/:id", apiHandler.deleteClass)
		admin.GET("/classes/:id/students", apiHandler.getClassStudents)

		admin.GET("/courses", apiHandler.listCourses)
		admin.POST("/courses", apiHandler.createCourse)
		admin.GET("/courses/:id", apiHandler.getCourse)
		admin.PUT("/courses/:id", apiHandler.updateCourse)
		admin.DELETE("/courses/:id", apiHandler.deleteCourse)
		admin.PUT("/courses/:id/students", apiHandler.replaceCourseStudents)
		admin.PUT("/courses/:id/classes", apiHandler.replaceCourseClasses)
		admin.PUT("/courses/:id/sessions", apiHandler.replaceCourseSessions)

		admin.GET("/admin/course-calendar", apiHandler.adminCourseCalendar)
		admin.GET("/admin/attendance-dashboard", apiHandler.adminAttendanceDashboard)
		admin.GET("/admin/attendance-results", apiHandler.adminAttendanceResults)
		admin.GET("/admin/free-time-calendar", apiHandler.adminFreeTimeCalendar)
		admin.GET("/admin/attendance-checks/:id", apiHandler.adminGetAttendanceCheck)
		admin.PATCH("/admin/attendance-details/:id/status", apiHandler.adminUpdateAttendanceStatus)
		admin.GET("/admin/attendance-details/:id/logs", apiHandler.adminAttendanceDetailLogs)
		admin.GET("/admin/operation-logs", apiHandler.listAdminOperationLogs)
		admin.GET("/admin/attendance-detail-logs", apiHandler.listAttendanceDetailLogs)

		student := protected.Group("")
		student.Use(requireRole(2))
		student.GET("/student/courses/available", apiHandler.studentAvailableCourses)
		student.POST("/student/attendance-checks", apiHandler.studentEnterAttendanceCheck)
		student.GET("/student/attendance-checks/:id", apiHandler.studentGetAttendanceCheck)
		student.PATCH("/student/attendance-details/:id/status", apiHandler.studentUpdateAttendanceStatus)
		student.POST("/student/attendance-checks/:id/complete", apiHandler.studentCompleteAttendanceCheck)
	}

	router.GET("/healthz", func(c *gin.Context) {
		ok(c, gin.H{
			"time": time.Now().Format("2006-01-02 15:04:05"),
		})
	})

	return router, nil
}
