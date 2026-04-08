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
	authGroup.POST("/logout", authHandler.logout)

	protectedAuth := api.Group("")
	protectedAuth.Use(cfgMiddle)
	protectedAuth.GET("/auth/me", authHandler.me)
	protectedAuth.POST("/auth/change-password", authHandler.changePassword)
	protectedAuth.PUT("/auth/profile", authHandler.updateProfile)
}

func mountFreeTimeRoutes(protected *gin.RouterGroup, apiHandler *apiHandler) {
	protected.GET("/free-times", apiHandler.listFreeTimes)
	protected.GET("/free-time-editor", apiHandler.listFreeTimeEditor)
	protected.POST("/free-times", apiHandler.createFreeTime)
	protected.PUT("/free-times/:id", apiHandler.updateFreeTime)
	protected.DELETE("/free-times/:id", apiHandler.deleteFreeTime)
	protected.GET("/system-settings", apiHandler.getSystemSetting)
}

func mountMetaRoutes(protected *gin.RouterGroup, apiHandler *apiHandler) {
	protected.GET("/meta/context", apiHandler.metaContext)
	protected.GET("/meta/terms", apiHandler.metaTerms)
	protected.GET("/meta/terms/:term_id/weeks", apiHandler.metaTermWeeks)
	protected.GET("/meta/sections", apiHandler.metaSections)
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
	admin.GET("/class-options", apiHandler.listClassOptions)
	admin.GET("/class-students", apiHandler.listClassStudentCandidates)
	admin.POST("/classes", apiHandler.createClass)
	admin.GET("/classes/:id", apiHandler.getClass)
	admin.PUT("/classes/:id", apiHandler.updateClass)
	admin.DELETE("/classes/:id", apiHandler.deleteClass)
	admin.GET("/classes/:id/attendance-records", apiHandler.getClassAttendanceRecords)
	admin.PATCH("/classes/:id/attendance-records/statuses", apiHandler.bulkUpdateClassAttendanceRecordStatuses)
	admin.GET("/classes/:id/students", apiHandler.getClassStudents)
	admin.POST("/classes/:id/students/import", apiHandler.importClassStudents)
	admin.POST("/classes/:id/students", apiHandler.createClassStudent)
	admin.PUT("/classes/:id/students/:student_id", apiHandler.updateClassStudent)
	admin.DELETE("/classes/:id/students/:student_id", apiHandler.deleteClassStudent)
}

func mountStudentRoutes(admin *gin.RouterGroup, apiHandler *apiHandler) {
	admin.GET("/student-options", apiHandler.listStudentOptions)
	admin.GET("/students", apiHandler.listStudents)
	admin.GET("/students/:id/attendance-records", apiHandler.listStudentAttendanceRecords)
	admin.PATCH("/students/:id/attendance-records/statuses", apiHandler.bulkUpdateStudentAttendanceRecordStatuses)
	admin.POST("/students", apiHandler.createStudent)
	admin.PUT("/students/:id", apiHandler.updateStudent)
	admin.DELETE("/students/:id", apiHandler.deleteStudent)
}

func mountCourseRoutes(admin *gin.RouterGroup, apiHandler *apiHandler) {
	admin.GET("/courses", apiHandler.listCourses)
	admin.GET("/courses/:id/summary", apiHandler.getCourseSummary)
	admin.POST("/courses", apiHandler.createCourse)
	admin.GET("/courses/:id/groups", apiHandler.listCourseGroups)
	admin.POST("/courses/:id/groups", apiHandler.createCourseGroup)
	admin.GET("/courses/:id/groups/:group_id", apiHandler.getCourseGroup)
	admin.DELETE("/courses/:id/groups/:group_id", apiHandler.deleteCourseGroup)
	admin.GET("/courses/:id/groups/:group_id/sessions", apiHandler.listCourseGroupSessions)
	admin.POST("/courses/:id/groups/:group_id/sessions", apiHandler.createCourseGroupSession)
	admin.PUT("/courses/:id/groups/:group_id/sessions/:session_id", apiHandler.updateCourseGroupSession)
	admin.DELETE("/courses/:id/groups/:group_id/sessions/:session_id", apiHandler.deleteCourseGroupSession)
	admin.GET("/courses/:id/groups/:group_id/students", apiHandler.listCourseGroupStudents)
	admin.GET("/courses/:id/groups/:group_id/available-classes", apiHandler.listAvailableCourseGroupClasses)
	admin.POST("/courses/:id/groups/:group_id/classes", apiHandler.addCourseGroupClasses)
	admin.DELETE("/courses/:id/groups/:group_id/classes/:class_id", apiHandler.removeCourseGroupClass)
	admin.GET("/courses/:id/groups/:group_id/available-students", apiHandler.listAvailableCourseGroupStudents)
	admin.POST("/courses/:id/groups/:group_id/students", apiHandler.addCourseGroupStudents)
	admin.DELETE("/courses/:id/groups/:group_id/students/:student_id", apiHandler.removeCourseGroupStudent)
	admin.PUT("/courses/:id", apiHandler.updateCourse)
	admin.DELETE("/courses/:id", apiHandler.deleteCourse)
	admin.GET("/course-calendar-outline", apiHandler.adminCourseCalendarOutline)
	admin.GET("/course-calendar", apiHandler.adminCourseCalendar)
}

func mountSystemSettingRoutes(admin *gin.RouterGroup, apiHandler *apiHandler) {
	admin.GET("/system-settings", apiHandler.getSystemSetting)
	admin.PUT("/system-settings", apiHandler.updateSystemSetting)
	admin.POST("/terms", apiHandler.adminCreateTerm)
	admin.PUT("/terms/:term_id", apiHandler.adminUpdateTerm)
}

func mountAttendanceRoutes(admin, student *gin.RouterGroup, apiHandler *apiHandler) {
	admin.GET("/overview", apiHandler.adminOverview)
	admin.GET("/attendance-dashboard", apiHandler.adminAttendanceDashboard)
	admin.GET("/attendance-results", apiHandler.adminAttendanceResults)
	admin.GET("/free-time-calendar", apiHandler.adminFreeTimeCalendar)
	admin.GET("/attendance-sessions", apiHandler.adminAttendanceSessions)
	admin.GET("/attendance-sessions/:id", apiHandler.adminGetAttendanceSession)
	admin.PATCH("/attendance-sessions/:id/students/:student_id/status", apiHandler.adminUpsertAttendanceStatus)
	admin.PATCH("/attendance-sessions/:id/students/statuses", apiHandler.adminBulkUpsertAttendanceStatuses)
	admin.PATCH("/attendance-records/:id/status", apiHandler.adminUpdateAttendanceStatus)
	admin.GET("/attendance-records/:id/logs", apiHandler.adminAttendanceRecordLogs)
	admin.GET("/attendance-record-logs", apiHandler.listAttendanceRecordLogs)

	student.GET("/courses/available", apiHandler.studentAvailableCourses)
	student.GET("/managed-class", apiHandler.studentManagedClass)
	student.POST("/attendance-sessions", apiHandler.studentEnterAttendanceSession)
	student.GET("/attendance-sessions/:id", apiHandler.studentGetAttendanceSession)
	student.PATCH("/attendance-records/:id/status", apiHandler.studentUpdateAttendanceStatus)
	student.POST("/attendance-sessions/:id/submit", apiHandler.studentSubmitAttendanceStatuses)
	student.POST("/attendance-sessions/:id/complete", apiHandler.studentCompleteAttendanceSession)
	student.DELETE("/attendance-sessions/:id", apiHandler.studentAbandonAttendanceSession)
}
