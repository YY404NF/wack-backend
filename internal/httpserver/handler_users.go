package httpserver

import (
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"wack-backend/internal/model"
)

func (h *apiHandler) listUsers(c *gin.Context) {
	page, pageSize := parsePage(c)
	query := h.db.Model(&model.User{})
	if role := c.Query("role"); role != "" {
		query = query.Where("role = ?", role)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if keyword := strings.TrimSpace(c.Query("keyword")); keyword != "" {
		query = query.Where("student_id LIKE ? OR real_name LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	var users []model.User
	total, err := paginate(query.Order("created_at DESC"), page, pageSize, &users)
	if err != nil {
		fail(c, 500, "list users failed")
		return
	}

	ok(c, pageResult{Items: users, Page: page, PageSize: pageSize, Total: total})
}

func (h *apiHandler) createUser(c *gin.Context) {
	var req struct {
		StudentID string   `json:"student_id" binding:"required"`
		RealName  string   `json:"real_name" binding:"required"`
		Password  string   `json:"password" binding:"required,min=6"`
		Role      int      `json:"role" binding:"required"`
		Status    int      `json:"status"`
		ClassIDs  []uint64 `json:"class_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "invalid request")
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		fail(c, 500, "hash password failed")
		return
	}
	user := model.User{
		StudentID:    req.StudentID,
		PasswordHash: string(hash),
		RealName:     req.RealName,
		Role:         req.Role,
		Status:       req.Status,
	}
	if user.Status == 0 {
		user.Status = model.UserStatusActive
	}

	err = h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&user).Error; err != nil {
			return err
		}
		if len(req.ClassIDs) > 0 {
			var relations []model.UserClass
			for _, classID := range req.ClassIDs {
				relations = append(relations, model.UserClass{UserID: user.ID, ClassID: classID})
			}
			return tx.Create(&relations).Error
		}
		return nil
	})
	if err != nil {
		fail(c, 400, "create user failed")
		return
	}
	ok(c, user)
}

func (h *apiHandler) getUser(c *gin.Context) {
	var user model.User
	if err := h.db.First(&user, "student_id = ?", c.Param("student_id")).Error; err != nil {
		fail(c, 404, "user not found")
		return
	}
	var classes []model.UserClass
	_ = h.db.Where("user_id = ?", user.ID).Find(&classes).Error
	ok(c, gin.H{"user": user, "class_relations": classes})
}

func (h *apiHandler) updateUser(c *gin.Context) {
	studentID := c.Param("student_id")
	var user model.User
	if err := h.db.First(&user, "student_id = ?", studentID).Error; err != nil {
		fail(c, 404, "user not found")
		return
	}

	var req struct {
		StudentID string   `json:"student_id" binding:"required"`
		RealName  string   `json:"real_name" binding:"required"`
		Role      int      `json:"role" binding:"required"`
		Status    int      `json:"status" binding:"required"`
		ClassIDs  []uint64 `json:"class_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "invalid request")
		return
	}

	current, _ := currentUser(c)
	if current.ID == user.ID && req.Role != model.RoleAdmin {
		fail(c, 400, "admin cannot remove own admin role")
		return
	}
	if current.ID == user.ID && req.Status != model.UserStatusActive {
		fail(c, 400, "admin cannot freeze self")
		return
	}

	err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&user).Updates(map[string]interface{}{
			"student_id": req.StudentID,
			"real_name":  req.RealName,
			"role":       req.Role,
			"status":     req.Status,
		}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", user.ID).Delete(&model.UserClass{}).Error; err != nil {
			return err
		}
		if len(req.ClassIDs) == 0 {
			return nil
		}
		var relations []model.UserClass
		for _, classID := range req.ClassIDs {
			relations = append(relations, model.UserClass{UserID: user.ID, ClassID: classID})
		}
		return tx.Create(&relations).Error
	})
	if err != nil {
		fail(c, 400, "update user failed")
		return
	}

	h.getUser(c)
}

func (h *apiHandler) resetUserPassword(c *gin.Context) {
	current, _ := currentUser(c)
	targetUser, err := h.findUserByStudentID(c.Param("student_id"))
	if err != nil {
		fail(c, 404, "user not found")
		return
	}
	if current.ID == targetUser.ID {
		fail(c, 400, "admin cannot reset own password here")
		return
	}

	var req struct {
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "invalid request")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		fail(c, 500, "hash password failed")
		return
	}
	if err := h.db.Model(&targetUser).Update("password_hash", string(hash)).Error; err != nil {
		fail(c, 500, "reset password failed")
		return
	}
	ok(c, gin.H{})
}

func (h *apiHandler) updateUserStatus(c *gin.Context) {
	current, _ := currentUser(c)
	targetID := c.Param("student_id")
	targetUser, err := h.findUserByStudentID(targetID)
	if err != nil {
		fail(c, 404, "user not found")
		return
	}
	if current.ID == targetUser.ID {
		fail(c, 400, "admin cannot freeze self")
		return
	}

	var req struct {
		Status int `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "invalid request")
		return
	}
	if err := h.db.Model(&model.User{}).Where("id = ?", targetUser.ID).Update("status", req.Status).Error; err != nil {
		fail(c, 400, "update status failed")
		return
	}
	ok(c, gin.H{})
}

func (h *apiHandler) updateUserRole(c *gin.Context) {
	current, _ := currentUser(c)
	targetID := c.Param("student_id")
	targetUser, err := h.findUserByStudentID(targetID)
	if err != nil {
		fail(c, 404, "user not found")
		return
	}
	if current.ID == targetUser.ID {
		fail(c, 400, "admin cannot remove own admin role")
		return
	}

	var req struct {
		Role int `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "invalid request")
		return
	}
	if err := h.db.Model(&model.User{}).Where("id = ?", targetUser.ID).Update("role", req.Role).Error; err != nil {
		fail(c, 400, "update role failed")
		return
	}
	ok(c, gin.H{})
}
