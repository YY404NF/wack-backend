package httpserver

import (
	"github.com/gin-gonic/gin"

	"wack-backend/internal/httpserver/dto"
	"wack-backend/internal/model"
	"wack-backend/internal/service"
)

func (h *apiHandler) listUsers(c *gin.Context) {
	page, pageSize := parsePage(c)
	users, total, err := h.users.ListUsers(service.ListUsersInput{
		Page:             page,
		PageSize:         pageSize,
		Role:             c.Query("role"),
		Status:           c.Query("status"),
		Keyword:          c.Query("keyword"),
		LoginID:          c.Query("login_id"),
		RealName:         c.Query("real_name"),
		ManagedClassName: c.Query("managed_class_name"),
	})
	if err != nil {
		fail(c, 500, "list users failed")
		return
	}

	ok(c, pageResult[model.User]{Items: users, Page: page, PageSize: pageSize, Total: total})
}

func (h *apiHandler) createUser(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "invalid request")
		return
	}

	user, err := h.users.CreateUser(service.CreateUserInput{
		LoginID:        req.LoginID,
		RealName:       req.RealName,
		Password:       req.Password,
		Role:           req.Role,
		Status:         req.Status,
		ManagedClassID: req.ManagedClassID,
	})
	if err != nil {
		switch {
		case service.IsServiceError(err, service.ErrInvalidInput):
			fail(c, 400, "invalid request")
		default:
			fail(c, 400, "create user failed")
		}
		return
	}
	ok(c, user)
}

func (h *apiHandler) getUser(c *gin.Context) {
	user, err := h.users.GetUser(c.Param("student_id"))
	if err != nil {
		fail(c, 404, "user not found")
		return
	}
	ok(c, user)
}

func (h *apiHandler) updateUser(c *gin.Context) {
	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "invalid request")
		return
	}

	current, _ := currentUser(c)
	user, err := h.users.UpdateUser(current.ID, c.Param("student_id"), service.UpdateUserInput{
		LoginID:        req.LoginID,
		RealName:       req.RealName,
		Role:           req.Role,
		Status:         req.Status,
		ManagedClassID: req.ManagedClassID,
	})
	if err != nil {
		switch {
		case service.IsServiceError(err, service.ErrUserNotFound):
			fail(c, 404, "user not found")
		case service.IsServiceError(err, service.ErrInvalidInput):
			fail(c, 400, "invalid request")
		case service.IsServiceError(err, service.ErrAdminRemoveOwnRole):
			fail(c, 400, "admin cannot remove own admin role")
		case service.IsServiceError(err, service.ErrAdminFreezeSelf):
			fail(c, 400, "admin cannot freeze self")
		default:
			fail(c, 400, "update user failed")
		}
		return
	}
	if user.Status == model.UserStatusFrozen {
		if err := h.sessions.DeleteAllUserSessions(c.Request.Context(), user.ID); err != nil {
			fail(c, 500, "clear sessions failed")
			return
		}
	}

	ok(c, user)
}

func (h *apiHandler) resetUserPassword(c *gin.Context) {
	current, _ := currentUser(c)

	var req dto.ResetUserPasswordRequest
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
	targetUser, err := h.users.GetUser(c.Param("student_id"))
	if err != nil {
		fail(c, 404, "user not found")
		return
	}
	if err := h.sessions.DeleteAllUserSessions(c.Request.Context(), targetUser.ID); err != nil {
		fail(c, 500, "clear sessions failed")
		return
	}
	ok(c, gin.H{})
}

func (h *apiHandler) updateUserStatus(c *gin.Context) {
	current, _ := currentUser(c)

	var req dto.UpdateUserStatusRequest
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
	if req.Status == model.UserStatusFrozen {
		targetUser, err := h.users.GetUser(c.Param("student_id"))
		if err != nil {
			fail(c, 404, "user not found")
			return
		}
		if err := h.sessions.DeleteAllUserSessions(c.Request.Context(), targetUser.ID); err != nil {
			fail(c, 500, "clear sessions failed")
			return
		}
	}
	ok(c, gin.H{})
}

func (h *apiHandler) updateUserRole(c *gin.Context) {
	current, _ := currentUser(c)

	var req dto.UpdateUserRoleRequest
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
