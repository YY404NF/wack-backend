package dto

type UpdateSystemSettingRequest struct {
	CurrentTermStartDate string `json:"current_term_start_date" binding:"required"`
}
