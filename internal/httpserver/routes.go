package httpserver

import "github.com/gin-gonic/gin"

func mountSetupRoutes(api *gin.RouterGroup, authHandler *authHandler) {
	setupGroup := api.Group("/setup")
	setupGroup.GET("/status", authHandler.setupStatus)
	setupGroup.POST("/initialize", authHandler.initializeSystem)
}

func mountAuthRoutes(api *gin.RouterGroup, cfgMiddle gin.HandlerFunc, authHandler *authHandler) {
	authGroup := api.Group("/auth")
	authGroup.POST("/login", authHandler.login)

	protectedAuth := api.Group("")
	protectedAuth.Use(cfgMiddle)
	protectedAuth.GET("/auth/me", authHandler.me)
	protectedAuth.POST("/auth/change-password", authHandler.changePassword)
	protectedAuth.PUT("/auth/profile", authHandler.updateProfile)
}

func mountFreeTimeRoutes(protected *gin.RouterGroup, apiHandler *apiHandler) {
	protected.GET("/free-times", apiHandler.listFreeTimes)
	protected.POST("/free-times", apiHandler.createFreeTime)
	protected.PUT("/free-times/:id", apiHandler.updateFreeTime)
	protected.DELETE("/free-times/:id", apiHandler.deleteFreeTime)
	protected.GET("/system-settings", apiHandler.getSystemSetting)
}

func mountUserRoutes(admin *gin.RouterGroup, apiHandler *apiHandler) {
	admin.GET("/users", apiHandler.listUsers)
	admin.POST("/users", apiHandler.createUser)
	admin.GET("/users/:student_id", apiHandler.getUser)
	admin.PUT("/users/:student_id", apiHandler.updateUser)
	admin.PATCH("/users/:student_id/password", apiHandler.resetUserPassword)
	admin.PATCH("/users/:student_id/status", apiHandler.updateUserStatus)
	admin.PATCH("/users/:student_id/role", apiHandler.updateUserRole)
}

func mountClassRoutes(admin *gin.RouterGroup, apiHandler *apiHandler) {
	admin.GET("/classes", apiHandler.listClasses)
	admin.GET("/class-students", apiHandler.listClassStudentCandidates)
	admin.POST("/classes", apiHandler.createClass)
	admin.GET("/classes/:id", apiHandler.getClass)
	admin.PUT("/classes/:id", apiHandler.updateClass)
	admin.DELETE("/classes/:id", apiHandler.deleteClass)
	admin.GET("/classes/:id/students", apiHandler.getClassStudents)
	admin.POST("/classes/:id/students", apiHandler.createClassStudent)
	admin.POST("/classes/:id/students/import", apiHandler.importClassStudents)
	admin.PUT("/classes/:id/students/:student_id", apiHandler.updateClassStudent)
	admin.DELETE("/classes/:id/students/:student_id", apiHandler.deleteClassStudent)
}

func mountCourseRoutes(admin *gin.RouterGroup, apiHandler *apiHandler) {
	admin.GET("/courses", apiHandler.listCourses)
	admin.POST("/courses", apiHandler.createCourse)
	admin.GET("/courses/:id", apiHandler.getCourse)
	admin.PUT("/courses/:id", apiHandler.updateCourse)
	admin.DELETE("/courses/:id", apiHandler.deleteCourse)
	admin.PUT("/courses/:id/students", apiHandler.replaceCourseStudents)
	admin.PUT("/courses/:id/classes", apiHandler.replaceCourseClasses)
	admin.PUT("/courses/:id/sessions", apiHandler.replaceCourseSessions)
	admin.GET("/course-calendar", apiHandler.adminCourseCalendar)
}

func mountSystemSettingRoutes(admin *gin.RouterGroup, apiHandler *apiHandler) {
	admin.GET("/system-settings", apiHandler.getSystemSetting)
	admin.PUT("/system-settings", apiHandler.updateSystemSetting)
}

func mountAttendanceRoutes(admin, student *gin.RouterGroup, apiHandler *apiHandler) {
	admin.GET("/attendance-dashboard", apiHandler.adminAttendanceDashboard)
	admin.GET("/attendance-results", apiHandler.adminAttendanceResults)
	admin.GET("/free-time-calendar", apiHandler.adminFreeTimeCalendar)
	admin.GET("/attendance-checks/:id", apiHandler.adminGetAttendanceCheck)
	admin.PATCH("/attendance-details/:id/status", apiHandler.adminUpdateAttendanceStatus)
	admin.GET("/attendance-details/:id/logs", apiHandler.adminAttendanceDetailLogs)
	admin.GET("/operation-logs", apiHandler.listAdminOperationLogs)
	admin.GET("/attendance-detail-logs", apiHandler.listAttendanceDetailLogs)

	student.GET("/courses/available", apiHandler.studentAvailableCourses)
	student.POST("/attendance-checks", apiHandler.studentEnterAttendanceCheck)
	student.GET("/attendance-checks/:id", apiHandler.studentGetAttendanceCheck)
	student.PATCH("/attendance-details/:id/status", apiHandler.studentUpdateAttendanceStatus)
	student.POST("/attendance-checks/:id/complete", apiHandler.studentCompleteAttendanceCheck)
}
