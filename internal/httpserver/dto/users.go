package dto

type CreateUserRequest struct {
	LoginID        string  `json:"login_id" binding:"required,max=32"`
	RealName       string  `json:"real_name" binding:"required,max=50"`
	Password       string  `json:"password" binding:"required,min=6"`
	Role           int     `json:"role" binding:"required,oneof=1 2 3"`
	Status         int     `json:"status" binding:"omitempty,oneof=1 2"`
	ManagedClassID *uint64 `json:"managed_class_id"`
}

type UpdateUserRequest struct {
	LoginID        string  `json:"login_id" binding:"required,max=32"`
	RealName       string  `json:"real_name" binding:"required,max=50"`
	Role           int     `json:"role" binding:"required,oneof=1 2 3"`
	Status         int     `json:"status" binding:"required,oneof=1 2"`
	ManagedClassID *uint64 `json:"managed_class_id"`
}

type ResetUserPasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type UpdateUserStatusRequest struct {
	Status int `json:"status" binding:"required,oneof=1 2"`
}

type UpdateUserRoleRequest struct {
	Role int `json:"role" binding:"required,oneof=1 2 3"`
}
