package httpserver

import (
	"time"

	"github.com/gin-gonic/gin"

	"wack-backend/internal/service"
)

type saveTermRequest struct {
	Name          string `json:"name" binding:"required,max=20"`
	TermStartDate string `json:"term_start_date" binding:"required,len=10"`
}

func (h *apiHandler) metaContext(c *gin.Context) {
	user, exists := currentUser(c)
	if !exists {
		fail(c, 401, "unauthorized")
		return
	}

	now := time.Now()
	setting, err := h.systemSettings.GetSystemSetting()
	if err != nil {
		fail(c, 500, "load context failed")
		return
	}
	latestTerm, err := h.meta.GetLatestTerm()
	if err != nil {
		fail(c, 500, "load context failed")
		return
	}

	currentWeek := 0
	var termData any = nil
	if latestTerm != nil {
		_, currentWeek, err = h.meta.TermWeeks(*latestTerm, now)
		if err != nil {
			fail(c, 500, "load context failed")
			return
		}
		termData = latestTerm
	}

	ok(c, gin.H{
		"user": gin.H{
			"id":               user.ID,
			"login_id":         user.LoginID,
			"real_name":        user.RealName,
			"role":             user.Role,
			"status":           user.Status,
			"managed_class_id": user.ManagedClassID,
		},
		"current_term":     termData,
		"current_week":     currentWeek,
		"current_schedule": currentScheduleName(setting),
		"sections":         h.meta.Sections(now),
	})
}

func (h *apiHandler) metaTerms(c *gin.Context) {
	terms, err := h.meta.ListTerms()
	if err != nil {
		fail(c, 500, "load terms failed")
		return
	}
	ok(c, gin.H{"list": terms})
}

func (h *apiHandler) metaTermWeeks(c *gin.Context) {
	termID, err := parseUintParam(c, "term_id")
	if err != nil {
		fail(c, 400, "invalid term id")
		return
	}

	term, err := h.meta.GetTerm(termID)
	if err != nil {
		if service.IsServiceError(err, service.ErrTermNotFound) {
			fail(c, 404, "term not found")
			return
		}
		fail(c, 500, "load term weeks failed")
		return
	}

	list, currentWeek, err := h.meta.TermWeeks(term, time.Now())
	if err != nil {
		fail(c, 500, "load term weeks failed")
		return
	}

	ok(c, gin.H{
		"term":         term,
		"current_week": currentWeek,
		"list":         list,
	})
}

func (h *apiHandler) metaSections(c *gin.Context) {
	now := time.Now()
	setting, err := h.systemSettings.GetSystemSetting()
	if err != nil {
		fail(c, 500, "load sections failed")
		return
	}
	ok(c, gin.H{
		"schedule": currentScheduleName(setting),
		"date":     now.Format("2006-01-02"),
		"list":     h.meta.Sections(now),
	})
}

func (h *apiHandler) adminCreateTerm(c *gin.Context) {
	var req saveTermRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "invalid term payload")
		return
	}

	term, err := h.meta.CreateTerm(req.Name, req.TermStartDate)
	if err != nil {
		if service.IsServiceError(err, service.ErrTermStartDateMustBeMonday) {
			fail(c, 400, "term start date must be monday")
			return
		}
		if service.IsServiceError(err, service.ErrInvalidInput) {
			fail(c, 400, "invalid term payload")
			return
		}
		fail(c, 500, "create term failed")
		return
	}
	ok(c, term)
}

func (h *apiHandler) adminUpdateTerm(c *gin.Context) {
	termID, err := parseUintParam(c, "term_id")
	if err != nil {
		fail(c, 400, "invalid term id")
		return
	}

	var req saveTermRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "invalid term payload")
		return
	}

	term, err := h.meta.UpdateTerm(termID, req.TermStartDate)
	if err != nil {
		switch {
		case service.IsServiceError(err, service.ErrTermStartDateMustBeMonday):
			fail(c, 400, "term start date must be monday")
		case service.IsServiceError(err, service.ErrInvalidInput):
			fail(c, 400, "invalid term payload")
		case service.IsServiceError(err, service.ErrTermNotFound):
			fail(c, 404, "term not found")
		default:
			fail(c, 500, "update term failed")
		}
		return
	}
	ok(c, term)
}
