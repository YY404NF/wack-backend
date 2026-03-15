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

func (s *LogService) AdminOperationLogs(page, pageSize int) ([]query.AdminOperationLogItem, int64, error) {
	return s.logs.AdminOperationLogs(page, pageSize)
}

func (s *LogService) AttendanceDetailLogs(page, pageSize int) ([]query.AttendanceDetailLogItem, int64, error) {
	return s.logs.AttendanceDetailLogs(page, pageSize)
}

func (s *LogService) DetailLogs(detailID uint64) ([]query.AttendanceDetailLogItem, error) {
	return s.attendance.AttendanceDetailLogs(detailID)
}
