package query

import (
	"strings"
	"time"

	"gorm.io/gorm"

	"wack-backend/internal/model"
)

type ClassStudentCandidateItem struct {
	ID        uint64 `json:"id"`
	ClassID   uint64 `json:"class_id"`
	StudentID string `json:"student_id"`
	RealName  string `json:"real_name"`
	ClassName string `json:"class_name"`
	Grade     int    `json:"grade"`
	MajorName string `json:"major_name"`
}

type ClassOptionItem struct {
	ID        uint64 `json:"id"`
	ClassName string `json:"class_name"`
	Grade     int    `json:"grade"`
	MajorName string `json:"major_name"`
}

type ClassAttendanceItem struct {
	ID           uint64     `json:"id"`
	StudentID    string     `json:"student_id"`
	RealName     string     `json:"real_name"`
	CourseID     uint64     `json:"course_id"`
	TermID       uint64     `json:"term_id"`
	Term         string     `json:"term"`
	LessonDate   string     `json:"lesson_date"`
	WeekNo       int        `json:"week_no"`
	Weekday      int        `json:"weekday"`
	Section      int        `json:"section"`
	CourseName   string     `json:"course_name"`
	TeacherName  string     `json:"teacher_name"`
	Status       int        `json:"status"`
	OperatorName string     `json:"operator_name"`
	OperatedAt   *time.Time `json:"operated_at"`
}

type ListClassAttendanceInput struct {
	Page         int
	PageSize     int
	Term         string
	LessonDate   string
	Section      string
	CourseName   string
	TeacherName  string
	StudentID    string
	RealName     string
	Status       string
	OperatorName string
	OperatedDate string
}

type ClassQuery struct {
	db *gorm.DB
}

func NewClassQuery(db *gorm.DB) *ClassQuery {
	return &ClassQuery{db: db}
}

func (q *ClassQuery) classAttendanceSummaryBase(termName string) *gorm.DB {
	base := q.db.Table("attendance_record").
		Select(`
			attendance_record.class_id,
			SUM(CASE WHEN attendance_record.attendance_status = 1 THEN 1 ELSE 0 END) AS late_count,
			SUM(CASE WHEN attendance_record.attendance_status = 2 THEN 1 ELSE 0 END) AS absent_count,
			SUM(CASE WHEN attendance_record.attendance_status = 3 THEN 1 ELSE 0 END) AS leave_count
		`).
		Joins("JOIN term ON term.id = attendance_record.term_id")
	if termName = strings.TrimSpace(termName); termName != "" {
		base = base.Where("term.name = ?", termName)
	}
	return base.Group("attendance_record.class_id")
}

func (q *ClassQuery) classListFilterBase(grade, majorName, className, termName, attendanceSummaryStatus string) *gorm.DB {
	base := q.db.Table("class").
		Joins("LEFT JOIN (?) AS attendance_summary ON attendance_summary.class_id = class.id", q.classAttendanceSummaryBase(termName)).
		Where("class.status = 1")

	if grade = strings.TrimSpace(grade); grade != "" {
		base = base.Where("CAST(class.grade AS TEXT) LIKE ?", "%"+grade+"%")
	}
	if majorName = strings.TrimSpace(majorName); majorName != "" {
		base = base.Where("class.major_name LIKE ?", "%"+majorName+"%")
	}
	if className = strings.TrimSpace(className); className != "" {
		base = base.Where("class.class_name LIKE ?", "%"+className+"%")
	}
	switch strings.TrimSpace(attendanceSummaryStatus) {
	case "late":
		base = base.Where("COALESCE(attendance_summary.late_count, 0) > 0")
	case "absent":
		base = base.Where("COALESCE(attendance_summary.absent_count, 0) > 0")
	case "leave":
		base = base.Where("COALESCE(attendance_summary.leave_count, 0) > 0")
	}

	return base
}

func (q *ClassQuery) ListClasses(grade, majorName, className, termName, attendanceSummaryStatus string, page, pageSize int) ([]model.Class, int64, error) {
	filterBase := q.classListFilterBase(grade, majorName, className, termName, attendanceSummaryStatus)
	base := filterBase.
		Select(`
			class.id,
			class.class_name,
			class.grade,
			class.major_name,
			class.status,
			class.created_at,
			class.updated_at,
			COUNT(student.id) AS student_count,
			COALESCE(attendance_summary.late_count, 0) AS late_count,
			COALESCE(attendance_summary.absent_count, 0) AS absent_count,
			COALESCE(attendance_summary.leave_count, 0) AS leave_count
		`).
		Joins("LEFT JOIN student ON student.class_id = class.id AND student.status = 1").
		Group("class.id")

	var total int64
	if err := filterBase.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []model.Class
	err := base.
		Order("class.id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&items).Error
	return items, total, err
}

func (q *ClassQuery) LocateClassPage(grade, majorName, className, termName, attendanceSummaryStatus string, focusClassID uint64, pageSize int) (FocusPageResult, error) {
	base := q.classListFilterBase(grade, majorName, className, termName, attendanceSummaryStatus)

	var target struct {
		ID uint64 `gorm:"column:id"`
	}
	if err := base.Select("class.id").
		Where("class.id = ?", focusClassID).
		Take(&target).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return FocusPageResult{}, nil
		}
		return FocusPageResult{}, err
	}

	var rowNo int64
	if err := q.classListFilterBase(grade, majorName, className, termName, attendanceSummaryStatus).
		Where("class.id >= ?", target.ID).
		Count(&rowNo).Error; err != nil {
		return FocusPageResult{}, err
	}
	if rowNo <= 0 {
		return FocusPageResult{}, nil
	}

	return FocusPageResult{
		Found:  true,
		Page:   int((rowNo-1)/int64(pageSize)) + 1,
		RowKey: target.ID,
	}, nil
}

func (q *ClassQuery) ClassByID(classID uint64) (model.Class, error) {
	var item model.Class
	err := q.db.Table("class").
		Select(`
			class.id,
			class.class_name,
			class.grade,
			class.major_name,
			class.status,
			class.created_at,
			class.updated_at,
			COUNT(student.id) AS student_count,
			0 AS late_count,
			0 AS absent_count,
			0 AS leave_count
		`).
		Joins("LEFT JOIN student ON student.class_id = class.id AND student.status = 1").
		Where("class.id = ?", classID).
		Group("class.id").
		Scan(&item).Error
	if err != nil {
		return model.Class{}, err
	}
	if item.ID == 0 {
		return model.Class{}, gorm.ErrRecordNotFound
	}
	return item, nil
}

func (q *ClassQuery) listClassAttendanceBase(classID uint64, input ListClassAttendanceInput) *gorm.DB {
	lessonDateExpr := "date(term.term_start_date, printf('+%d days', (course_group_lesson.week_no - 1) * 7 + (course_group_lesson.weekday - 1)))"
	base := q.db.Table("attendance_record").
		Select(`
			attendance_record.id,
			student.student_no AS student_id,
			student.student_name AS real_name,
			course.id AS course_id,
			term.id AS term_id,
			term.name AS term,
			`+lessonDateExpr+` AS lesson_date,
			course_group_lesson.week_no,
			course_group_lesson.weekday,
			course_group_lesson.section,
			course.course_name,
			course.teacher_name,
			attendance_record.attendance_status AS status,
			COALESCE(operator_user.real_name, '') AS operator_name,
			attendance_record.updated_at AS operated_at
		`).
		Joins("JOIN student ON student.id = attendance_record.student_id").
		Joins("JOIN course_group_lesson ON course_group_lesson.id = attendance_record.course_group_lesson_id").
		Joins("JOIN course_group ON course_group.id = course_group_lesson.course_group_id").
		Joins("JOIN course ON course.id = course_group.course_id").
		Joins("JOIN term ON term.id = course.term_id").
		Joins("LEFT JOIN user AS operator_user ON operator_user.id = attendance_record.updated_by_user_id").
		Where("attendance_record.class_id = ?", classID)

	if value := strings.TrimSpace(input.Term); value != "" {
		base = base.Where("term.name = ?", value)
	}
	if value := strings.TrimSpace(input.LessonDate); value != "" {
		base = base.Where(lessonDateExpr+" = ?", value)
	}
	if value := strings.TrimSpace(input.Section); value != "" {
		base = base.Where("course_group_lesson.section = ?", value)
	}
	if value := strings.TrimSpace(input.CourseName); value != "" {
		base = base.Where("course.course_name LIKE ?", "%"+value+"%")
	}
	if value := strings.TrimSpace(input.TeacherName); value != "" {
		base = base.Where("course.teacher_name LIKE ?", "%"+value+"%")
	}
	if value := strings.TrimSpace(input.StudentID); value != "" {
		base = base.Where("student.student_no LIKE ?", "%"+value+"%")
	}
	if value := strings.TrimSpace(input.RealName); value != "" {
		base = base.Where("student.student_name LIKE ?", "%"+value+"%")
	}
	if value := strings.TrimSpace(input.Status); value != "" {
		base = base.Where("attendance_record.attendance_status = ?", value)
	}
	if value := strings.TrimSpace(input.OperatorName); value != "" {
		base = base.Where("operator_user.real_name LIKE ?", "%"+value+"%")
	}
	if value := strings.TrimSpace(input.OperatedDate); value != "" {
		base = base.Where("DATE(attendance_record.updated_at) = ?", value)
	}

	return base
}

func (q *ClassQuery) ListClassAttendance(classID uint64, input ListClassAttendanceInput) ([]ClassAttendanceItem, int64, error) {
	base := q.listClassAttendanceBase(classID, input)

	var total int64
	if err := q.db.Table("(?) AS class_attendance_records", base).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []ClassAttendanceItem
	err := q.db.Table("(?) AS class_attendance_records", base).
		Order("lesson_date DESC, section DESC, id DESC").
		Offset((input.Page - 1) * input.PageSize).
		Limit(input.PageSize).
		Scan(&items).Error
	return items, total, err
}

type ClassStudentItem struct {
	ID        uint64    `json:"id"`
	ClassID   uint64    `json:"class_id"`
	StudentID string    `json:"student_id"`
	RealName  string    `json:"real_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (q *ClassQuery) ClassStudents(classID uint64) ([]ClassStudentItem, error) {
	var items []ClassStudentItem
	err := q.db.Table("student").
		Select("student.id, student.class_id, student.student_no AS student_id, student.student_name AS real_name, student.created_at, student.updated_at").
		Where("student.class_id = ? AND student.status = 1", classID).
		Order("student.student_no ASC").
		Scan(&items).Error
	return items, err
}

func (q *ClassQuery) StudentCandidates() ([]ClassStudentCandidateItem, error) {
	var items []ClassStudentCandidateItem
	err := q.db.Table("student").
		Select("student.id, student.class_id, student.student_no AS student_id, student.student_name AS real_name, class.class_name, class.grade, class.major_name").
		Joins("JOIN class ON class.id = student.class_id").
		Where("student.status = 1").
		Order("student.student_no ASC").
		Scan(&items).Error
	return items, err
}

func (q *ClassQuery) ClassOptions(keyword string) ([]ClassOptionItem, error) {
	query := q.db.Table("class").
		Select("class.id, class.class_name, class.grade, class.major_name").
		Where("class.status = 1")
	if keyword != "" {
		query = query.Where("class.class_name LIKE ? OR class.major_name LIKE ? OR CAST(class.grade AS TEXT) LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	var items []ClassOptionItem
	err := query.Order("class.grade DESC, class.major_name ASC, class.class_name ASC").Scan(&items).Error
	return items, err
}
