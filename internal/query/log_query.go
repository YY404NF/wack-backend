package query

import (
	"strings"

	"gorm.io/gorm"
)

type LogQuery struct {
	db *gorm.DB
}

func NewLogQuery(db *gorm.DB) *LogQuery {
	return &LogQuery{db: db}
}

type AttendanceRecordLogListInput struct {
	StudentID       string
	OperatorLoginID string
	OperationType   string
	NewStatus       string
	OperatedDate    string
	Page            int
	PageSize        int
}

func (q *LogQuery) AttendanceRecordLogs(input AttendanceRecordLogListInput) ([]AttendanceRecordLogItem, int64, error) {
	query := q.db.Table("attendance_record_log").
		Select(`
			attendance_record_log.id,
			attendance_record_log.attendance_record_id,
			attendance_record.course_group_lesson_id AS course_group_lesson_id,
			student.student_no AS student_id,
			student.student_name AS real_name,
			attendance_record_log.operated_by_user_id AS operator_user_id,
			operator_user.login_id AS operator_login_id,
			attendance_record_log.old_attendance_status AS old_status,
			attendance_record_log.new_attendance_status AS new_status,
			CASE
				WHEN attendance_record_log.old_attendance_status = attendance_record_log.new_attendance_status THEN 'create_record'
				ELSE 'set_status'
			END AS operation_type,
			attendance_record_log.created_at AS operated_at,
			attendance_record_log.created_at
		`).
		Joins("JOIN attendance_record ON attendance_record.id = attendance_record_log.attendance_record_id").
		Joins("JOIN student ON student.id = attendance_record.student_id").
		Joins("JOIN user AS operator_user ON operator_user.id = attendance_record_log.operated_by_user_id")

	if value := strings.TrimSpace(input.StudentID); value != "" {
		query = query.Where("student.student_no LIKE ?", "%"+value+"%")
	}
	if value := strings.TrimSpace(input.OperatorLoginID); value != "" {
		query = query.Where("operator_user.login_id LIKE ?", "%"+value+"%")
	}
	if value := strings.TrimSpace(input.OperationType); value != "" {
		query = query.Where(`
			CASE
				WHEN attendance_record_log.old_attendance_status = attendance_record_log.new_attendance_status THEN 'create_record'
				ELSE 'set_status'
			END LIKE ?
		`, "%"+value+"%")
	}
	if value := strings.TrimSpace(input.NewStatus); value != "" {
		query = query.Where("attendance_record_log.new_attendance_status = ?", value)
	}
	if value := strings.TrimSpace(input.OperatedDate); value != "" {
		query = query.Where("DATE(attendance_record_log.created_at) = ?", value)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []AttendanceRecordLogItem
	if err := query.Order("attendance_record_log.created_at DESC").Offset((input.Page - 1) * input.PageSize).Limit(input.PageSize).Scan(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}
