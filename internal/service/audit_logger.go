package service

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"

	"wack-backend/internal/model"
)

type attendanceStatusLogValue struct {
	Status int `json:"status"`
}

type auditLogger struct{}

func newAuditLogger() *auditLogger {
	return &auditLogger{}
}

func (l *auditLogger) logAttendanceStatusChange(tx *gorm.DB, detail model.AttendanceDetail, operatorUserID uint64, oldStatus, newStatus int, operatedAt time.Time, writeAdminLog bool) error {
	logItem := model.AttendanceDetailLog{
		AttendanceDetailID: detail.ID,
		AttendanceCheckID:  detail.AttendanceCheckID,
		StudentID:          detail.StudentID,
		RealName:           detail.RealName,
		OperatorUserID:     operatorUserID,
		OldStatus:          &oldStatus,
		NewStatus:          newStatus,
		OperationType:      "set_status",
		OperatedAt:         operatedAt,
	}
	if err := tx.Create(&logItem).Error; err != nil {
		return err
	}
	if !writeAdminLog {
		return nil
	}

	oldValueBytes, _ := json.Marshal(attendanceStatusLogValue{Status: oldStatus})
	newValueBytes, _ := json.Marshal(attendanceStatusLogValue{Status: newStatus})
	oldValue := string(oldValueBytes)
	newValue := string(newValueBytes)
	adminLog := model.AdminOperationLog{
		OperatorUserID: operatorUserID,
		TargetTable:    "attendance_detail",
		TargetID:       detail.ID,
		ActionType:     "update",
		OldValue:       &oldValue,
		NewValue:       &newValue,
	}
	return tx.Create(&adminLog).Error
}
