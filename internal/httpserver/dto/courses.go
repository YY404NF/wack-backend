package dto

type CourseGroupClassesRequest struct {
	ClassIDs []uint64 `json:"class_ids"`
}

type CourseGroupStudentsRequest struct {
	StudentIDs []uint64 `json:"student_ids"`
}

type CourseGroupLessonRequest struct {
	WeekNo       int    `json:"week_no"`
	Weekday      int    `json:"weekday"`
	Section      int    `json:"section"`
	BuildingName string `json:"building_name"`
	RoomName     string `json:"room_name"`
}
