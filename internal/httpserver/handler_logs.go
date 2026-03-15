package httpserver

import (
	"time"

	"github.com/gin-gonic/gin"
)

func (h *apiHandler) adminAttendanceDetailLogs(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	type logItem struct {
		ID                 uint64    `json:"id"`
		AttendanceDetailID uint64    `json:"attendance_detail_id"`
		AttendanceCheckID  uint64    `json:"attendance_check_id"`
		UserID             uint64    `json:"user_id"`
		StudentID          string    `json:"student_id"`
		OperatorUserID     uint64    `json:"operator_user_id"`
		OperatorStudentID  string    `json:"operator_student_id"`
		OldStatus          *int      `json:"old_status"`
		NewStatus          int       `json:"new_status"`
		OperationType      string    `json:"operation_type"`
		OperatedAt         time.Time `json:"operated_at"`
		CreatedAt          time.Time `json:"created_at"`
	}
	var logs []logItem
	if err := h.db.Table("attendance_detail_log").
		Select("attendance_detail_log.id, attendance_detail_log.attendance_detail_id, attendance_detail_log.attendance_check_id, attendance_detail_log.user_id, target_user.student_id, attendance_detail_log.operator_user_id, operator_user.student_id AS operator_student_id, attendance_detail_log.old_status, attendance_detail_log.new_status, attendance_detail_log.operation_type, attendance_detail_log.operated_at, attendance_detail_log.created_at").
		Joins("JOIN user AS target_user ON target_user.id = attendance_detail_log.user_id").
		Joins("JOIN user AS operator_user ON operator_user.id = attendance_detail_log.operator_user_id").
		Where("attendance_detail_log.attendance_detail_id = ?", id).
		Order("attendance_detail_log.operated_at DESC").
		Scan(&logs).Error; err != nil {
		fail(c, 500, "load attendance detail logs failed")
		return
	}
	ok(c, logs)
}

func (h *apiHandler) listAdminOperationLogs(c *gin.Context) {
	page, pageSize := parsePage(c)
	query := h.db.Table("admin_operation_log").
		Select("admin_operation_log.id, admin_operation_log.operator_user_id, operator_user.student_id AS operator_student_id, admin_operation_log.target_table, admin_operation_log.target_id, admin_operation_log.action_type, admin_operation_log.old_value, admin_operation_log.new_value, admin_operation_log.created_at").
		Joins("JOIN user AS operator_user ON operator_user.id = admin_operation_log.operator_user_id")
	type logItem struct {
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
	var logs []logItem
	total, err := paginate(query.Order("admin_operation_log.created_at DESC"), page, pageSize, &logs)
	if err != nil {
		fail(c, 500, "list admin operation logs failed")
		return
	}
	ok(c, pageResult{Items: logs, Page: page, PageSize: pageSize, Total: total})
}

func (h *apiHandler) listAttendanceDetailLogs(c *gin.Context) {
	page, pageSize := parsePage(c)
	query := h.db.Table("attendance_detail_log").
		Select("attendance_detail_log.id, attendance_detail_log.attendance_detail_id, attendance_detail_log.attendance_check_id, attendance_detail_log.user_id, target_user.student_id, attendance_detail_log.operator_user_id, operator_user.student_id AS operator_student_id, attendance_detail_log.old_status, attendance_detail_log.new_status, attendance_detail_log.operation_type, attendance_detail_log.operated_at, attendance_detail_log.created_at").
		Joins("JOIN user AS target_user ON target_user.id = attendance_detail_log.user_id").
		Joins("JOIN user AS operator_user ON operator_user.id = attendance_detail_log.operator_user_id")
	type logItem struct {
		ID                 uint64    `json:"id"`
		AttendanceDetailID uint64    `json:"attendance_detail_id"`
		AttendanceCheckID  uint64    `json:"attendance_check_id"`
		UserID             uint64    `json:"user_id"`
		StudentID          string    `json:"student_id"`
		OperatorUserID     uint64    `json:"operator_user_id"`
		OperatorStudentID  string    `json:"operator_student_id"`
		OldStatus          *int      `json:"old_status"`
		NewStatus          int       `json:"new_status"`
		OperationType      string    `json:"operation_type"`
		OperatedAt         time.Time `json:"operated_at"`
		CreatedAt          time.Time `json:"created_at"`
	}
	var logs []logItem
	total, err := paginate(query.Order("attendance_detail_log.operated_at DESC"), page, pageSize, &logs)
	if err != nil {
		fail(c, 500, "list attendance detail logs failed")
		return
	}
	ok(c, pageResult{Items: logs, Page: page, PageSize: pageSize, Total: total})
}
