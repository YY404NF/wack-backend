package dto

import "wack-backend/internal/model"

type ReplaceCourseStudentsRequest struct {
	Students []struct {
		StudentID string `json:"student_id"`
		RealName  string `json:"real_name"`
	} `json:"students"`
}

type ReplaceCourseClassesRequest struct {
	ClassIDs []uint64 `json:"class_ids"`
}

type ReplaceCourseSessionsRequest struct {
	Sessions []model.CourseSession `json:"sessions"`
}
