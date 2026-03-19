package service

import (
	"errors"
	"strings"

	"gorm.io/gorm"
	"wack-backend/internal/model"
	"wack-backend/internal/query"
)

type ListStudentsInput struct {
	Page     int
	PageSize int
	ClassID  uint64
	Keyword  string
}

type StudentService struct {
	db       *gorm.DB
	students *query.StudentQuery
}

func NewStudentService(db *gorm.DB) *StudentService {
	return &StudentService{
		db:       db,
		students: query.NewStudentQuery(db),
	}
}

func (s *StudentService) ListStudents(input ListStudentsInput) ([]query.StudentItem, int64, error) {
	if input.Page <= 0 {
		input.Page = 1
	}
	if input.PageSize <= 0 {
		input.PageSize = 20
	}
	return s.students.ListStudents(query.ListStudentsInput{
		Page:     input.Page,
		PageSize: input.PageSize,
		ClassID:  input.ClassID,
		Keyword:  input.Keyword,
	})
}

func (s *StudentService) CreateStudent(student model.Student) (query.StudentItem, error) {
	student.StudentNo = strings.TrimSpace(student.StudentNo)
	student.StudentName = strings.TrimSpace(student.StudentName)
	student.Status = 1
	if err := validateStudent(student); err != nil {
		return query.StudentItem{}, err
	}
	if student.ClassID != nil {
		if err := ensureClassExists(s.db, *student.ClassID); err != nil {
			return query.StudentItem{}, err
		}
	}
	if err := s.db.Create(&student).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) || strings.Contains(strings.ToLower(err.Error()), "unique") {
			return query.StudentItem{}, ErrStudentNoAlreadyExists
		}
		return query.StudentItem{}, err
	}
	return s.students.GetStudent(student.ID)
}

func (s *StudentService) UpdateStudent(id uint64, input model.Student) (query.StudentItem, error) {
	input.StudentNo = strings.TrimSpace(input.StudentNo)
	input.StudentName = strings.TrimSpace(input.StudentName)
	input.Status = 1
	if err := validateStudent(input); err != nil {
		return query.StudentItem{}, err
	}
	if input.ClassID != nil {
		if err := ensureClassExists(s.db, *input.ClassID); err != nil {
			return query.StudentItem{}, err
		}
	}

	var student model.Student
	if err := s.db.First(&student, "id = ? AND status = 1", id).Error; err != nil {
		return query.StudentItem{}, ErrStudentNotFound
	}
	if err := s.db.Model(&student).Updates(map[string]interface{}{
		"student_no":   input.StudentNo,
		"student_name": input.StudentName,
		"class_id":     input.ClassID,
	}).Error; err != nil {
		return query.StudentItem{}, err
	}
	return s.students.GetStudent(id)
}

func (s *StudentService) DeleteStudent(id uint64) error {
	result := s.db.Where("id = ? AND status = 1", id).Delete(&model.Student{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrStudentNotFound
	}
	return nil
}
