package service

import (
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
	query := s.db.Table("class")
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var classes []model.Class
	if err := s.db.Table("class AS class_item").
		Select("class_item.id, class_item.class_name, class_item.grade, class_item.major_name, class_item.created_at, class_item.updated_at, COUNT(class_student.id) AS student_count").
		Joins("LEFT JOIN class_student ON class_student.class_id = class_item.id").
		Group("class_item.id, class_item.class_name, class_item.grade, class_item.major_name, class_item.created_at, class_item.updated_at").
		Order("class_item.grade DESC, class_item.major_name ASC, class_item.class_name ASC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&classes).Error; err != nil {
		return nil, 0, err
	}
	return classes, total, nil
}

func (s *ClassService) CreateClass(class model.Class) (model.Class, error) {
	return class, s.db.Create(&class).Error
}

func (s *ClassService) GetClass(id uint64) (model.Class, error) {
	var class model.Class
	if err := s.db.First(&class, id).Error; err != nil {
		return model.Class{}, ErrClassNotFound
	}
	return class, nil
}

func (s *ClassService) UpdateClass(id uint64, req model.Class) (model.Class, error) {
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
	if err := s.db.First(&class, id).Error; err != nil {
		return model.Class{}, err
	}
	return class, nil
}

func (s *ClassService) DeleteClass(id uint64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("class_id = ?", id).Delete(&model.ClassStudent{}).Error; err != nil {
			return err
		}
		if err := tx.Where("class_id = ?", id).Delete(&model.CourseClass{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.Class{}, id).Error
	})
}

func (s *ClassService) GetClassStudents(id uint64) ([]model.ClassStudent, error) {
	var students []model.ClassStudent
	err := s.db.Where("class_id = ?", id).
		Order("student_id ASC").
		Find(&students).Error
	return students, err
}

func (s *ClassService) ListStudentCandidates() ([]query.ClassStudentCandidateItem, error) {
	return s.classes.StudentCandidates()
}

func (s *ClassService) CreateClassStudent(classID uint64, student model.ClassStudent) (model.ClassStudent, error) {
	student.ClassID = classID
	err := s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "class_id"}, {Name: "student_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"real_name", "updated_at"}),
	}).Create(&student).Error
	if err != nil {
		return model.ClassStudent{}, err
	}
	var created model.ClassStudent
	if err := s.db.First(&created, "class_id = ? AND student_id = ?", classID, student.StudentID).Error; err != nil {
		return model.ClassStudent{}, err
	}
	return created, nil
}

func (s *ClassService) ImportClassStudents(classID uint64, students []model.ClassStudent) error {
	if len(students) == 0 {
		return nil
	}
	for i := range students {
		students[i].ClassID = classID
	}
	return s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "class_id"}, {Name: "student_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"real_name", "updated_at"}),
	}).Create(&students).Error
}

func (s *ClassService) UpdateClassStudent(classID, studentID uint64, input model.ClassStudent) (model.ClassStudent, error) {
	var student model.ClassStudent
	if err := s.db.First(&student, "id = ? AND class_id = ?", studentID, classID).Error; err != nil {
		return model.ClassStudent{}, ErrClassNotFound
	}
	if err := s.db.Model(&student).Updates(map[string]interface{}{
		"student_id": input.StudentID,
		"real_name":  input.RealName,
	}).Error; err != nil {
		return model.ClassStudent{}, err
	}
	if err := s.db.First(&student, student.ID).Error; err != nil {
		return model.ClassStudent{}, err
	}
	return student, nil
}

func (s *ClassService) DeleteClassStudent(classID, studentID uint64) error {
	result := s.db.Where("id = ? AND class_id = ?", studentID, classID).Delete(&model.ClassStudent{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrClassNotFound
	}
	return nil
}
