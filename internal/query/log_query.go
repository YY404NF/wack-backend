package query

import (
	"time"

	"gorm.io/gorm"
)

type AdminOperationLogItem struct {
	ID                uint64    `json:"id"`
	OperatorUserID    uint64    `json:"operator_user_id"`
	OperatorStudentID string    `json:"operator_student_id"`
	TargetTable       string    `json:"target_table"`
	TargetID          uint64    `json:"target_id"`
	ActionType        string    `json:"action_type"`
	OldValue          *string   `json:"old_value"`
	NewValue          *string   `json:"new_value"`
	CreatedAt         time.Time `json:"created_at"`
}

type LogQuery struct {
	db *gorm.DB
}

func NewLogQuery(db *gorm.DB) *LogQuery {
	return &LogQuery{db: db}
}

func (q *LogQuery) AdminOperationLogs(page, pageSize int) ([]AdminOperationLogItem, int64, error) {
	query := q.db.Table("admin_operation_log").
		Select("admin_operation_log.id, admin_operation_log.operator_user_id, operator_user.student_id AS operator_student_id, admin_operation_log.target_table, admin_operation_log.target_id, admin_operation_log.action_type, admin_operation_log.old_value, admin_operation_log.new_value, admin_operation_log.created_at").
		Joins("JOIN user AS operator_user ON operator_user.id = admin_operation_log.operator_user_id")

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []AdminOperationLogItem
	if err := query.Order("admin_operation_log.created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Scan(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (q *LogQuery) AttendanceDetailLogs(page, pageSize int) ([]AttendanceDetailLogItem, int64, error) {
	query := q.db.Table("attendance_detail_log").
		Select("attendance_detail_log.id, attendance_detail_log.attendance_detail_id, attendance_detail_log.attendance_check_id, attendance_detail_log.user_id, target_user.student_id, attendance_detail_log.operator_user_id, operator_user.student_id AS operator_student_id, attendance_detail_log.old_status, attendance_detail_log.new_status, attendance_detail_log.operation_type, attendance_detail_log.operated_at, attendance_detail_log.created_at").
		Joins("JOIN user AS target_user ON target_user.id = attendance_detail_log.user_id").
		Joins("JOIN user AS operator_user ON operator_user.id = attendance_detail_log.operator_user_id")

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []AttendanceDetailLogItem
	if err := query.Order("attendance_detail_log.operated_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Scan(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}
