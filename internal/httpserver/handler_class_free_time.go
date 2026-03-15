package httpserver

import (
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"wack-backend/internal/model"
)

func (h *apiHandler) listClasses(c *gin.Context) {
	page, pageSize := parsePage(c)
	query := h.db.Model(&model.Class{})
	var classes []model.Class
	total, err := paginate(query.Order("id DESC"), page, pageSize, &classes)
	if err != nil {
		fail(c, 500, "list classes failed")
		return
	}
	ok(c, pageResult{Items: classes, Page: page, PageSize: pageSize, Total: total})
}

func (h *apiHandler) createClass(c *gin.Context) {
	var class model.Class
	if err := c.ShouldBindJSON(&class); err != nil {
		fail(c, 400, "invalid request")
		return
	}
	if err := h.db.Create(&class).Error; err != nil {
		fail(c, 400, "create class failed")
		return
	}
	ok(c, class)
}

func (h *apiHandler) getClass(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	var class model.Class
	if err := h.db.First(&class, id).Error; err != nil {
		fail(c, 404, "class not found")
		return
	}
	ok(c, class)
}

func (h *apiHandler) updateClass(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	var class model.Class
	if err := h.db.First(&class, id).Error; err != nil {
		fail(c, 404, "class not found")
		return
	}
	var req model.Class
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "invalid request")
		return
	}
	if err := h.db.Model(&class).Updates(map[string]interface{}{
		"class_code": req.ClassCode,
		"class_name": req.ClassName,
		"grade":      req.Grade,
		"major_name": req.MajorName,
	}).Error; err != nil {
		fail(c, 400, "update class failed")
		return
	}
	h.getClass(c)
}

func (h *apiHandler) deleteClass(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("class_id = ?", id).Delete(&model.UserClass{}).Error; err != nil {
			return err
		}
		if err := tx.Where("class_id = ?", id).Delete(&model.CourseClass{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.Class{}, id).Error
	}); err != nil {
		fail(c, 400, "delete class failed")
		return
	}
	ok(c, gin.H{})
}

func (h *apiHandler) getClassStudents(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	var users []model.User
	if err := h.db.Table("user").
		Joins("JOIN user_class ON user_class.user_id = user.id").
		Where("user_class.class_id = ?", id).
		Find(&users).Error; err != nil {
		fail(c, 500, "get class students failed")
		return
	}
	ok(c, users)
}

func (h *apiHandler) listFreeTimes(c *gin.Context) {
	page, pageSize := parsePage(c)
	query := h.db.Table("student_free_time").
		Joins("JOIN user ON user.id = student_free_time.user_id")
	user, _ := currentUser(c)
	if term := c.Query("term"); term != "" {
		query = query.Where("student_free_time.term = ?", term)
	}
	if studentID := c.Query("student_id"); studentID != "" {
		query = query.Where("user.student_id = ?", studentID)
	} else if user.Role == model.RoleStudent {
		query = query.Where("student_free_time.user_id = ?", user.ID)
	}
	type item struct {
		ID        uint64    `json:"id"`
		Term      string    `json:"term"`
		UserID    uint64    `json:"user_id"`
		StudentID string    `json:"student_id"`
		Weekday   int       `json:"weekday"`
		Section   int       `json:"section"`
		FreeWeeks string    `json:"free_weeks"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}
	var items []item
	total, err := paginate(query.Order("student_free_time.id DESC"), page, pageSize, &items)
	if err != nil {
		fail(c, 500, "list free times failed")
		return
	}
	ok(c, pageResult{Items: items, Page: page, PageSize: pageSize, Total: total})
}

func (h *apiHandler) createFreeTime(c *gin.Context) {
	user, _ := currentUser(c)
	var req struct {
		Term      string `json:"term" binding:"required"`
		StudentID string `json:"student_id"`
		Weekday   int    `json:"weekday" binding:"required"`
		Section   int    `json:"section" binding:"required"`
		FreeWeeks string `json:"free_weeks" binding:"required"`
	}
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
	item := model.StudentFreeTime{Term: req.Term, UserID: targetUserID, Weekday: req.Weekday, Section: req.Section, FreeWeeks: req.FreeWeeks}
	if err := h.db.Create(&item).Error; err != nil {
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
	var item model.StudentFreeTime
	if err := h.db.First(&item, id).Error; err != nil {
		fail(c, 404, "free time not found")
		return
	}
	if user.Role == model.RoleStudent && item.UserID != user.ID {
		fail(c, 403, "cannot modify other user's free time")
		return
	}
	var req struct {
		Term      string `json:"term" binding:"required"`
		StudentID string `json:"student_id"`
		Weekday   int    `json:"weekday" binding:"required"`
		Section   int    `json:"section" binding:"required"`
		FreeWeeks string `json:"free_weeks" binding:"required"`
	}
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
	if err := h.db.Model(&item).Updates(map[string]interface{}{
		"term":       req.Term,
		"user_id":    targetUserID,
		"weekday":    req.Weekday,
		"section":    req.Section,
		"free_weeks": req.FreeWeeks,
	}).Error; err != nil {
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
	var item model.StudentFreeTime
	if err := h.db.First(&item, id).Error; err != nil {
		fail(c, 404, "free time not found")
		return
	}
	if user.Role == model.RoleStudent && item.UserID != user.ID {
		fail(c, 403, "cannot delete other user's free time")
		return
	}
	if err := h.db.Delete(&item).Error; err != nil {
		fail(c, 500, "delete free time failed")
		return
	}
	ok(c, gin.H{})
}
