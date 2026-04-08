package query

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

type StudentItem struct {
	ID          uint64    `json:"id"`
	ClassID     *uint64   `json:"class_id"`
	StudentID   string    `json:"student_id"`
	RealName    string    `json:"real_name"`
	ClassName   *string   `json:"class_name"`
	Grade       *int      `json:"grade"`
	MajorName   *string   `json:"major_name"`
	LateCount   int64     `json:"late_count"`
	AbsentCount int64     `json:"absent_count"`
	LeaveCount  int64     `json:"leave_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type StudentOptionItem struct {
	ID        uint64  `json:"id"`
	StudentID string  `json:"student_id"`
	RealName  string  `json:"real_name"`
	ClassID   *uint64 `json:"class_id"`
	ClassName *string `json:"class_name"`
	Grade     *int    `json:"grade"`
	MajorName *string `json:"major_name"`
}

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

type StudentAttendanceItem struct {
	ID                  uint64     `json:"id"`
	CourseGroupLessonID uint64     `json:"course_group_lesson_id"`
	CourseGroupID       uint64     `json:"course_group_id"`
	CourseID            uint64     `json:"course_id"`
	TermID              uint64     `json:"term_id"`
	Term                string     `json:"term"`
	LessonDate          string     `json:"lesson_date"`
	WeekNo              int        `json:"week_no"`
	Weekday             int        `json:"weekday"`
	Section             int        `json:"section"`
	CourseName          string     `json:"course_name"`
	TeacherName         string     `json:"teacher_name"`
	Status              int        `json:"status"`
	OperatorName        string     `json:"operator_name"`
	OperatedAt          *time.Time `json:"operated_at"`
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

type StudentQuery struct {
	db *gorm.DB
}

func NewStudentQuery(db *gorm.DB) *StudentQuery {
	return &StudentQuery{db: db}
}

func (q *StudentQuery) attendanceSummaryBase(termName string) *gorm.DB {
	base := q.db.Table("attendance_record").
		Select(`
			attendance_record.student_id,
			SUM(CASE WHEN attendance_record.attendance_status = 1 THEN 1 ELSE 0 END) AS late_count,
			SUM(CASE WHEN attendance_record.attendance_status = 2 THEN 1 ELSE 0 END) AS absent_count,
			SUM(CASE WHEN attendance_record.attendance_status = 3 THEN 1 ELSE 0 END) AS leave_count
		`).
		Joins("JOIN term ON term.id = attendance_record.term_id")
	if value := strings.TrimSpace(termName); value != "" {
		base = base.Where("term.name = ?", value)
	}
	return base.Group("attendance_record.student_id")
}

func (q *StudentQuery) listStudentsBase(input ListStudentsInput) *gorm.DB {
	base := q.db.Table("student").
		Joins("LEFT JOIN class ON class.id = student.class_id").
		Joins("LEFT JOIN (?) AS attendance_summary ON attendance_summary.student_id = student.id", q.attendanceSummaryBase(input.Term)).
		Where("student.status = 1 AND (class.status = 1 OR student.class_id IS NULL)")

	if input.ClassID > 0 {
		base = base.Where("student.class_id = ?", input.ClassID)
	}

	keyword := strings.TrimSpace(input.Keyword)
	if keyword != "" {
		like := "%" + keyword + "%"
		base = base.Where(
			"student.student_no LIKE ? OR student.student_name LIKE ? OR COALESCE(class.class_name, '') LIKE ? OR COALESCE(class.major_name, '') LIKE ?",
			like,
			like,
			like,
			like,
		)
	}
	if value := strings.TrimSpace(input.StudentID); value != "" {
		base = base.Where("student.student_no LIKE ?", "%"+value+"%")
	}
	if value := strings.TrimSpace(input.RealName); value != "" {
		base = base.Where("student.student_name LIKE ?", "%"+value+"%")
	}
	if value := strings.TrimSpace(input.ClassName); value != "" {
		base = base.Where("COALESCE(class.class_name, '') LIKE ?", "%"+value+"%")
	}
	switch strings.TrimSpace(input.AttendanceSummaryStatus) {
	case "late":
		base = base.Where("COALESCE(attendance_summary.late_count, 0) > 0")
	case "absent":
		base = base.Where("COALESCE(attendance_summary.absent_count, 0) > 0")
	case "leave":
		base = base.Where("COALESCE(attendance_summary.leave_count, 0) > 0")
	}

	return base
}

func (q *StudentQuery) ListStudents(input ListStudentsInput) ([]StudentItem, int64, error) {
	base := q.listStudentsBase(input).
		Select(`
			student.id,
			student.class_id,
			student.student_no AS student_id,
			student.student_name AS real_name,
			class.class_name,
			class.grade,
			class.major_name,
			COALESCE(attendance_summary.late_count, 0) AS late_count,
			COALESCE(attendance_summary.absent_count, 0) AS absent_count,
			COALESCE(attendance_summary.leave_count, 0) AS leave_count,
			student.created_at,
			student.updated_at
		`)

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []StudentItem
	err := base.
		Order("student.id DESC").
		Offset((input.Page - 1) * input.PageSize).
		Limit(input.PageSize).
		Scan(&items).Error
	return items, total, err
}

func (q *StudentQuery) LocateStudentPage(input ListStudentsInput, focusStudentID uint64, pageSize int) (FocusPageResult, error) {
	base := q.listStudentsBase(input)

	var target struct {
		ID uint64 `gorm:"column:id"`
	}
	if err := base.Select("student.id").
		Where("student.id = ?", focusStudentID).
		Take(&target).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return FocusPageResult{}, nil
		}
		return FocusPageResult{}, err
	}

	var rowNo int64
	if err := q.listStudentsBase(input).
		Where("student.id >= ?", target.ID).
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

func (q *StudentQuery) GetStudent(id uint64) (StudentItem, error) {
	var item StudentItem
	err := q.db.Table("student").
		Select(`
			student.id,
			student.class_id,
			student.student_no AS student_id,
			student.student_name AS real_name,
			class.class_name,
			class.grade,
			class.major_name,
			0 AS late_count,
			0 AS absent_count,
			0 AS leave_count,
			student.created_at,
			student.updated_at
		`).
		Joins("LEFT JOIN class ON class.id = student.class_id").
		Where("student.id = ? AND student.status = 1 AND (class.status = 1 OR student.class_id IS NULL)", id).
		Scan(&item).Error
	if err == nil && item.ID == 0 {
		return StudentItem{}, gorm.ErrRecordNotFound
	}
	return item, err
}

func (q *StudentQuery) listStudentAttendanceBase(studentID uint64, input ListStudentAttendanceInput) *gorm.DB {
	lessonDateExpr := "date(term.term_start_date, printf('+%d days', (course_group_lesson.week_no - 1) * 7 + (course_group_lesson.weekday - 1)))"
	base := q.db.Table("attendance_record").
		Select(`
			attendance_record.id,
			attendance_record.course_group_lesson_id,
			course_group_lesson.course_group_id,
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
		Joins("JOIN course_group_lesson ON course_group_lesson.id = attendance_record.course_group_lesson_id").
		Joins("JOIN course_group ON course_group.id = course_group_lesson.course_group_id").
		Joins("JOIN course ON course.id = course_group.course_id").
		Joins("JOIN term ON term.id = course.term_id").
		Joins("LEFT JOIN user AS operator_user ON operator_user.id = attendance_record.updated_by_user_id").
		Where("attendance_record.student_id = ?", studentID)

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

func (q *StudentQuery) ListStudentAttendance(studentID uint64, input ListStudentAttendanceInput) ([]StudentAttendanceItem, int64, error) {
	base := q.listStudentAttendanceBase(studentID, input)

	var total int64
	if err := q.db.Table("(?) AS student_attendance_records", base).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []StudentAttendanceItem
	err := q.db.Table("(?) AS student_attendance_records", base).
		Order("lesson_date DESC, section DESC, id DESC").
		Offset((input.Page - 1) * input.PageSize).
		Limit(input.PageSize).
		Scan(&items).Error
	return items, total, err
}

func (q *StudentQuery) StudentOptions(keyword string, onlyUnbound bool) ([]StudentOptionItem, error) {
	base := q.db.Table("student").
		Select(`
			student.id,
			student.student_no AS student_id,
			student.student_name AS real_name,
			student.class_id,
			class.class_name,
			class.grade,
			class.major_name
		`).
		Joins("LEFT JOIN class ON class.id = student.class_id").
		Where("student.status = 1 AND (class.status = 1 OR student.class_id IS NULL)")
	if onlyUnbound {
		base = base.Where("student.class_id IS NULL")
	}
	if value := strings.TrimSpace(keyword); value != "" {
		like := "%" + value + "%"
		base = base.Where("student.student_no LIKE ? OR student.student_name LIKE ? OR COALESCE(class.class_name, '') LIKE ?", like, like, like)
	}
	var items []StudentOptionItem
	err := base.Order("student.student_no ASC, student.id ASC").Scan(&items).Error
	return items, err
}
