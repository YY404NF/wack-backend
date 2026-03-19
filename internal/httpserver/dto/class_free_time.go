package dto

type FreeTimeRequest struct {
	Term      string `json:"term" binding:"required,max=20"`
	LoginID   string `json:"login_id"`
	Weekday   int    `json:"weekday" binding:"required,oneof=1 2 3 4 5 6 7"`
	Section   int    `json:"section" binding:"required,oneof=1 2 3 4 5"`
	FreeWeeks string `json:"free_weeks" binding:"required,max=100"`
}
