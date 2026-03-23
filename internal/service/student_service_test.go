package service

import (
	"path/filepath"
	"testing"
	"time"

	"gorm.io/gorm"
	"wack-backend/internal/database"
	"wack-backend/internal/model"
)

func setupStudentServiceTest(t *testing.T) (*StudentService, *gorm.DB, model.Term, model.CourseGroup, model.CourseGroupLesson, model.Class, model.Student, model.CourseGroupStudent) {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "wack.db")
	db, err := database.OpenAndMigrate(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	term := model.Term{Name: "2025-2026-2", TermStartDate: activeTermStartDate()}
	if err := db.Create(&term).Error; err != nil {
		t.Fatalf("create term: %v", err)
	}

	class := model.Class{Grade: 2025, MajorName: "密码科学与技术", ClassName: "密码科学与技术2501班", Status: 1}
	if err := db.Create(&class).Error; err != nil {
		t.Fatalf("create class: %v", err)
	}

	student := model.Student{StudentNo: "20250001", StudentName: "甲", ClassID: &class.ID, Status: 1}
	if err := db.Create(&student).Error; err != nil {
		t.Fatalf("create student: %v", err)
	}

	course := model.Course{TermID: term.ID, Grade: 2025, CourseName: "测试课程", TeacherName: "教师", Status: 1}
	if err := db.Create(&course).Error; err != nil {
		t.Fatalf("create course: %v", err)
	}

	group := model.CourseGroup{TermID: term.ID, CourseID: course.ID, Status: 1}
	if err := db.Create(&group).Error; err != nil {
		t.Fatalf("create course group: %v", err)
	}

	relation := model.CourseGroupStudent{TermID: term.ID, CourseGroupID: group.ID, StudentID: student.ID, ClassID: &class.ID, Status: 1}
	if err := db.Create(&relation).Error; err != nil {
		t.Fatalf("create course group student: %v", err)
	}

	lesson := model.CourseGroupLesson{
		TermID:        term.ID,
		CourseGroupID: group.ID,
		WeekNo:        3,
		Weekday:       3,
		Section:       1,
		BuildingName:  "教4",
		RoomName:      "509",
		Status:        1,
	}
	if err := db.Create(&lesson).Error; err != nil {
		t.Fatalf("create lesson: %v", err)
	}

	return NewStudentService(db), db, term, group, lesson, class, student, relation
}

func activeTermStartDate() string {
	now := time.Now()
	start := now.AddDate(0, 0, -7)
	return start.Format("2006-01-02")
}

func TestDeleteStudentRemovesCourseGroupStudentWhenNoAttendanceHistory(t *testing.T) {
	service, db, _, _, _, _, student, _ := setupStudentServiceTest(t)

	if err := service.DeleteStudent(student.ID); err != nil {
		t.Fatalf("delete student: %v", err)
	}

	var studentCount int64
	if err := db.Model(&model.Student{}).Where("id = ?", student.ID).Count(&studentCount).Error; err != nil {
		t.Fatalf("count student: %v", err)
	}
	if studentCount != 0 {
		t.Fatalf("expected student to be physically deleted, got count %d", studentCount)
	}

	var relationCount int64
	if err := db.Model(&model.CourseGroupStudent{}).Where("student_id = ?", student.ID).Count(&relationCount).Error; err != nil {
		t.Fatalf("count relation: %v", err)
	}
	if relationCount != 0 {
		t.Fatalf("expected course_group_student to be deleted, got count %d", relationCount)
	}
}

func TestDeleteStudentMarksCourseGroupStudentDeletedWhenAttendanceHistoryExists(t *testing.T) {
	service, db, term, group, lesson, class, student, _ := setupStudentServiceTest(t)

	record := model.AttendanceRecord{
		TermID:              term.ID,
		CourseID:            group.CourseID,
		CourseGroupLessonID: lesson.ID,
		StudentID:           student.ID,
		ClassID:             &class.ID,
		AttendanceStatus:    model.AttendanceLate,
		UpdatedByUserID:     nil,
	}
	if err := db.Create(&record).Error; err != nil {
		t.Fatalf("create attendance record: %v", err)
	}

	if err := service.DeleteStudent(student.ID); err != nil {
		t.Fatalf("delete student: %v", err)
	}

	var savedStudent model.Student
	if err := db.First(&savedStudent, "id = ?", student.ID).Error; err != nil {
		t.Fatalf("reload student: %v", err)
	}
	if savedStudent.Status != 2 {
		t.Fatalf("expected student status 2, got %d", savedStudent.Status)
	}

	var relations []model.CourseGroupStudent
	if err := db.Where("student_id = ?", student.ID).Find(&relations).Error; err != nil {
		t.Fatalf("reload relations: %v", err)
	}
	if len(relations) != 1 {
		t.Fatalf("expected 1 relation, got %d", len(relations))
	}
	if relations[0].Status != 2 {
		t.Fatalf("expected relation status 2, got %d", relations[0].Status)
	}
}

func TestUpdateStudentAddsStudentToCurrentTermClassCourseGroups(t *testing.T) {
	service, db, term, _, _, class, student, _ := setupStudentServiceTest(t)

	if err := db.Model(&model.Student{}).Where("id = ?", student.ID).Update("class_id", nil).Error; err != nil {
		t.Fatalf("unbind seeded student: %v", err)
	}
	if err := db.Where("student_id = ?", student.ID).Delete(&model.CourseGroupStudent{}).Error; err != nil {
		t.Fatalf("delete seeded relation: %v", err)
	}

	peer := model.Student{StudentNo: "20250002", StudentName: "乙", ClassID: &class.ID, Status: 1}
	if err := db.Create(&peer).Error; err != nil {
		t.Fatalf("create peer student: %v", err)
	}

	activeCourse := model.Course{TermID: term.ID, Grade: 2025, CourseName: "当前学期课程", TeacherName: "教师A", Status: 1}
	if err := db.Create(&activeCourse).Error; err != nil {
		t.Fatalf("create active course: %v", err)
	}
	activeGroup := model.CourseGroup{TermID: term.ID, CourseID: activeCourse.ID, Status: 1}
	if err := db.Create(&activeGroup).Error; err != nil {
		t.Fatalf("create active group: %v", err)
	}
	if err := db.Create(&model.CourseGroupStudent{
		TermID:        term.ID,
		CourseGroupID: activeGroup.ID,
		StudentID:     peer.ID,
		ClassID:       &class.ID,
		Status:        1,
	}).Error; err != nil {
		t.Fatalf("seed active class relation: %v", err)
	}

	futureTerm := model.Term{Name: "2026-2027-1", TermStartDate: time.Now().AddDate(1, 0, 0).Format("2006-01-02")}
	if err := db.Create(&futureTerm).Error; err != nil {
		t.Fatalf("create future term: %v", err)
	}
	futureCourse := model.Course{TermID: futureTerm.ID, Grade: 2025, CourseName: "未来学期课程", TeacherName: "教师B", Status: 1}
	if err := db.Create(&futureCourse).Error; err != nil {
		t.Fatalf("create future course: %v", err)
	}
	futureGroup := model.CourseGroup{TermID: futureTerm.ID, CourseID: futureCourse.ID, Status: 1}
	if err := db.Create(&futureGroup).Error; err != nil {
		t.Fatalf("create future group: %v", err)
	}
	if err := db.Create(&model.CourseGroupStudent{
		TermID:        futureTerm.ID,
		CourseGroupID: futureGroup.ID,
		StudentID:     peer.ID,
		ClassID:       &class.ID,
		Status:        1,
	}).Error; err != nil {
		t.Fatalf("seed future class relation: %v", err)
	}

	if _, err := service.UpdateStudent(student.ID, model.Student{
		StudentNo:   student.StudentNo,
		StudentName: student.StudentName,
		ClassID:     &class.ID,
	}); err != nil {
		t.Fatalf("update student class binding: %v", err)
	}

	var activeRelationCount int64
	if err := db.Model(&model.CourseGroupStudent{}).
		Where("term_id = ? AND course_group_id = ? AND student_id = ? AND class_id = ? AND status = 1", term.ID, activeGroup.ID, student.ID, class.ID).
		Count(&activeRelationCount).Error; err != nil {
		t.Fatalf("count active relation: %v", err)
	}
	if activeRelationCount != 1 {
		t.Fatalf("expected current term relation to be added, got %d", activeRelationCount)
	}

	var futureRelationCount int64
	if err := db.Model(&model.CourseGroupStudent{}).
		Where("term_id = ? AND course_group_id = ? AND student_id = ?", futureTerm.ID, futureGroup.ID, student.ID).
		Count(&futureRelationCount).Error; err != nil {
		t.Fatalf("count future relation: %v", err)
	}
	if futureRelationCount != 0 {
		t.Fatalf("expected future term relation to stay untouched, got %d", futureRelationCount)
	}
}

func TestUpdateStudentRemovesOnlyCurrentTermClassCourseGroupsOnUnbind(t *testing.T) {
	service, db, term, _, lesson, class, student, _ := setupStudentServiceTest(t)

	currentHistoryCourse := model.Course{TermID: term.ID, Grade: 2025, CourseName: "有历史课程", TeacherName: "教师A", Status: 1}
	if err := db.Create(&currentHistoryCourse).Error; err != nil {
		t.Fatalf("create current history course: %v", err)
	}
	currentHistoryGroup := model.CourseGroup{TermID: term.ID, CourseID: currentHistoryCourse.ID, Status: 1}
	if err := db.Create(&currentHistoryGroup).Error; err != nil {
		t.Fatalf("create current history group: %v", err)
	}
	historyRelation := model.CourseGroupStudent{TermID: term.ID, CourseGroupID: currentHistoryGroup.ID, StudentID: student.ID, ClassID: &class.ID, Status: 1}
	if err := db.Create(&historyRelation).Error; err != nil {
		t.Fatalf("create current history relation: %v", err)
	}
	if err := db.Model(&lesson).Updates(map[string]interface{}{"course_group_id": currentHistoryGroup.ID}).Error; err != nil {
		t.Fatalf("rebind lesson group: %v", err)
	}
	record := model.AttendanceRecord{
		TermID:              term.ID,
		CourseID:            currentHistoryCourse.ID,
		CourseGroupLessonID: lesson.ID,
		StudentID:           student.ID,
		ClassID:             &class.ID,
		AttendanceStatus:    model.AttendanceLate,
	}
	if err := db.Create(&record).Error; err != nil {
		t.Fatalf("create attendance record: %v", err)
	}

	currentNoHistoryCourse := model.Course{TermID: term.ID, Grade: 2025, CourseName: "无历史课程", TeacherName: "教师B", Status: 1}
	if err := db.Create(&currentNoHistoryCourse).Error; err != nil {
		t.Fatalf("create current no-history course: %v", err)
	}
	currentNoHistoryGroup := model.CourseGroup{TermID: term.ID, CourseID: currentNoHistoryCourse.ID, Status: 1}
	if err := db.Create(&currentNoHistoryGroup).Error; err != nil {
		t.Fatalf("create current no-history group: %v", err)
	}
	noHistoryRelation := model.CourseGroupStudent{TermID: term.ID, CourseGroupID: currentNoHistoryGroup.ID, StudentID: student.ID, ClassID: &class.ID, Status: 1}
	if err := db.Create(&noHistoryRelation).Error; err != nil {
		t.Fatalf("create current no-history relation: %v", err)
	}

	futureTerm := model.Term{Name: "2026-2027-1", TermStartDate: time.Now().AddDate(1, 0, 0).Format("2006-01-02")}
	if err := db.Create(&futureTerm).Error; err != nil {
		t.Fatalf("create future term: %v", err)
	}
	futureCourse := model.Course{TermID: futureTerm.ID, Grade: 2025, CourseName: "未来课程", TeacherName: "教师C", Status: 1}
	if err := db.Create(&futureCourse).Error; err != nil {
		t.Fatalf("create future course: %v", err)
	}
	futureGroup := model.CourseGroup{TermID: futureTerm.ID, CourseID: futureCourse.ID, Status: 1}
	if err := db.Create(&futureGroup).Error; err != nil {
		t.Fatalf("create future group: %v", err)
	}
	futureRelation := model.CourseGroupStudent{TermID: futureTerm.ID, CourseGroupID: futureGroup.ID, StudentID: student.ID, ClassID: &class.ID, Status: 1}
	if err := db.Create(&futureRelation).Error; err != nil {
		t.Fatalf("create future relation: %v", err)
	}

	if _, err := service.UpdateStudent(student.ID, model.Student{
		StudentNo:   student.StudentNo,
		StudentName: student.StudentName,
		ClassID:     nil,
	}); err != nil {
		t.Fatalf("update student class unbinding: %v", err)
	}

	var savedHistoryRelation model.CourseGroupStudent
	if err := db.First(&savedHistoryRelation, historyRelation.ID).Error; err != nil {
		t.Fatalf("reload history relation: %v", err)
	}
	if savedHistoryRelation.Status != 2 {
		t.Fatalf("expected current term relation with history to be logically deleted, got %d", savedHistoryRelation.Status)
	}

	var noHistoryCount int64
	if err := db.Model(&model.CourseGroupStudent{}).Where("id = ?", noHistoryRelation.ID).Count(&noHistoryCount).Error; err != nil {
		t.Fatalf("count no-history relation: %v", err)
	}
	if noHistoryCount != 0 {
		t.Fatalf("expected current term relation without history to be removed, got %d", noHistoryCount)
	}

	var savedFutureRelation model.CourseGroupStudent
	if err := db.First(&savedFutureRelation, futureRelation.ID).Error; err != nil {
		t.Fatalf("reload future relation: %v", err)
	}
	if savedFutureRelation.Status != 1 {
		t.Fatalf("expected future term relation to remain untouched, got %d", savedFutureRelation.Status)
	}
}
