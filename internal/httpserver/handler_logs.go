package httpserver

import (
	"github.com/gin-gonic/gin"

	"wack-backend/internal/query"
)

func (h *apiHandler) adminAttendanceRecordLogs(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	logs, err := h.logs.RecordLogs(id)
	if err != nil {
		fail(c, 500, "load attendance record logs failed")
		return
	}
	ok(c, logs)
}

func (h *apiHandler) listAttendanceRecordLogs(c *gin.Context) {
	page, pageSize := parsePage(c)
	logs, total, err := h.logs.AttendanceRecordLogs(query.AttendanceRecordLogListInput{
		StudentID:       c.Query("student_id"),
		OperatorLoginID: c.Query("operator_login_id"),
		OperationType:   c.Query("operation_type"),
		NewStatus:       c.Query("new_status"),
		OperatedDate:    c.Query("operated_date"),
		Page:            page,
		PageSize:        pageSize,
	})
	if err != nil {
		fail(c, 500, "list attendance record logs failed")
		return
	}
	ok(c, pageResult[query.AttendanceRecordLogItem]{Items: logs, Page: page, PageSize: pageSize, Total: total})
}
