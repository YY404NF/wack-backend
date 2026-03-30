package query

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

type LogQuery struct {
	db *gorm.DB
}

func NewLogQuery(db *gorm.DB) *LogQuery {
	return &LogQuery{db: db}
}

type AttendanceRecordLogListInput struct {
	Term                string
	CourseGroupLessonID string
	LessonDate          string
	Section             string
	CourseName          string
	TeacherName         string
	StudentID           string
	RealName            string
	ClassName           string
	OldStatus           string
	NewStatus           string
	OperatorName        string
	OperatedDate        string
	Page                int
	PageSize            int
}

type AttendanceRecordLogListItem struct {
	ID                  uint64    `json:"id"`
	AttendanceRecordID  uint64    `json:"attendance_record_id"`
	CourseGroupLessonID uint64    `json:"course_group_lesson_id"`
	Term                string    `json:"term"`
	LessonDate          string    `json:"lesson_date"`
	Section             int       `json:"section"`
	CourseName          string    `json:"course_name"`
	TeacherName         string    `json:"teacher_name"`
	StudentID           string    `json:"student_id"`
	RealName            string    `json:"real_name"`
	ClassName           string    `json:"class_name"`
	OldStatus           *int      `json:"old_status"`
	NewStatus           int       `json:"new_status"`
	OperatorName        string    `json:"operator_name"`
	OperatedAt          time.Time `json:"operated_at"`
}

func (q *LogQuery) AttendanceRecordLogs(input AttendanceRecordLogListInput) ([]AttendanceRecordLogListItem, int64, error) {
	lessonDateExpr := "date(term.term_start_date, printf('+%d days', (course_group_lesson.week_no - 1) * 7 + (course_group_lesson.weekday - 1)))"

	base := q.db.Table("attendance_record_log").
		Joins("JOIN attendance_record ON attendance_record.id = attendance_record_log.attendance_record_id").
		Joins("JOIN course_group_lesson ON course_group_lesson.id = attendance_record.course_group_lesson_id").
		Joins("JOIN course ON course.id = attendance_record.course_id").
		Joins("JOIN term ON term.id = attendance_record.term_id").
		Joins("JOIN student ON student.id = attendance_record.student_id").
		Joins("LEFT JOIN class ON class.id = attendance_record.class_id").
		Joins("JOIN user AS operator_user ON operator_user.id = attendance_record_log.operated_by_user_id")
	if value := strings.TrimSpace(input.Term); value != "" {
		base = base.Where("term.name = ?", value)
	}
	if value := strings.TrimSpace(input.CourseGroupLessonID); value != "" {
		base = base.Where("attendance_record.course_group_lesson_id = ?", value)
	}
	if value := strings.TrimSpace(input.LessonDate); value != "" {
		base = base.Where(fmt.Sprintf("%s = ?", lessonDateExpr), value)
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
	if value := strings.TrimSpace(input.ClassName); value != "" {
		base = base.Where("COALESCE(class.class_name, '其他学生') LIKE ?", "%"+value+"%")
	}
	if value := strings.TrimSpace(input.OldStatus); value != "" {
		if value == "none" {
			base = base.Where("attendance_record_log.old_attendance_status IS NULL")
		} else {
			base = base.Where("attendance_record_log.old_attendance_status = ?", value)
		}
	}
	if value := strings.TrimSpace(input.NewStatus); value != "" {
		base = base.Where("attendance_record_log.new_attendance_status = ?", value)
	}
	if value := strings.TrimSpace(input.OperatorName); value != "" {
		base = base.Where("operator_user.real_name LIKE ?", "%"+value+"%")
	}
	if value := strings.TrimSpace(input.OperatedDate); value != "" {
		base = base.Where("DATE(attendance_record_log.created_at) = ?", value)
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []AttendanceRecordLogListItem
	selectSQL := fmt.Sprintf(`
		attendance_record_log.id,
		attendance_record_log.attendance_record_id,
		attendance_record.course_group_lesson_id AS course_group_lesson_id,
		term.name AS term,
		%s AS lesson_date,
		course_group_lesson.section,
		course.course_name,
		course.teacher_name,
		student.student_no AS student_id,
		student.student_name AS real_name,
		COALESCE(class.class_name, '其他学生') AS class_name,
		attendance_record_log.old_attendance_status AS old_status,
		attendance_record_log.new_attendance_status AS new_status,
		operator_user.real_name AS operator_name,
		attendance_record_log.created_at AS operated_at
	`, lessonDateExpr)
	if err := base.
		Select(selectSQL).
		Order("attendance_record_log.created_at DESC").
		Offset((input.Page - 1) * input.PageSize).
		Limit(input.PageSize).
		Scan(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}
