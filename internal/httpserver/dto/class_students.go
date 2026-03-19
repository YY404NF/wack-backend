package dto

type ClassStudentRequest struct {
	StudentID string `json:"student_id" binding:"required,max=32"`
	RealName  string `json:"real_name" binding:"required,max=50"`
}

type ClassStudentImportResponse struct {
	ImportedCount int `json:"imported_count"`
}
