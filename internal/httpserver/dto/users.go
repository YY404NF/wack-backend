package dto

type CreateUserRequest struct {
	StudentID string `json:"student_id" binding:"required"`
	RealName  string `json:"real_name" binding:"required"`
	Password  string `json:"password" binding:"required,min=6"`
	Role      int    `json:"role" binding:"required"`
	Status    int    `json:"status"`
}

type UpdateUserRequest struct {
	StudentID string `json:"student_id" binding:"required"`
	RealName  string `json:"real_name" binding:"required"`
	Role      int    `json:"role" binding:"required"`
	Status    int    `json:"status" binding:"required"`
}

type ResetUserPasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type UpdateUserStatusRequest struct {
	Status int `json:"status" binding:"required"`
}

type UpdateUserRoleRequest struct {
	Role int `json:"role" binding:"required"`
}
