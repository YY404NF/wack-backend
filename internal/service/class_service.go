package service

import (
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"wack-backend/internal/model"
	"wack-backend/internal/query"
)

type ClassService struct {
	db      *gorm.DB
	classes *query.ClassQuery
}

func NewClassService(db *gorm.DB) *ClassService {
	return &ClassService{db: db, classes: query.NewClassQuery(db)}
}

func (s *ClassService) ListClasses(page, pageSize int) ([]model.Class, int64, error) {
	return s.classes.ListClasses(page, pageSize)
}

func (s *ClassService) CreateClass(class model.Class) (model.Class, error) {
	class.ClassName = strings.TrimSpace(class.ClassName)
	class.MajorName = strings.TrimSpace(class.MajorName)
	if class.ClassName == "" || class.MajorName == "" || class.Grade <= 0 {
		return model.Class{}, ErrInvalidInput
	}
	if len(class.ClassName) > 100 || len(class.MajorName) > 100 {
		return model.Class{}, ErrInvalidInput
	}
	return class, s.db.Create(&class).Error
}

func (s *ClassService) GetClass(id uint64) (model.Class, error) {
	classItem, err := s.classes.ClassByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return model.Class{}, ErrClassNotFound
		}
		return model.Class{}, err
	}
	return classItem, nil
}

func (s *ClassService) UpdateClass(id uint64, req model.Class) (model.Class, error) {
	req.ClassName = strings.TrimSpace(req.ClassName)
	req.MajorName = strings.TrimSpace(req.MajorName)
	if req.ClassName == "" || req.MajorName == "" || req.Grade <= 0 {
		return model.Class{}, ErrInvalidInput
	}
	if len(req.ClassName) > 100 || len(req.MajorName) > 100 {
		return model.Class{}, ErrInvalidInput
	}

	var class model.Class
	if err := s.db.First(&class, id).Error; err != nil {
		return model.Class{}, ErrClassNotFound
	}
	if err := s.db.Model(&class).Updates(map[string]interface{}{
		"class_name": req.ClassName,
		"grade":      req.Grade,
		"major_name": req.MajorName,
	}).Error; err != nil {
		return model.Class{}, err
	}
	return s.GetClass(id)
}

func (s *ClassService) DeleteClass(id uint64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("class_id = ?", id).Delete(&model.Student{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.Class{}, id).Error
	})
}

func (s *ClassService) GetClassStudents(id uint64) ([]query.ClassStudentItem, error) {
	if err := ensureClassExists(s.db, id); err != nil {
		return nil, err
	}
	return s.classes.ClassStudents(id)
}

func (s *ClassService) ListStudentCandidates() ([]query.ClassStudentCandidateItem, error) {
	return s.classes.StudentCandidates()
}

func (s *ClassService) CreateClassStudent(classID uint64, student model.Student) (query.ClassStudentItem, error) {
	student.ClassID = &classID
	student.Status = 1
	student.StudentNo = strings.TrimSpace(student.StudentNo)
	student.StudentName = strings.TrimSpace(student.StudentName)
	if err := validateStudent(student); err != nil {
		return query.ClassStudentItem{}, err
	}
	if err := ensureClassExists(s.db, classID); err != nil {
		return query.ClassStudentItem{}, err
	}
	if err := s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "student_no"}},
		DoUpdates: clause.AssignmentColumns([]string{"student_name", "class_id", "status", "updated_at"}),
	}).Create(&student).Error; err != nil {
		return query.ClassStudentItem{}, err
	}
	students, err := s.classes.ClassStudents(classID)
	if err != nil {
		return query.ClassStudentItem{}, err
	}
	for _, item := range students {
		if item.StudentID == student.StudentNo {
			return item, nil
		}
	}
	return query.ClassStudentItem{}, ErrClassNotFound
}

func (s *ClassService) UpdateClassStudent(classID, studentID uint64, input model.Student) (query.ClassStudentItem, error) {
	input.StudentNo = strings.TrimSpace(input.StudentNo)
	input.StudentName = strings.TrimSpace(input.StudentName)
	input.ClassID = &classID
	input.Status = 1
	if err := validateStudent(input); err != nil {
		return query.ClassStudentItem{}, err
	}
	var student model.Student
	if err := s.db.First(&student, "id = ? AND class_id = ?", studentID, classID).Error; err != nil {
		return query.ClassStudentItem{}, ErrClassNotFound
	}
	if err := s.db.Model(&student).Updates(map[string]interface{}{
		"student_no":   input.StudentNo,
		"student_name": input.StudentName,
	}).Error; err != nil {
		return query.ClassStudentItem{}, err
	}
	students, err := s.classes.ClassStudents(classID)
	if err != nil {
		return query.ClassStudentItem{}, err
	}
	for _, item := range students {
		if item.ID == student.ID {
			return item, nil
		}
	}
	return query.ClassStudentItem{}, ErrClassNotFound
}

func (s *ClassService) DeleteClassStudent(classID, studentID uint64) error {
	result := s.db.Where("id = ? AND class_id = ?", studentID, classID).Delete(&model.Student{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrClassNotFound
	}
	return nil
}

func ensureClassExists(db *gorm.DB, classID uint64) error {
	var count int64
	if err := db.Model(&model.Class{}).Where("id = ?", classID).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return ErrClassNotFound
	}
	return nil
}

func validateStudent(student model.Student) error {
	if student.StudentNo == "" || student.StudentName == "" {
		return ErrInvalidInput
	}
	if len(student.StudentNo) > 32 || len(student.StudentName) > 50 {
		return ErrInvalidInput
	}
	return nil
}
