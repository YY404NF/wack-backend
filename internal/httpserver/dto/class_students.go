package dto

type ClassStudentRequest struct {
	StudentID string `json:"student_id" binding:"required"`
	RealName  string `json:"real_name" binding:"required"`
}

type ImportClassStudentsRequest []ClassStudentRequest
