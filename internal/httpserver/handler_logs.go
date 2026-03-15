package httpserver

import (
	"github.com/gin-gonic/gin"

	"wack-backend/internal/query"
)

func (h *apiHandler) adminAttendanceDetailLogs(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	logs, err := h.logs.DetailLogs(id)
	if err != nil {
		fail(c, 500, "load attendance detail logs failed")
		return
	}
	ok(c, logs)
}

func (h *apiHandler) listAdminOperationLogs(c *gin.Context) {
	page, pageSize := parsePage(c)
	logs, total, err := h.logs.AdminOperationLogs(page, pageSize)
	if err != nil {
		fail(c, 500, "list admin operation logs failed")
		return
	}
	ok(c, pageResult[query.AdminOperationLogItem]{Items: logs, Page: page, PageSize: pageSize, Total: total})
}

func (h *apiHandler) listAttendanceDetailLogs(c *gin.Context) {
	page, pageSize := parsePage(c)
	logs, total, err := h.logs.AttendanceDetailLogs(page, pageSize)
	if err != nil {
		fail(c, 500, "list attendance detail logs failed")
		return
	}
	ok(c, pageResult[query.AttendanceDetailLogItem]{Items: logs, Page: page, PageSize: pageSize, Total: total})
}
