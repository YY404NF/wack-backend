package service

import (
	"testing"

	"wack-backend/internal/model"
)

func TestReplaceCourseStudentsUpdatesCount(t *testing.T) {
	db := openTestDB(t)
	svc := NewCourseService(db)

	course := model.Course{
		Term:                   "2025-2026-1",
		CourseName:             "Software Engineering",
		TeacherName:            "Teacher A",
		AttendanceStudentCount: 0,
	}
	if err := db.Create(&course).Error; err != nil {
		t.Fatalf("create course: %v", err)
	}

	students := []model.CourseStudent{
		{StudentID: "20210001", RealName: "Alice"},
		{StudentID: "20210002", RealName: "Bob"},
	}
	if err := svc.ReplaceCourseStudents(course.ID, students); err != nil {
		t.Fatalf("replace course students: %v", err)
	}

	var relations []model.CourseStudent
	if err := db.Where("course_id = ?", course.ID).Find(&relations).Error; err != nil {
		t.Fatalf("load course students: %v", err)
	}
	if len(relations) != 2 {
		t.Fatalf("expected 2 course students, got %d", len(relations))
	}

	var updated model.Course
	if err := db.First(&updated, course.ID).Error; err != nil {
		t.Fatalf("reload course: %v", err)
	}
	if updated.AttendanceStudentCount != 2 {
		t.Fatalf("expected attendance student count 2, got %d", updated.AttendanceStudentCount)
	}
}

func TestDeleteCourseCascadesAttendanceArtifacts(t *testing.T) {
	db := openTestDB(t)
	svc := NewCourseService(db)

	course := model.Course{
		Term:                   "2025-2026-1",
		CourseName:             "Distributed Systems",
		TeacherName:            "Teacher B",
		AttendanceStudentCount: 1,
	}
	if err := db.Create(&course).Error; err != nil {
		t.Fatalf("create course: %v", err)
	}

	courseClass := model.CourseClass{CourseID: course.ID, ClassID: 1}
	courseStudent := model.CourseStudent{CourseID: course.ID, StudentID: "20210001", RealName: "Alice"}
	session := model.CourseSession{
		CourseID:     course.ID,
		SessionNo:    1,
		WeekNo:       1,
		Weekday:      1,
		Section:      1,
		BuildingName: "A",
		RoomName:     "101",
	}
	if err := db.Create(&courseClass).Error; err != nil {
		t.Fatalf("create course class: %v", err)
	}
	if err := db.Create(&courseStudent).Error; err != nil {
		t.Fatalf("create course student: %v", err)
	}
	if err := db.Create(&session).Error; err != nil {
		t.Fatalf("create course session: %v", err)
	}

	check := model.AttendanceCheck{CourseSessionID: session.ID, StartedByUserID: 1}
	if err := db.Create(&check).Error; err != nil {
		t.Fatalf("create attendance check: %v", err)
	}
	detail := model.AttendanceDetail{
		AttendanceCheckID: check.ID,
		StudentID:         "20210001",
		RealName:          "Alice",
		Status:            model.AttendancePresent,
	}
	if err := db.Create(&detail).Error; err != nil {
		t.Fatalf("create attendance detail: %v", err)
	}
	detailLog := model.AttendanceDetailLog{
		AttendanceDetailID: detail.ID,
		AttendanceCheckID:  check.ID,
		StudentID:          detail.StudentID,
		RealName:           detail.RealName,
		OperatorUserID:     1,
		NewStatus:          model.AttendancePresent,
		OperationType:      "set_status",
	}
	if err := db.Create(&detailLog).Error; err != nil {
		t.Fatalf("create attendance detail log: %v", err)
	}

	if err := svc.DeleteCourse(course.ID); err != nil {
		t.Fatalf("delete course: %v", err)
	}

	for name, table := range map[string]any{
		"course":                 &model.Course{},
		"course class":           &model.CourseClass{},
		"course student":         &model.CourseStudent{},
		"course session":         &model.CourseSession{},
		"attendance check":       &model.AttendanceCheck{},
		"attendance detail":      &model.AttendanceDetail{},
		"attendance detail log":  &model.AttendanceDetailLog{},
	} {
		var count int64
		if err := db.Model(table).Count(&count).Error; err != nil {
			t.Fatalf("count %s: %v", name, err)
		}
		if count != 0 {
			t.Fatalf("expected %s count 0, got %d", name, count)
		}
	}
}
