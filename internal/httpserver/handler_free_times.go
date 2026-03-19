package httpserver

import (
	"github.com/gin-gonic/gin"

	"wack-backend/internal/httpserver/dto"
	"wack-backend/internal/model"
	"wack-backend/internal/query"
	"wack-backend/internal/service"
)

func (h *apiHandler) listFreeTimes(c *gin.Context) {
	page, pageSize := parsePage(c)
	user, _ := currentUser(c)
	if user.Role == model.RoleCommissioner {
		fail(c, 403, "forbidden")
		return
	}
	items, total, err := h.freeTimes.ListFreeTimes(c.Query("term"), c.Query("login_id"), user, page, pageSize)
	if err != nil {
		fail(c, 500, "list free times failed")
		return
	}
	ok(c, pageResult[query.FreeTimeItem]{Items: items, Page: page, PageSize: pageSize, Total: total})
}

func (h *apiHandler) listFreeTimeEditor(c *gin.Context) {
	user, _ := currentUser(c)
	if user.Role == model.RoleCommissioner {
		fail(c, 403, "forbidden")
		return
	}
	items, err := h.freeTimes.ListFreeTimeEditor(c.Query("term"), c.Query("login_id"), user)
	if err != nil {
		fail(c, 500, "list free time editor failed")
		return
	}
	ok(c, items)
}

func (h *apiHandler) createFreeTime(c *gin.Context) {
	user, _ := currentUser(c)
	if user.Role == model.RoleCommissioner {
		fail(c, 403, "forbidden")
		return
	}
	var req dto.FreeTimeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "invalid request")
		return
	}

	targetUserID := user.ID
	if user.Role == model.RoleAdmin && req.LoginID != "" {
		target, err := h.findUserByLoginID(req.LoginID)
		if err != nil {
			fail(c, 404, "user not found")
			return
		}
		targetUserID = target.ID
	}

	item, err := h.freeTimes.CreateFreeTime(req.Term, targetUserID, req.Weekday, req.Section, req.FreeWeeks)
	if err != nil {
		switch {
		case service.IsServiceError(err, service.ErrUserNotFound):
			fail(c, 404, "user not found")
		case service.IsServiceError(err, service.ErrInvalidInput):
			fail(c, 400, "invalid request")
		default:
			fail(c, 400, "create free time failed")
		}
		return
	}
	ok(c, item)
}

func (h *apiHandler) updateFreeTime(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}

	user, _ := currentUser(c)
	if user.Role == model.RoleCommissioner {
		fail(c, 403, "forbidden")
		return
	}
	item, err := h.freeTimes.GetFreeTime(id)
	if err != nil {
		fail(c, 404, "free time not found")
		return
	}
	if user.Role == model.RoleStudent && item.UserID != user.ID {
		fail(c, 403, "cannot modify other user's free time")
		return
	}

	var req dto.FreeTimeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "invalid request")
		return
	}

	targetUserID := item.UserID
	if user.Role == model.RoleAdmin && req.LoginID != "" {
		target, err := h.findUserByLoginID(req.LoginID)
		if err != nil {
			fail(c, 404, "user not found")
			return
		}
		targetUserID = target.ID
	} else if user.Role == model.RoleStudent {
		targetUserID = user.ID
	}

	if err := h.freeTimes.UpdateFreeTime(id, req.Term, targetUserID, req.Weekday, req.Section, req.FreeWeeks); err != nil {
		switch {
		case service.IsServiceError(err, service.ErrFreeTimeNotFound):
			fail(c, 404, "free time not found")
		case service.IsServiceError(err, service.ErrUserNotFound):
			fail(c, 404, "user not found")
		case service.IsServiceError(err, service.ErrInvalidInput):
			fail(c, 400, "invalid request")
		default:
			fail(c, 400, "update free time failed")
		}
		return
	}
	ok(c, gin.H{})
}

func (h *apiHandler) deleteFreeTime(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}

	user, _ := currentUser(c)
	if user.Role == model.RoleCommissioner {
		fail(c, 403, "forbidden")
		return
	}
	item, err := h.freeTimes.GetFreeTime(id)
	if err != nil {
		fail(c, 404, "free time not found")
		return
	}
	if user.Role == model.RoleStudent && item.UserID != user.ID {
		fail(c, 403, "cannot delete other user's free time")
		return
	}

	if err := h.freeTimes.DeleteFreeTime(id); err != nil {
		fail(c, 500, "delete free time failed")
		return
	}
	ok(c, gin.H{})
}
