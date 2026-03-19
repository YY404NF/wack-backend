package httpserver

import (
	"github.com/gin-gonic/gin"

	"wack-backend/internal/httpserver/dto"
	"wack-backend/internal/service"
)

func (h *apiHandler) getSystemSetting(c *gin.Context) {
	setting, err := h.systemSettings.GetSystemSetting()
	if err != nil {
		fail(c, 500, "load system setting failed")
		return
	}
	ok(c, setting)
}

func (h *apiHandler) updateSystemSetting(c *gin.Context) {
	var req dto.UpdateSystemSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "invalid request")
		return
	}

	setting, err := h.systemSettings.UpdateSystemSetting(req.CurrentTermStartDate)
	if err != nil {
		switch {
		case service.IsServiceError(err, service.ErrInvalidInput):
			fail(c, 400, "invalid request")
		default:
			fail(c, 400, "update system setting failed")
		}
		return
	}
	ok(c, setting)
}
