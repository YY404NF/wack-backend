package httpserver

import (
	"github.com/gin-gonic/gin"

	"wack-backend/internal/service"
)

func (h *authHandler) setupStatus(c *gin.Context) {
	initialized, err := h.auth.HasAnyAdmin()
	if err != nil {
		fail(c, 500, "load setup status failed")
		return
	}

	ok(c, gin.H{
		"initialized": initialized,
	})
}

func (h *authHandler) initializeSystem(c *gin.Context) {
	var req struct {
		StudentID string `json:"student_id" binding:"required"`
		RealName  string `json:"real_name" binding:"required"`
		Password  string `json:"password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "invalid request")
		return
	}

	initialized, err := h.auth.HasAnyAdmin()
	if err != nil {
		fail(c, 500, "load setup status failed")
		return
	}
	if initialized {
		fail(c, 409, "system already initialized")
		return
	}

	if err := h.auth.InitializeSystem(req.StudentID, req.RealName, req.Password); err != nil {
		if service.IsServiceError(err, service.ErrSystemAlreadyInitialized) {
			fail(c, 409, "system already initialized")
			return
		}
		fail(c, 500, "initialize system failed")
		return
	}

	ok(c, gin.H{
		"initialized": true,
	})
}
