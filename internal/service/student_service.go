package service

import (
	"errors"
	"strings"

	"gorm.io/gorm"
	"wack-backend/internal/model"
	"wack-backend/internal/query"
)

type ListStudentsInput struct {
	Page                    int
	PageSize                int
	ClassID                 uint64
	Keyword                 string
	StudentID               string
	RealName                string
	ClassName               string
	Term                    string
	AttendanceSummaryStatus string
}

type ListStudentAttendanceInput struct {
	Page         int
	PageSize     int
	Term         string
	LessonDate   string
	Section      string
	CourseName   string
	TeacherName  string
	Status       string
	OperatorName string
	OperatedDate string
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
		Page:                    input.Page,
		PageSize:                input.PageSize,
		ClassID:                 input.ClassID,
		Keyword:                 input.Keyword,
		StudentID:               input.StudentID,
		RealName:                input.RealName,
		ClassName:               input.ClassName,
		Term:                    input.Term,
		AttendanceSummaryStatus: input.AttendanceSummaryStatus,
	})
}

func (s *StudentService) LocateStudentPage(input ListStudentsInput, focusStudentID uint64) (query.FocusPageResult, error) {
	if input.PageSize <= 0 {
		input.PageSize = 20
	}
	return s.students.LocateStudentPage(query.ListStudentsInput{
		Page:                    input.Page,
		PageSize:                input.PageSize,
		ClassID:                 input.ClassID,
		Keyword:                 input.Keyword,
		StudentID:               input.StudentID,
		RealName:                input.RealName,
		ClassName:               input.ClassName,
		Term:                    input.Term,
		AttendanceSummaryStatus: input.AttendanceSummaryStatus,
	}, focusStudentID, input.PageSize)
}

func (s *StudentService) GetStudentAttendancePage(studentID uint64, input ListStudentAttendanceInput) (query.StudentItem, []query.StudentAttendanceItem, int64, error) {
	if input.Page <= 0 {
		input.Page = 1
	}
	if input.PageSize <= 0 {
		input.PageSize = 20
	}

	student, err := s.students.GetStudent(studentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return query.StudentItem{}, nil, 0, ErrStudentNotFound
		}
		return query.StudentItem{}, nil, 0, err
	}
	if student.ID == 0 {
		return query.StudentItem{}, nil, 0, ErrStudentNotFound
	}

	items, total, err := s.students.ListStudentAttendance(studentID, query.ListStudentAttendanceInput{
		Page:         input.Page,
		PageSize:     input.PageSize,
		Term:         input.Term,
		LessonDate:   input.LessonDate,
		Section:      input.Section,
		CourseName:   input.CourseName,
		TeacherName:  input.TeacherName,
		Status:       input.Status,
		OperatorName: input.OperatorName,
		OperatedDate: input.OperatedDate,
	})
	if err != nil {
		return query.StudentItem{}, nil, 0, err
	}

	return student, items, total, nil
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
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&student).Error; err != nil {
			return err
		}
		return syncStudentClassMembership(tx, student.ID, nil, student.ClassID)
	}); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) || strings.Contains(strings.ToLower(err.Error()), "unique") {
			return query.StudentItem{}, ErrStudentNoAlreadyExists
		}
		return query.StudentItem{}, err
	}
	return s.students.GetStudent(student.ID)
}

func (s *StudentService) GetStudent(id uint64) (query.StudentItem, error) {
	student, err := s.students.GetStudent(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return query.StudentItem{}, ErrStudentNotFound
		}
		return query.StudentItem{}, err
	}
	if student.ID == 0 {
		return query.StudentItem{}, ErrStudentNotFound
	}
	return student, nil
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
	oldClassID := student.ClassID
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&student).Updates(map[string]interface{}{
			"student_no":   input.StudentNo,
			"student_name": input.StudentName,
			"class_id":     input.ClassID,
		}).Error; err != nil {
			return err
		}
		return syncStudentClassMembership(tx, student.ID, oldClassID, input.ClassID)
	}); err != nil {
		return query.StudentItem{}, err
	}
	return s.students.GetStudent(id)
}

func (s *StudentService) DeleteStudent(id uint64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var student model.Student
		if err := tx.Where("id = ? AND status = 1", id).First(&student).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return ErrStudentNotFound
			}
			return err
		}

		var recordCount int64
		if err := tx.Model(&model.AttendanceRecord{}).
			Where("student_id = ?", id).
			Count(&recordCount).Error; err != nil {
			return err
		}

		if recordCount > 0 {
			if err := tx.Model(&model.CourseGroupStudent{}).
				Where("student_id = ? AND status = 1", id).
				Updates(map[string]interface{}{
					"status":     2,
					"updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
				}).Error; err != nil {
				return err
			}
			return tx.Model(&model.Student{}).
				Where("id = ? AND status = 1", id).
				Updates(map[string]interface{}{
					"status":     2,
					"updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
				}).Error
		}

		if err := tx.Where("student_id = ?", id).Delete(&model.CourseGroupStudent{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.Student{}, id).Error
	})
}

func (s *StudentService) ListStudentOptions(keyword string, onlyUnbound bool) ([]query.StudentOptionItem, error) {
	return s.students.StudentOptions(keyword, onlyUnbound)
}
