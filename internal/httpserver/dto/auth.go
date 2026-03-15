package dto

type LoginRequest struct {
	StudentID string `json:"student_id" binding:"required"`
	Password  string `json:"password" binding:"required"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type UpdateProfileRequest struct {
	StudentID string `json:"student_id" binding:"required"`
	RealName  string `json:"real_name" binding:"required"`
}

type InitializeSystemRequest struct {
	StudentID string `json:"student_id" binding:"required"`
	RealName  string `json:"real_name" binding:"required"`
	Password  string `json:"password" binding:"required,min=6"`
}
