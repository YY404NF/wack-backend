package dto

type EnterAttendanceSessionRequest struct {
	CourseGroupLessonID uint64 `json:"course_group_lesson_id" binding:"required"`
}

type UpdateAttendanceStatusRequest struct {
	Status int `json:"status" binding:"required"`
}

type SubmitAttendanceStatusItem struct {
	StudentRefID uint64 `json:"student_ref_id" binding:"required"`
	Status       int    `json:"status" binding:"required"`
}

type SubmitAttendanceStatusesRequest struct {
	Items []SubmitAttendanceStatusItem `json:"items" binding:"required"`
}

type CompleteAttendanceSessionRequest struct {
	SubmittedByUserID uint64 `json:"submitted_by_user_id"`
}
