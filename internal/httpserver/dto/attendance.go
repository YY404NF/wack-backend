package dto

type EnterAttendanceCheckRequest struct {
	CourseSessionID uint64 `json:"course_session_id" binding:"required"`
}

type UpdateAttendanceStatusRequest struct {
	Status int `json:"status" binding:"required"`
}
