package service

import (
	"path/filepath"
	"testing"
	"time"

	"wack-backend/internal/database"
	"wack-backend/internal/model"
)

func setupAttendanceServiceTest(t *testing.T) (*AttendanceService, *model.CourseGroupLesson, *model.CourseGroupStudent, *model.CourseGroupStudent) {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "wack.db")
	db, err := database.OpenAndMigrate(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	term := model.Term{Name: "2025-2026-2", TermStartDate: "2026-03-02"}
	if err := db.Create(&term).Error; err != nil {
		t.Fatalf("create term: %v", err)
	}

	course := model.Course{TermID: term.ID, Grade: 2025, CourseName: "测试课程", TeacherName: "教师", Status: 1}
	if err := db.Create(&course).Error; err != nil {
		t.Fatalf("create course: %v", err)
	}

	group := model.CourseGroup{TermID: term.ID, CourseID: course.ID, Status: 1}
	if err := db.Create(&group).Error; err != nil {
		t.Fatalf("create group: %v", err)
	}

	class := model.Class{Grade: 2025, MajorName: "密码科学与技术", ClassName: "密码科学与技术2501班", Status: 1}
	if err := db.Create(&class).Error; err != nil {
		t.Fatalf("create class: %v", err)
	}

	student1 := model.Student{StudentNo: "20250001", StudentName: "甲", ClassID: &class.ID, Status: 1}
	student2 := model.Student{StudentNo: "20250002", StudentName: "乙", ClassID: &class.ID, Status: 1}
	if err := db.Create(&student1).Error; err != nil {
		t.Fatalf("create student1: %v", err)
	}
	if err := db.Create(&student2).Error; err != nil {
		t.Fatalf("create student2: %v", err)
	}

	relation1 := model.CourseGroupStudent{TermID: term.ID, CourseGroupID: group.ID, StudentID: student1.ID, ClassID: &class.ID, Status: 1}
	relation2 := model.CourseGroupStudent{TermID: term.ID, CourseGroupID: group.ID, StudentID: student2.ID, ClassID: &class.ID, Status: 1}
	if err := db.Create(&relation1).Error; err != nil {
		t.Fatalf("create relation1: %v", err)
	}
	if err := db.Create(&relation2).Error; err != nil {
		t.Fatalf("create relation2: %v", err)
	}

	lesson := model.CourseGroupLesson{
		TermID:        term.ID,
		CourseGroupID: group.ID,
		WeekNo:        3,
		Weekday:       5,
		Section:       3,
		BuildingName:  "教4",
		RoomName:      "502",
		Status:        1,
	}
	if err := db.Create(&lesson).Error; err != nil {
		t.Fatalf("create lesson: %v", err)
	}

	return NewAttendanceService(db), &lesson, &relation1, &relation2
}

func TestSubmitAttendanceStatusesForClassIgnoresExistingRecordsWithoutOverwrite(t *testing.T) {
	service, lesson, relation1, relation2 := setupAttendanceServiceTest(t)

	if err := service.UpsertAttendanceStatusForStudent(lesson.ID, relation1.StudentID, model.AttendanceLate, 99, true); err != nil {
		t.Fatalf("seed attendance record: %v", err)
	}

	result, err := service.SubmitAttendanceStatusesForClass(lesson.ID, 100, []AttendanceStatusInput{
		{StudentRefID: relation1.StudentID, Status: model.AttendanceAbsent},
		{StudentRefID: relation2.StudentID, Status: model.AttendancePresent},
	}, nil, func(model.CourseGroupLesson, time.Time) bool { return true })
	if err != nil {
		t.Fatalf("submit statuses: %v", err)
	}

	if result.AppliedCount != 1 || len(result.AcceptedItems) != 1 || result.AcceptedItems[0] != relation2.StudentID {
		t.Fatalf("unexpected accepted result: %+v", result)
	}
	if result.IgnoredCount != 1 || len(result.IgnoredItems) != 1 || result.IgnoredItems[0] != relation1.StudentID {
		t.Fatalf("unexpected ignored result: %+v", result)
	}

	records, err := service.AttendanceRecords(lesson.ID)
	if err != nil {
		t.Fatalf("load records: %v", err)
	}

	statusByStudent := map[string]int{}
	for _, record := range records {
		if record.Status != nil {
			statusByStudent[record.StudentID] = *record.Status
		}
	}
	if statusByStudent["20250001"] != model.AttendanceLate {
		t.Fatalf("existing record was overwritten: %+v", statusByStudent)
	}
	if statusByStudent["20250002"] != model.AttendancePresent {
		t.Fatalf("new record was not created: %+v", statusByStudent)
	}
}

func TestAbandonAttendanceSessionDoesNotDeleteSubmittedResults(t *testing.T) {
	service, lesson, relation1, _ := setupAttendanceServiceTest(t)

	if err := service.UpsertAttendanceStatusForStudent(lesson.ID, relation1.StudentID, model.AttendanceLate, 99, true); err != nil {
		t.Fatalf("seed attendance record: %v", err)
	}

	if err := service.AbandonAttendanceSession(lesson.ID, func(model.CourseGroupLesson, time.Time) bool { return true }); err != nil {
		t.Fatalf("abandon attendance session: %v", err)
	}

	records, err := service.AttendanceRecords(lesson.ID)
	if err != nil {
		t.Fatalf("load records: %v", err)
	}

	recordCount := 0
	for _, record := range records {
		if record.AttendanceRecordID != nil {
			recordCount++
		}
	}
	if recordCount != 1 {
		t.Fatalf("expected submitted attendance record to remain, got %d", recordCount)
	}
}

func TestAvailableCourseGroupLessonsOnlyReturnsCurrentTermSessions(t *testing.T) {
	service, lesson, _, _ := setupAttendanceServiceTest(t)

	currentSessions, err := service.AvailableCourseGroupLessons(lesson.TermID, lesson.Weekday, lesson.WeekNo)
	if err != nil {
		t.Fatalf("load current term sessions: %v", err)
	}
	if len(currentSessions) != 1 || currentSessions[0].ID != lesson.ID {
		t.Fatalf("unexpected current term sessions: %+v", currentSessions)
	}

	otherTerm := model.Term{Name: "2025-2026-1", TermStartDate: "2025-09-01"}
	if err := service.db.Create(&otherTerm).Error; err != nil {
		t.Fatalf("create other term: %v", err)
	}
	otherCourse := model.Course{TermID: otherTerm.ID, Grade: 2025, CourseName: "其他学期课程", TeacherName: "教师", Status: 1}
	if err := service.db.Create(&otherCourse).Error; err != nil {
		t.Fatalf("create other course: %v", err)
	}
	otherGroup := model.CourseGroup{TermID: otherTerm.ID, CourseID: otherCourse.ID, Status: 1}
	if err := service.db.Create(&otherGroup).Error; err != nil {
		t.Fatalf("create other group: %v", err)
	}
	otherLesson := model.CourseGroupLesson{
		TermID:        otherTerm.ID,
		CourseGroupID: otherGroup.ID,
		WeekNo:        lesson.WeekNo,
		Weekday:       lesson.Weekday,
		Section:       lesson.Section,
		BuildingName:  "教3",
		RoomName:      "301",
		Status:        1,
	}
	if err := service.db.Create(&otherLesson).Error; err != nil {
		t.Fatalf("create other lesson: %v", err)
	}

	currentSessions, err = service.AvailableCourseGroupLessons(lesson.TermID, lesson.Weekday, lesson.WeekNo)
	if err != nil {
		t.Fatalf("reload current term sessions: %v", err)
	}
	if len(currentSessions) != 1 || currentSessions[0].ID != lesson.ID {
		t.Fatalf("expected only current term lesson, got %+v", currentSessions)
	}

	otherSessions, err := service.AvailableCourseGroupLessons(otherTerm.ID, lesson.Weekday, lesson.WeekNo)
	if err != nil {
		t.Fatalf("load other term sessions: %v", err)
	}
	if len(otherSessions) != 1 || otherSessions[0].ID != otherLesson.ID {
		t.Fatalf("expected only other term lesson, got %+v", otherSessions)
	}
}
