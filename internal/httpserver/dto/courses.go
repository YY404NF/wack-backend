package dto

import "wack-backend/internal/model"

type ReplaceCourseStudentsRequest struct {
	StudentIDs []string `json:"student_ids"`
}

type ReplaceCourseClassesRequest struct {
	ClassIDs []uint64 `json:"class_ids"`
}

type ReplaceCourseSessionsRequest struct {
	Sessions []model.CourseSession `json:"sessions"`
}
