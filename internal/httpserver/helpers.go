package httpserver

import (
	"errors"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"wack-backend/internal/model"
	"wack-backend/internal/service"
)

type apiHandler struct {
	db         *gorm.DB
	users      *service.UserService
	classes    *service.ClassService
	courses    *service.CourseService
	freeTimes  *service.FreeTimeService
	attendance *service.AttendanceService
	logs       *service.LogService
}

func newAPIHandler(db *gorm.DB) *apiHandler {
	return &apiHandler{
		db:         db,
		users:      service.NewUserService(db),
		classes:    service.NewClassService(db),
		courses:    service.NewCourseService(db),
		freeTimes:  service.NewFreeTimeService(db),
		attendance: service.NewAttendanceService(db),
		logs:       service.NewLogService(db),
	}
}

type pageResult[T any] struct {
	Items    []T   `json:"items"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
	Total    int64 `json:"total"`
}

func parsePage(c *gin.Context) (int, int) {
	page := 1
	pageSize := 20
	if value, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil && value > 0 {
		page = value
	}
	if value, err := strconv.Atoi(c.DefaultQuery("page_size", "20")); err == nil && value > 0 && value <= 100 {
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

func (h *apiHandler) findUserByStudentID(studentID string) (model.User, error) {
	var user model.User
	err := h.db.First(&user, "student_id = ?", studentID).Error
	return user, err
}

func (h *apiHandler) findUsersByStudentIDs(studentIDs []string) ([]model.User, error) {
	var users []model.User
	if len(studentIDs) == 0 {
		return users, nil
	}
	err := h.db.Where("student_id IN ?", studentIDs).Find(&users).Error
	return users, err
}

func sectionEndTime(section int) (int, int, error) {
	switch section {
	case 1:
		return 9, 35, nil
	case 2:
		return 11, 35, nil
	case 3:
		return 15, 35, nil
	case 4:
		return 17, 35, nil
	case 5:
		return 21, 35, nil
	default:
		return 0, 0, errors.New("invalid section")
	}
}

func (h *apiHandler) attendanceDeadline(session model.CourseSession, now time.Time) (time.Time, error) {
	currentWeekday := int(now.Weekday())
	if currentWeekday == 0 {
		currentWeekday = 7
	}
	weekdayDelta := session.Weekday - currentWeekday
	if weekdayDelta < 0 {
		weekdayDelta += 7
	}
	baseDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, weekdayDelta)
	hour, minute, err := sectionEndTime(session.Section)
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(baseDate.Year(), baseDate.Month(), baseDate.Day(), hour, minute, 0, 0, now.Location()).Add(30 * time.Minute), nil
}

func (h *apiHandler) withinDeadline(session model.CourseSession, now time.Time) bool {
	deadline, err := h.attendanceDeadline(session, now)
	if err != nil {
		return false
	}
	return now.Before(deadline)
}
