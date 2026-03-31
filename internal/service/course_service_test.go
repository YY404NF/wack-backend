package service

import (
	"path/filepath"
	"testing"

	"gorm.io/gorm"

	"wack-backend/internal/database"
	"wack-backend/internal/model"
)

func setupCourseServiceTest(t *testing.T) (*CourseService, *gorm.DB) {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "wack.db")
	db, err := database.OpenAndMigrate(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	return NewCourseService(db), db
}

func TestAddCourseGroupStudentsUsesExistingClassIdentity(t *testing.T) {
	service, db := setupCourseServiceTest(t)

	term := model.Term{Name: "2025-2026-2", TermStartDate: activeTermStartDate()}
	if err := db.Create(&term).Error; err != nil {
		t.Fatalf("create term: %v", err)
	}

	classA := model.Class{Grade: 2025, MajorName: "密码科学与技术", ClassName: "密码科学与技术2501班", Status: 1}
	classB := model.Class{Grade: 2025, MajorName: "密码科学与技术", ClassName: "密码科学与技术2502班", Status: 1}
	if err := db.Create(&classA).Error; err != nil {
		t.Fatalf("create class A: %v", err)
	}
	if err := db.Create(&classB).Error; err != nil {
		t.Fatalf("create class B: %v", err)
	}

	studentClassSeed := model.Student{StudentNo: "20250001", StudentName: "甲", ClassID: &classA.ID, Status: 1}
	studentSameClass := model.Student{StudentNo: "20250002", StudentName: "乙", ClassID: &classA.ID, Status: 1}
	studentOtherClass := model.Student{StudentNo: "20250003", StudentName: "丙", ClassID: &classB.ID, Status: 1}
	if err := db.Create(&studentClassSeed).Error; err != nil {
		t.Fatalf("create seed student: %v", err)
	}
	if err := db.Create(&studentSameClass).Error; err != nil {
		t.Fatalf("create same-class student: %v", err)
	}
	if err := db.Create(&studentOtherClass).Error; err != nil {
		t.Fatalf("create other-class student: %v", err)
	}

	course := model.Course{TermID: term.ID, Grade: 2025, CourseName: "测试课程", TeacherName: "教师", Status: 1}
	if err := db.Create(&course).Error; err != nil {
		t.Fatalf("create course: %v", err)
	}
	group := model.CourseGroup{TermID: term.ID, CourseID: course.ID, Status: 1}
	if err := db.Create(&group).Error; err != nil {
		t.Fatalf("create course group: %v", err)
	}
	if err := db.Create(&model.CourseGroupStudent{
		TermID:        term.ID,
		CourseGroupID: group.ID,
		StudentID:     studentClassSeed.ID,
		ClassID:       &classA.ID,
		Status:        1,
	}).Error; err != nil {
		t.Fatalf("seed class member: %v", err)
	}

	if err := service.AddCourseGroupStudents(course.ID, group.ID, []uint64{studentSameClass.ID, studentOtherClass.ID}); err != nil {
		t.Fatalf("add course group students: %v", err)
	}

	var sameClassRelation model.CourseGroupStudent
	if err := db.Where("course_group_id = ? AND student_id = ?", group.ID, studentSameClass.ID).First(&sameClassRelation).Error; err != nil {
		t.Fatalf("reload same-class relation: %v", err)
	}
	if sameClassRelation.ClassID == nil || *sameClassRelation.ClassID != classA.ID {
		t.Fatalf("expected same-class student to inherit class identity %d, got %+v", classA.ID, sameClassRelation.ClassID)
	}

	var otherClassRelation model.CourseGroupStudent
	if err := db.Where("course_group_id = ? AND student_id = ?", group.ID, studentOtherClass.ID).First(&otherClassRelation).Error; err != nil {
		t.Fatalf("reload other-class relation: %v", err)
	}
	if otherClassRelation.ClassID != nil {
		t.Fatalf("expected other-class student to stay personal, got class_id %d", *otherClassRelation.ClassID)
	}
}

func TestListCoursesSupportsGradeClassAndStudentCountFilters(t *testing.T) {
	service, db := setupCourseServiceTest(t)

	term := model.Term{Name: "2025-2026-2", TermStartDate: activeTermStartDate()}
	if err := db.Create(&term).Error; err != nil {
		t.Fatalf("create term: %v", err)
	}

	classA := model.Class{Grade: 2025, MajorName: "密码科学与技术", ClassName: "密码科学与技术2501班", Status: 1}
	classB := model.Class{Grade: 2024, MajorName: "信息安全", ClassName: "信息安全2401班", Status: 1}
	if err := db.Create(&classA).Error; err != nil {
		t.Fatalf("create class A: %v", err)
	}
	if err := db.Create(&classB).Error; err != nil {
		t.Fatalf("create class B: %v", err)
	}

	studentA1 := model.Student{StudentNo: "20251001", StudentName: "甲", ClassID: &classA.ID, Status: 1}
	studentA2 := model.Student{StudentNo: "20251002", StudentName: "乙", ClassID: &classA.ID, Status: 1}
	studentB1 := model.Student{StudentNo: "20241001", StudentName: "丙", ClassID: &classB.ID, Status: 1}
	if err := db.Create(&studentA1).Error; err != nil {
		t.Fatalf("create student A1: %v", err)
	}
	if err := db.Create(&studentA2).Error; err != nil {
		t.Fatalf("create student A2: %v", err)
	}
	if err := db.Create(&studentB1).Error; err != nil {
		t.Fatalf("create student B1: %v", err)
	}

	courseMatched := model.Course{TermID: term.ID, Grade: 2025, CourseName: "公钥密码", TeacherName: "刘珍", Status: 1}
	courseOther := model.Course{TermID: term.ID, Grade: 2024, CourseName: "网络空间安全导论", TeacherName: "王五", Status: 1}
	if err := db.Create(&courseMatched).Error; err != nil {
		t.Fatalf("create matched course: %v", err)
	}
	if err := db.Create(&courseOther).Error; err != nil {
		t.Fatalf("create other course: %v", err)
	}

	groupMatched := model.CourseGroup{TermID: term.ID, CourseID: courseMatched.ID, Status: 1}
	groupOther := model.CourseGroup{TermID: term.ID, CourseID: courseOther.ID, Status: 1}
	if err := db.Create(&groupMatched).Error; err != nil {
		t.Fatalf("create matched group: %v", err)
	}
	if err := db.Create(&groupOther).Error; err != nil {
		t.Fatalf("create other group: %v", err)
	}

	for _, relation := range []model.CourseGroupStudent{
		{TermID: term.ID, CourseGroupID: groupMatched.ID, StudentID: studentA1.ID, ClassID: &classA.ID, Status: 1},
		{TermID: term.ID, CourseGroupID: groupMatched.ID, StudentID: studentA2.ID, ClassID: &classA.ID, Status: 1},
		{TermID: term.ID, CourseGroupID: groupOther.ID, StudentID: studentB1.ID, ClassID: &classB.ID, Status: 1},
	} {
		relation := relation
		if err := db.Create(&relation).Error; err != nil {
			t.Fatalf("create relation: %v", err)
		}
	}

	items, total, err := service.ListCourses(term.Name, "2025", "", "", classA.ClassName, "2", 1, 20)
	if err != nil {
		t.Fatalf("list courses: %v", err)
	}
	if total != 1 {
		t.Fatalf("expected 1 matched course, got total %d", total)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 course item, got %d", len(items))
	}
	if items[0].ID != courseMatched.ID {
		t.Fatalf("expected matched course %d, got %d", courseMatched.ID, items[0].ID)
	}
}
