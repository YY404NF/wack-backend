package httpserver

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"wack-backend/internal/model"
	"wack-backend/internal/service"
)

type apiHandler struct {
	db             *gorm.DB
	users          *service.UserService
	students       *service.StudentService
	sessions       *service.SessionService
	classes        *service.ClassService
	courses        *service.CourseService
	freeTimes      *service.FreeTimeService
	systemSettings *service.SystemSettingService
	meta           *service.MetaService
	attendance     *service.AttendanceService
	logs           *service.LogService
}

func newAPIHandler(db *gorm.DB, sessions *service.SessionService) *apiHandler {
	return &apiHandler{
		db:             db,
		users:          service.NewUserService(db),
		students:       service.NewStudentService(db),
		sessions:       sessions,
		classes:        service.NewClassService(db),
		courses:        service.NewCourseService(db),
		freeTimes:      service.NewFreeTimeService(db),
		systemSettings: service.NewSystemSettingService(db),
		meta:           service.NewMetaService(db),
		attendance:     service.NewAttendanceService(db),
		logs:           service.NewLogService(db),
	}
}

type pageResult[T any] struct {
	Items       []T     `json:"items"`
	Page        int     `json:"page"`
	PageSize    int     `json:"page_size"`
	Total       int64   `json:"total"`
	FocusFound  *bool   `json:"focus_found,omitempty"`
	FocusPage   *int    `json:"focus_page,omitempty"`
	FocusRowKey *uint64 `json:"focus_row_key,omitempty"`
}

func parsePage(c *gin.Context) (int, int) {
	page := 1
	pageSize := 20
	if value, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil && value > 0 {
		page = value
	}
	if value, err := strconv.Atoi(c.DefaultQuery("page_size", "20")); err == nil && value > 0 && value <= 1000 {
		pageSize = value
	}
	return page, pageSize
}

func paginate[T any](query *gorm.DB, page, pageSize int, out *[]T) (int64, error) {
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return 0, err
	}
	if err := query.Offset((page - 1) * pageSize).Limit(pageSize).Find(out).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func parseUintParam(c *gin.Context, name string) (uint64, error) {
	value, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil {
		return 0, errors.New("invalid path param")
	}
	return value, nil
}

func parseUintQuery(c *gin.Context, name string) (uint64, error) {
	value := c.Query(name)
	if value == "" {
		return 0, nil
	}
	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, errors.New("invalid query param")
	}
	return parsed, nil
}

func (h *apiHandler) findUserByLoginID(loginID string) (model.User, error) {
	var user model.User
	err := h.db.First(&user, "login_id = ?", loginID).Error
	return user, err
}

func (h *apiHandler) findUsersByLoginIDs(loginIDs []string) ([]model.User, error) {
	var users []model.User
	if len(loginIDs) == 0 {
		return users, nil
	}
	err := h.db.Where("login_id IN ?", loginIDs).Find(&users).Error
	return users, err
}
