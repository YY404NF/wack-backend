package dto

type StudentRequest struct {
	StudentID string `json:"student_id" binding:"required,max=32"`
	RealName  string `json:"real_name" binding:"required,max=50"`
	ClassID   *uint64 `json:"class_id"`
}
