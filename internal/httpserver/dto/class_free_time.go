package dto

type FreeTimeRequest struct {
	Term      string `json:"term" binding:"required"`
	StudentID string `json:"student_id"`
	Weekday   int    `json:"weekday" binding:"required"`
	Section   int    `json:"section" binding:"required"`
	FreeWeeks string `json:"free_weeks" binding:"required"`
}
