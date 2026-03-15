package httpserver

import (
	"github.com/gin-gonic/gin"

	"wack-backend/internal/service"
)

func (h *apiHandler) listUsers(c *gin.Context) {
	page, pageSize := parsePage(c)
	users, total, err := h.users.ListUsers(service.ListUsersInput{
		Page:     page,
		PageSize: pageSize,
		Role:     c.Query("role"),
		Status:   c.Query("status"),
		Keyword:  c.Query("keyword"),
	})
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

	user, err := h.users.CreateUser(service.CreateUserInput{
		StudentID: req.StudentID,
		RealName:  req.RealName,
		Password:  req.Password,
		Role:      req.Role,
		Status:    req.Status,
		ClassIDs:  req.ClassIDs,
	})
	if err != nil {
		fail(c, 400, "create user failed")
		return
	}
	ok(c, user)
}

func (h *apiHandler) getUser(c *gin.Context) {
	user, classes, err := h.users.GetUserWithClasses(c.Param("student_id"))
	if err != nil {
		fail(c, 404, "user not found")
		return
	}
	ok(c, gin.H{"user": user, "class_relations": classes})
}

func (h *apiHandler) updateUser(c *gin.Context) {
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
	user, classes, err := h.users.UpdateUser(current.ID, c.Param("student_id"), service.UpdateUserInput{
		StudentID: req.StudentID,
		RealName:  req.RealName,
		Role:      req.Role,
		Status:    req.Status,
		ClassIDs:  req.ClassIDs,
	})
	if err != nil {
		switch {
		case service.IsServiceError(err, service.ErrUserNotFound):
			fail(c, 404, "user not found")
		case service.IsServiceError(err, service.ErrAdminRemoveOwnRole):
			fail(c, 400, "admin cannot remove own admin role")
		case service.IsServiceError(err, service.ErrAdminFreezeSelf):
			fail(c, 400, "admin cannot freeze self")
		default:
			fail(c, 400, "update user failed")
		}
		return
	}

	ok(c, gin.H{"user": user, "class_relations": classes})
}

func (h *apiHandler) resetUserPassword(c *gin.Context) {
	current, _ := currentUser(c)

	var req struct {
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "invalid request")
		return
	}

	if err := h.users.ResetUserPassword(current.ID, c.Param("student_id"), req.NewPassword); err != nil {
		switch {
		case service.IsServiceError(err, service.ErrUserNotFound):
			fail(c, 404, "user not found")
		case service.IsServiceError(err, service.ErrAdminResetOwnPassword):
			fail(c, 400, "admin cannot reset own password here")
		default:
			fail(c, 500, "reset password failed")
		}
		return
	}
	ok(c, gin.H{})
}

func (h *apiHandler) updateUserStatus(c *gin.Context) {
	current, _ := currentUser(c)

	var req struct {
		Status int `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "invalid request")
		return
	}

	if err := h.users.UpdateUserStatus(current.ID, c.Param("student_id"), req.Status); err != nil {
		switch {
		case service.IsServiceError(err, service.ErrUserNotFound):
			fail(c, 404, "user not found")
		case service.IsServiceError(err, service.ErrAdminFreezeSelf):
			fail(c, 400, "admin cannot freeze self")
		default:
			fail(c, 400, "update status failed")
		}
		return
	}
	ok(c, gin.H{})
}

func (h *apiHandler) updateUserRole(c *gin.Context) {
	current, _ := currentUser(c)

	var req struct {
		Role int `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "invalid request")
		return
	}

	if err := h.users.UpdateUserRole(current.ID, c.Param("student_id"), req.Role); err != nil {
		switch {
		case service.IsServiceError(err, service.ErrUserNotFound):
			fail(c, 404, "user not found")
		case service.IsServiceError(err, service.ErrAdminRemoveOwnRole):
			fail(c, 400, "admin cannot remove own admin role")
		default:
			fail(c, 400, "update role failed")
		}
		return
	}
	ok(c, gin.H{})
}
