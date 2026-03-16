package httpserver

import (
	"github.com/gin-gonic/gin"

	"wack-backend/internal/httpserver/dto"
	"wack-backend/internal/model"
	"wack-backend/internal/query"
)

func (h *apiHandler) listFreeTimes(c *gin.Context) {
	page, pageSize := parsePage(c)
	user, _ := currentUser(c)
	items, total, err := h.freeTimes.ListFreeTimes(c.Query("term"), c.Query("student_id"), user, page, pageSize)
	if err != nil {
		fail(c, 500, "list free times failed")
		return
	}
	ok(c, pageResult[query.FreeTimeItem]{Items: items, Page: page, PageSize: pageSize, Total: total})
}

func (h *apiHandler) createFreeTime(c *gin.Context) {
	user, _ := currentUser(c)
	var req dto.FreeTimeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "invalid request")
		return
	}

	targetUserID := user.ID
	if user.Role == model.RoleAdmin && req.StudentID != "" {
		target, err := h.findUserByStudentID(req.StudentID)
		if err != nil {
			fail(c, 404, "user not found")
			return
		}
		targetUserID = target.ID
	}

	item, err := h.freeTimes.CreateFreeTime(model.StudentFreeTime{
		Term:      req.Term,
		UserID:    targetUserID,
		Weekday:   req.Weekday,
		Section:   req.Section,
		FreeWeeks: req.FreeWeeks,
	})
	if err != nil {
		fail(c, 400, "create free time failed")
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
	if user.Role == model.RoleAdmin && req.StudentID != "" {
		target, err := h.findUserByStudentID(req.StudentID)
		if err != nil {
			fail(c, 404, "user not found")
			return
		}
		targetUserID = target.ID
	} else if user.Role == model.RoleStudent {
		targetUserID = user.ID
	}

	if err := h.freeTimes.UpdateFreeTime(id, req.Term, targetUserID, req.Weekday, req.Section, req.FreeWeeks); err != nil {
		fail(c, 400, "update free time failed")
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
