package service

import (
	"gorm.io/gorm"

	"wack-backend/internal/query"
)

type LogService struct {
	logs       *query.LogQuery
	attendance *query.AttendanceQuery
}

func NewLogService(db *gorm.DB) *LogService {
	return &LogService{
		logs:       query.NewLogQuery(db),
		attendance: query.NewAttendanceQuery(db),
	}
}

func (s *LogService) AttendanceRecordLogs(input query.AttendanceRecordLogListInput) ([]query.AttendanceRecordLogItem, int64, error) {
	return s.logs.AttendanceRecordLogs(input)
}

func (s *LogService) RecordLogs(recordID uint64) ([]query.AttendanceRecordLogItem, error) {
	return s.attendance.AttendanceRecordLogs(recordID)
}
