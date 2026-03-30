package service

import (
	"time"

	"gorm.io/gorm"

	"wack-backend/internal/model"
)

type auditLogger struct{}

func newAuditLogger() *auditLogger {
	return &auditLogger{}
}

func (l *auditLogger) logAttendanceStatusChange(tx *gorm.DB, record model.AttendanceRecord, operatorUserID uint64, oldStatus *int, newStatus int, operatedAt time.Time) error {
	logItem := model.AttendanceRecordLog{
		TermID:              record.TermID,
		AttendanceRecordID:  record.ID,
		OperatedByUserID:    operatorUserID,
		OldAttendanceStatus: oldStatus,
		NewAttendanceStatus: newStatus,
		CreatedAt:           operatedAt,
	}
	if err := tx.Create(&logItem).Error; err != nil {
		return err
	}
	return nil
}

func (l *auditLogger) logAttendanceStatusCreate(tx *gorm.DB, record model.AttendanceRecord, operatorUserID uint64, status int, operatedAt time.Time) error {
	return l.logAttendanceStatusChange(tx, record, operatorUserID, nil, status, operatedAt)
}
