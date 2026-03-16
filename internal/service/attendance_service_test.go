package service

import (
	"testing"

	"wack-backend/internal/model"
)

func TestUpdateAttendanceStatusWritesAuditLogs(t *testing.T) {
	db := openTestDB(t)
	svc := NewAttendanceService(db)

	detail := model.AttendanceDetail{
		AttendanceCheckID: 1,
		StudentID:         "20210001",
		RealName:          "Alice",
		Status:            model.AttendanceUnset,
	}
	if err := db.Create(&detail).Error; err != nil {
		t.Fatalf("create attendance detail: %v", err)
	}

	if err := svc.UpdateAttendanceStatus(detail.ID, model.AttendancePresent, 99, true); err != nil {
		t.Fatalf("update attendance status: %v", err)
	}

	var updated model.AttendanceDetail
	if err := db.First(&updated, detail.ID).Error; err != nil {
		t.Fatalf("reload attendance detail: %v", err)
	}
	if updated.Status != model.AttendancePresent {
		t.Fatalf("expected status %d, got %d", model.AttendancePresent, updated.Status)
	}

	var detailLogs []model.AttendanceDetailLog
	if err := db.Find(&detailLogs).Error; err != nil {
		t.Fatalf("load attendance detail logs: %v", err)
	}
	if len(detailLogs) != 1 {
		t.Fatalf("expected 1 attendance detail log, got %d", len(detailLogs))
	}
	if detailLogs[0].NewStatus != model.AttendancePresent {
		t.Fatalf("expected new status %d, got %d", model.AttendancePresent, detailLogs[0].NewStatus)
	}

	var adminLogs []model.AdminOperationLog
	if err := db.Find(&adminLogs).Error; err != nil {
		t.Fatalf("load admin operation logs: %v", err)
	}
	if len(adminLogs) != 1 {
		t.Fatalf("expected 1 admin log, got %d", len(adminLogs))
	}
	if adminLogs[0].TargetTable != "attendance_detail" {
		t.Fatalf("expected target table attendance_detail, got %s", adminLogs[0].TargetTable)
	}
}

func TestUpdateAttendanceStatusSkipsAdminLogForStudentAction(t *testing.T) {
	db := openTestDB(t)
	svc := NewAttendanceService(db)

	detail := model.AttendanceDetail{
		AttendanceCheckID: 1,
		StudentID:         "20210002",
		RealName:          "Bob",
		Status:            model.AttendanceUnset,
	}
	if err := db.Create(&detail).Error; err != nil {
		t.Fatalf("create attendance detail: %v", err)
	}

	if err := svc.UpdateAttendanceStatus(detail.ID, model.AttendanceLate, 100, false); err != nil {
		t.Fatalf("update attendance status: %v", err)
	}

	var adminLogCount int64
	if err := db.Model(&model.AdminOperationLog{}).Count(&adminLogCount).Error; err != nil {
		t.Fatalf("count admin logs: %v", err)
	}
	if adminLogCount != 0 {
		t.Fatalf("expected 0 admin logs, got %d", adminLogCount)
	}
}
