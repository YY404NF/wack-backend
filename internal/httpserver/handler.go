package httpserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"wack-backend/internal/model"
)

type apiHandler struct {
	db *gorm.DB
}

func newAPIHandler(db *gorm.DB) *apiHandler {
	return &apiHandler{db: db}
}

type pageResult struct {
	Items    interface{} `json:"items"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
	Total    int64       `json:"total"`
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

func paginate(query *gorm.DB, page, pageSize int, out interface{}) (int64, error) {
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
				relations = append(relations, model.UserClass{StudentID: user.StudentID, ClassID: classID})
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
	_ = h.db.Where("student_id = ?", user.StudentID).Find(&classes).Error
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
		RealName string   `json:"real_name" binding:"required"`
		Role     int      `json:"role" binding:"required"`
		Status   int      `json:"status" binding:"required"`
		ClassIDs []uint64 `json:"class_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "invalid request")
		return
	}

	current, _ := currentUser(c)
	if current.StudentID == studentID && req.Role != model.RoleAdmin {
		fail(c, 400, "admin cannot remove own admin role")
		return
	}
	if current.StudentID == studentID && req.Status != model.UserStatusActive {
		fail(c, 400, "admin cannot freeze self")
		return
	}

	err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&user).Updates(map[string]interface{}{
			"real_name": req.RealName,
			"role":      req.Role,
			"status":    req.Status,
		}).Error; err != nil {
			return err
		}
		if err := tx.Where("student_id = ?", studentID).Delete(&model.UserClass{}).Error; err != nil {
			return err
		}
		if len(req.ClassIDs) == 0 {
			return nil
		}
		var relations []model.UserClass
		for _, classID := range req.ClassIDs {
			relations = append(relations, model.UserClass{StudentID: studentID, ClassID: classID})
		}
		return tx.Create(&relations).Error
	})
	if err != nil {
		fail(c, 400, "update user failed")
		return
	}

	h.getUser(c)
}

func (h *apiHandler) updateUserStatus(c *gin.Context) {
	current, _ := currentUser(c)
	targetID := c.Param("student_id")
	if current.StudentID == targetID {
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
	if err := h.db.Model(&model.User{}).Where("student_id = ?", targetID).Update("status", req.Status).Error; err != nil {
		fail(c, 400, "update status failed")
		return
	}
	ok(c, gin.H{})
}

func (h *apiHandler) updateUserRole(c *gin.Context) {
	current, _ := currentUser(c)
	targetID := c.Param("student_id")
	if current.StudentID == targetID {
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
	if err := h.db.Model(&model.User{}).Where("student_id = ?", targetID).Update("role", req.Role).Error; err != nil {
		fail(c, 400, "update role failed")
		return
	}
	ok(c, gin.H{})
}

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
		Joins("JOIN user_class ON user_class.student_id = user.student_id").
		Where("user_class.class_id = ?", id).
		Find(&users).Error; err != nil {
		fail(c, 500, "get class students failed")
		return
	}
	ok(c, users)
}

func (h *apiHandler) listFreeTimes(c *gin.Context) {
	page, pageSize := parsePage(c)
	query := h.db.Model(&model.StudentFreeTime{})
	user, _ := currentUser(c)
	if term := c.Query("term"); term != "" {
		query = query.Where("term = ?", term)
	}
	if studentID := c.Query("student_id"); studentID != "" {
		query = query.Where("student_id = ?", studentID)
	} else if user.Role == model.RoleStudent {
		query = query.Where("student_id = ?", user.StudentID)
	}
	var items []model.StudentFreeTime
	total, err := paginate(query.Order("id DESC"), page, pageSize, &items)
	if err != nil {
		fail(c, 500, "list free times failed")
		return
	}
	ok(c, pageResult{Items: items, Page: page, PageSize: pageSize, Total: total})
}

func (h *apiHandler) createFreeTime(c *gin.Context) {
	user, _ := currentUser(c)
	var item model.StudentFreeTime
	if err := c.ShouldBindJSON(&item); err != nil {
		fail(c, 400, "invalid request")
		return
	}
	if user.Role == model.RoleStudent {
		item.StudentID = user.StudentID
	}
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
	if user.Role == model.RoleStudent && item.StudentID != user.StudentID {
		fail(c, 403, "cannot modify other user's free time")
		return
	}
	var req model.StudentFreeTime
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "invalid request")
		return
	}
	if user.Role == model.RoleStudent {
		req.StudentID = user.StudentID
	}
	if err := h.db.Model(&item).Updates(map[string]interface{}{
		"term":       req.Term,
		"student_id": req.StudentID,
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
	if user.Role == model.RoleStudent && item.StudentID != user.StudentID {
		fail(c, 403, "cannot delete other user's free time")
		return
	}
	if err := h.db.Delete(&item).Error; err != nil {
		fail(c, 500, "delete free time failed")
		return
	}
	ok(c, gin.H{})
}

func (h *apiHandler) listCourses(c *gin.Context) {
	page, pageSize := parsePage(c)
	query := h.db.Model(&model.Course{})
	if term := c.Query("term"); term != "" {
		query = query.Where("term = ?", term)
	}
	if teacher := c.Query("teacher_name"); teacher != "" {
		query = query.Where("teacher_name LIKE ?", "%"+teacher+"%")
	}
	if keyword := strings.TrimSpace(c.Query("keyword")); keyword != "" {
		query = query.Where("course_name LIKE ?", "%"+keyword+"%")
	}
	var items []model.Course
	total, err := paginate(query.Order("id DESC"), page, pageSize, &items)
	if err != nil {
		fail(c, 500, "list courses failed")
		return
	}
	ok(c, pageResult{Items: items, Page: page, PageSize: pageSize, Total: total})
}

func (h *apiHandler) createCourse(c *gin.Context) {
	var course model.Course
	if err := c.ShouldBindJSON(&course); err != nil {
		fail(c, 400, "invalid request")
		return
	}
	if err := h.db.Create(&course).Error; err != nil {
		fail(c, 400, "create course failed")
		return
	}
	ok(c, course)
}

func (h *apiHandler) getCourse(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	var course model.Course
	if err := h.db.First(&course, id).Error; err != nil {
		fail(c, 404, "course not found")
		return
	}
	var students []model.CourseStudent
	var classes []model.CourseClass
	var sessions []model.CourseSession
	_ = h.db.Where("course_id = ?", id).Find(&students).Error
	_ = h.db.Where("course_id = ?", id).Find(&classes).Error
	_ = h.db.Where("course_id = ?", id).Order("session_no ASC").Find(&sessions).Error
	ok(c, gin.H{
		"course":   course,
		"students": students,
		"classes":  classes,
		"sessions": sessions,
	})
}

func (h *apiHandler) updateCourse(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	var course model.Course
	if err := h.db.First(&course, id).Error; err != nil {
		fail(c, 404, "course not found")
		return
	}
	var req model.Course
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "invalid request")
		return
	}
	if err := h.db.Model(&course).Updates(map[string]interface{}{
		"term":                     req.Term,
		"course_name":              req.CourseName,
		"teacher_name":             req.TeacherName,
		"attendance_student_count": req.AttendanceStudentCount,
	}).Error; err != nil {
		fail(c, 400, "update course failed")
		return
	}
	h.getCourse(c)
}

func (h *apiHandler) deleteCourse(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	err = h.db.Transaction(func(tx *gorm.DB) error {
		var sessionIDs []uint64
		if err := tx.Model(&model.CourseSession{}).Where("course_id = ?", id).Pluck("id", &sessionIDs).Error; err != nil {
			return err
		}
		if len(sessionIDs) > 0 {
			var checkIDs []uint64
			if err := tx.Model(&model.AttendanceCheck{}).Where("course_session_id IN ?", sessionIDs).Pluck("id", &checkIDs).Error; err != nil {
				return err
			}
			if len(checkIDs) > 0 {
				if err := tx.Where("attendance_check_id IN ?", checkIDs).Delete(&model.AttendanceDetailLog{}).Error; err != nil {
					return err
				}
				if err := tx.Where("attendance_check_id IN ?", checkIDs).Delete(&model.AttendanceDetail{}).Error; err != nil {
					return err
				}
				if err := tx.Where("id IN ?", checkIDs).Delete(&model.AttendanceCheck{}).Error; err != nil {
					return err
				}
			}
			if err := tx.Where("id IN ?", sessionIDs).Delete(&model.CourseSession{}).Error; err != nil {
				return err
			}
		}
		if err := tx.Where("course_id = ?", id).Delete(&model.CourseStudent{}).Error; err != nil {
			return err
		}
		if err := tx.Where("course_id = ?", id).Delete(&model.CourseClass{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.Course{}, id).Error
	})
	if err != nil {
		fail(c, 400, "delete course failed")
		return
	}
	ok(c, gin.H{})
}

func (h *apiHandler) replaceCourseStudents(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	var req struct {
		StudentIDs []string `json:"student_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "invalid request")
		return
	}
	err = h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("course_id = ?", id).Delete(&model.CourseStudent{}).Error; err != nil {
			return err
		}
		var relations []model.CourseStudent
		for _, studentID := range req.StudentIDs {
			relations = append(relations, model.CourseStudent{CourseID: id, StudentID: studentID})
		}
		if len(relations) > 0 {
			if err := tx.Create(&relations).Error; err != nil {
				return err
			}
		}
		return tx.Model(&model.Course{}).Where("id = ?", id).Update("attendance_student_count", len(req.StudentIDs)).Error
	})
	if err != nil {
		fail(c, 400, "replace course students failed")
		return
	}
	ok(c, gin.H{})
}

func (h *apiHandler) replaceCourseClasses(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	var req struct {
		ClassIDs []uint64 `json:"class_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "invalid request")
		return
	}
	err = h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("course_id = ?", id).Delete(&model.CourseClass{}).Error; err != nil {
			return err
		}
		var relations []model.CourseClass
		for _, classID := range req.ClassIDs {
			relations = append(relations, model.CourseClass{CourseID: id, ClassID: classID})
		}
		if len(relations) > 0 {
			return tx.Create(&relations).Error
		}
		return nil
	})
	if err != nil {
		fail(c, 400, "replace course classes failed")
		return
	}
	ok(c, gin.H{})
}

func (h *apiHandler) replaceCourseSessions(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	var req struct {
		Sessions []model.CourseSession `json:"sessions"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "invalid request")
		return
	}
	err = h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("course_id = ?", id).Delete(&model.CourseSession{}).Error; err != nil {
			return err
		}
		for i := range req.Sessions {
			req.Sessions[i].ID = 0
			req.Sessions[i].CourseID = id
		}
		if len(req.Sessions) > 0 {
			return tx.Create(&req.Sessions).Error
		}
		return nil
	})
	if err != nil {
		fail(c, 400, "replace course sessions failed")
		return
	}
	ok(c, gin.H{})
}

func (h *apiHandler) adminCourseCalendar(c *gin.Context) {
	query := h.db.Model(&model.CourseSession{}).
		Joins("JOIN course ON course.id = course_session.course_id")
	if weekNo := c.Query("week_no"); weekNo != "" {
		query = query.Where("week_no = ?", weekNo)
	}
	if term := c.Query("term"); term != "" {
		query = query.Where("course.term = ?", term)
	}
	type result struct {
		model.CourseSession
		CourseName  string `json:"course_name"`
		TeacherName string `json:"teacher_name"`
	}
	var items []result
	if err := query.Select("course_session.*, course.course_name, course.teacher_name").
		Order("week_no, weekday, section, session_no").
		Scan(&items).Error; err != nil {
		fail(c, 500, "load course calendar failed")
		return
	}
	ok(c, items)
}

func (h *apiHandler) adminAttendanceDashboard(c *gin.Context) {
	type summary struct {
		Present int64 `json:"present"`
		Late    int64 `json:"late"`
		Absent  int64 `json:"absent"`
		Leave   int64 `json:"leave"`
		Unset   int64 `json:"unset"`
	}
	result := summary{}
	base := h.db.Table("attendance_detail").
		Joins("JOIN attendance_check ON attendance_check.id = attendance_detail.attendance_check_id").
		Joins("JOIN course_session ON course_session.id = attendance_check.course_session_id").
		Joins("JOIN course ON course.id = course_session.course_id")
	if weekNo := c.Query("week_no"); weekNo != "" {
		base = base.Where("course_session.week_no = ?", weekNo)
	}
	if term := c.Query("term"); term != "" {
		base = base.Where("course.term = ?", term)
	}
	if courseID := c.Query("course_id"); courseID != "" {
		base = base.Where("course.id = ?", courseID)
	}
	statuses := map[string]int{"present": 1, "late": 2, "absent": 3, "leave": 4, "unset": 0}
	for key, status := range statuses {
		var count int64
		_ = base.Where("attendance_detail.status = ?", status).Count(&count).Error
		switch key {
		case "present":
			result.Present = count
		case "late":
			result.Late = count
		case "absent":
			result.Absent = count
		case "leave":
			result.Leave = count
		case "unset":
			result.Unset = count
		}
	}
	ok(c, result)
}

func (h *apiHandler) adminAttendanceResults(c *gin.Context) {
	page, pageSize := parsePage(c)
	type item struct {
		AttendanceCheckID  uint64 `json:"attendance_check_id"`
		AttendanceDetailID uint64 `json:"attendance_detail_id"`
		CourseID           uint64 `json:"course_id"`
		CourseName         string `json:"course_name"`
		TeacherName        string `json:"teacher_name"`
		WeekNo             int    `json:"week_no"`
		SessionNo          int    `json:"session_no"`
		StudentID          string `json:"student_id"`
		Status             int    `json:"status"`
	}
	query := h.db.Table("attendance_detail").
		Select("attendance_check.id AS attendance_check_id, attendance_detail.id AS attendance_detail_id, course.id AS course_id, course.course_name, course.teacher_name, course_session.week_no, course_session.session_no, attendance_detail.student_id, attendance_detail.status").
		Joins("JOIN attendance_check ON attendance_check.id = attendance_detail.attendance_check_id").
		Joins("JOIN course_session ON course_session.id = attendance_check.course_session_id").
		Joins("JOIN course ON course.id = course_session.course_id")
	if courseID := c.Query("course_id"); courseID != "" {
		query = query.Where("course.id = ?", courseID)
	}
	if weekNo := c.Query("week_no"); weekNo != "" {
		query = query.Where("course_session.week_no = ?", weekNo)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("attendance_detail.status = ?", status)
	}
	var items []item
	total, err := paginate(query.Order("attendance_detail.id DESC"), page, pageSize, &items)
	if err != nil {
		fail(c, 500, "load attendance results failed")
		return
	}
	ok(c, pageResult{Items: items, Page: page, PageSize: pageSize, Total: total})
}

func (h *apiHandler) adminFreeTimeCalendar(c *gin.Context) {
	query := h.db.Model(&model.StudentFreeTime{})
	if term := c.Query("term"); term != "" {
		query = query.Where("term = ?", term)
	}
	var items []model.StudentFreeTime
	if err := query.Order("weekday, section, student_id").Find(&items).Error; err != nil {
		fail(c, 500, "load free time calendar failed")
		return
	}
	ok(c, items)
}

func (h *apiHandler) studentAvailableCourses(c *gin.Context) {
	type result struct {
		CourseSessionID   uint64     `json:"course_session_id"`
		CourseID          uint64     `json:"course_id"`
		CourseName        string     `json:"course_name"`
		TeacherName       string     `json:"teacher_name"`
		WeekNo            int        `json:"week_no"`
		Weekday           int        `json:"weekday"`
		Section           int        `json:"section"`
		BuildingName      string     `json:"building_name"`
		RoomName          string     `json:"room_name"`
		StartedAt         *time.Time `json:"started_at"`
		CanEnter          bool       `json:"can_enter"`
		EnterDeadline     string     `json:"enter_deadline"`
		AttendanceCheckID *uint64    `json:"attendance_check_id"`
	}

	var sessions []struct {
		model.CourseSession
		CourseName        string     `json:"course_name"`
		TeacherName       string     `json:"teacher_name"`
		AttendanceCheckID *uint64    `json:"attendance_check_id"`
		StartedAt         *time.Time `json:"started_at"`
	}

	weekday := int(time.Now().Weekday())
	if weekday == 0 {
		weekday = 7
	}

	if err := h.db.Table("course_session").
		Select("course_session.*, course.course_name, course.teacher_name, attendance_check.id AS attendance_check_id, attendance_check.started_at").
		Joins("JOIN course ON course.id = course_session.course_id").
		Joins("LEFT JOIN attendance_check ON attendance_check.course_session_id = course_session.id").
		Where("course_session.weekday = ?", weekday).
		Order("course_session.section ASC").
		Scan(&sessions).Error; err != nil {
		fail(c, 500, "load available courses failed")
		return
	}

	items := make([]result, 0, len(sessions))
	for _, session := range sessions {
		deadline, err := h.attendanceDeadline(session.CourseSession, time.Now())
		if err != nil {
			continue
		}
		items = append(items, result{
			CourseSessionID:   session.ID,
			CourseID:          session.CourseID,
			CourseName:        session.CourseName,
			TeacherName:       session.TeacherName,
			WeekNo:            session.WeekNo,
			Weekday:           session.Weekday,
			Section:           session.Section,
			BuildingName:      session.BuildingName,
			RoomName:          session.RoomName,
			StartedAt:         session.StartedAt,
			CanEnter:          time.Now().Before(deadline),
			EnterDeadline:     deadline.Format("2006-01-02 15:04:05"),
			AttendanceCheckID: session.AttendanceCheckID,
		})
	}
	ok(c, items)
}

func (h *apiHandler) studentEnterAttendanceCheck(c *gin.Context) {
	user, _ := currentUser(c)
	var req struct {
		CourseSessionID uint64 `json:"course_session_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "invalid request")
		return
	}

	var session model.CourseSession
	if err := h.db.First(&session, req.CourseSessionID).Error; err != nil {
		fail(c, 404, "course session not found")
		return
	}
	if !h.withinDeadline(session, time.Now()) {
		fail(c, 403, "attendance entry deadline passed")
		return
	}

	var attendanceCheck model.AttendanceCheck
	err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&attendanceCheck, "course_session_id = ?", req.CourseSessionID).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			attendanceCheck = model.AttendanceCheck{
				CourseSessionID: req.CourseSessionID,
				StartedBy:       user.StudentID,
				StartedAt:       time.Now(),
			}
			if err := tx.Create(&attendanceCheck).Error; err != nil {
				return err
			}
			var students []model.CourseStudent
			if err := tx.Where("course_id = ?", session.CourseID).Find(&students).Error; err != nil {
				return err
			}
			details := make([]model.AttendanceDetail, 0, len(students))
			for _, student := range students {
				details = append(details, model.AttendanceDetail{
					AttendanceCheckID: attendanceCheck.ID,
					StudentID:         student.StudentID,
					Status:            model.AttendanceUnset,
				})
			}
			if len(details) > 0 {
				return tx.Create(&details).Error
			}
			return nil
		}
		return nil
	})
	if err != nil {
		fail(c, 500, "enter attendance check failed")
		return
	}
	c.Params = append(c.Params, gin.Param{Key: "id", Value: fmt.Sprintf("%d", attendanceCheck.ID)})
	h.studentGetAttendanceCheck(c)
}

func (h *apiHandler) studentGetAttendanceCheck(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	var attendanceCheck model.AttendanceCheck
	if err := h.db.First(&attendanceCheck, id).Error; err != nil {
		fail(c, 404, "attendance check not found")
		return
	}
	var session model.CourseSession
	if err := h.db.First(&session, attendanceCheck.CourseSessionID).Error; err != nil {
		fail(c, 404, "course session not found")
		return
	}
	if !h.withinDeadline(session, time.Now()) {
		fail(c, 403, "attendance entry deadline passed")
		return
	}
	var course model.Course
	var details []model.AttendanceDetail
	_ = h.db.First(&course, session.CourseID).Error
	_ = h.db.Where("attendance_check_id = ?", attendanceCheck.ID).Order("id ASC").Find(&details).Error
	ok(c, gin.H{
		"attendance_check": attendanceCheck,
		"course_session":   session,
		"course":           course,
		"students":         details,
	})
}

func (h *apiHandler) studentUpdateAttendanceStatus(c *gin.Context) {
	user, _ := currentUser(c)
	id, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	var req struct {
		Status int `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "invalid request")
		return
	}
	if req.Status < model.AttendanceUnset || req.Status > model.AttendanceOnLeave {
		fail(c, 400, "invalid status")
		return
	}

	var detail model.AttendanceDetail
	if err := h.db.First(&detail, id).Error; err != nil {
		fail(c, 404, "attendance detail not found")
		return
	}
	var attendanceCheck model.AttendanceCheck
	if err := h.db.First(&attendanceCheck, detail.AttendanceCheckID).Error; err != nil {
		fail(c, 404, "attendance check not found")
		return
	}
	var session model.CourseSession
	if err := h.db.First(&session, attendanceCheck.CourseSessionID).Error; err != nil {
		fail(c, 404, "course session not found")
		return
	}
	if !h.withinDeadline(session, time.Now()) {
		fail(c, 403, "attendance entry deadline passed")
		return
	}

	if err := h.updateAttendanceStatus(detail, req.Status, user.StudentID, false); err != nil {
		fail(c, 500, "update attendance status failed")
		return
	}
	ok(c, gin.H{})
}

func (h *apiHandler) studentCompleteAttendanceCheck(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	var attendanceCheck model.AttendanceCheck
	if err := h.db.First(&attendanceCheck, id).Error; err != nil {
		fail(c, 404, "attendance check not found")
		return
	}
	var session model.CourseSession
	if err := h.db.First(&session, attendanceCheck.CourseSessionID).Error; err != nil {
		fail(c, 404, "course session not found")
		return
	}
	if !h.withinDeadline(session, time.Now()) {
		fail(c, 403, "attendance entry deadline passed")
		return
	}
	ok(c, gin.H{"attendance_check_id": id, "completed": true})
}

func (h *apiHandler) adminGetAttendanceCheck(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	var attendanceCheck model.AttendanceCheck
	if err := h.db.First(&attendanceCheck, id).Error; err != nil {
		fail(c, 404, "attendance check not found")
		return
	}
	var details []model.AttendanceDetail
	_ = h.db.Where("attendance_check_id = ?", id).Order("id ASC").Find(&details).Error
	ok(c, gin.H{"attendance_check": attendanceCheck, "details": details})
}

func (h *apiHandler) adminUpdateAttendanceStatus(c *gin.Context) {
	user, _ := currentUser(c)
	id, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	var req struct {
		Status int `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "invalid request")
		return
	}
	var detail model.AttendanceDetail
	if err := h.db.First(&detail, id).Error; err != nil {
		fail(c, 404, "attendance detail not found")
		return
	}
	if err := h.updateAttendanceStatus(detail, req.Status, user.StudentID, true); err != nil {
		fail(c, 500, "update attendance status failed")
		return
	}
	ok(c, gin.H{})
}

func (h *apiHandler) updateAttendanceStatus(detail model.AttendanceDetail, newStatus int, operatorID string, writeAdminLog bool) error {
	now := time.Now()
	oldStatus := detail.Status
	statusSetBy := operatorID
	return h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&detail).Updates(map[string]interface{}{
			"status":        newStatus,
			"status_set_by": statusSetBy,
			"status_set_at": now,
		}).Error; err != nil {
			return err
		}
		logItem := model.AttendanceDetailLog{
			AttendanceDetailID: detail.ID,
			AttendanceCheckID:  detail.AttendanceCheckID,
			StudentID:          detail.StudentID,
			OperatorID:         operatorID,
			OldStatus:          &oldStatus,
			NewStatus:          newStatus,
			OperationType:      "set_status",
			OperatedAt:         now,
		}
		if err := tx.Create(&logItem).Error; err != nil {
			return err
		}
		if writeAdminLog {
			oldValueBytes, _ := json.Marshal(gin.H{"status": oldStatus})
			newValueBytes, _ := json.Marshal(gin.H{"status": newStatus})
			oldValue := string(oldValueBytes)
			newValue := string(newValueBytes)
			adminLog := model.AdminOperationLog{
				OperatorID:  operatorID,
				TargetTable: "attendance_detail",
				TargetID:    detail.ID,
				ActionType:  "update",
				OldValue:    &oldValue,
				NewValue:    &newValue,
			}
			if err := tx.Create(&adminLog).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (h *apiHandler) adminAttendanceDetailLogs(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	var logs []model.AttendanceDetailLog
	if err := h.db.Where("attendance_detail_id = ?", id).Order("operated_at DESC").Find(&logs).Error; err != nil {
		fail(c, 500, "load attendance detail logs failed")
		return
	}
	ok(c, logs)
}

func (h *apiHandler) listAdminOperationLogs(c *gin.Context) {
	page, pageSize := parsePage(c)
	query := h.db.Model(&model.AdminOperationLog{})
	var logs []model.AdminOperationLog
	total, err := paginate(query.Order("created_at DESC"), page, pageSize, &logs)
	if err != nil {
		fail(c, 500, "list admin operation logs failed")
		return
	}
	ok(c, pageResult{Items: logs, Page: page, PageSize: pageSize, Total: total})
}

func (h *apiHandler) listAttendanceDetailLogs(c *gin.Context) {
	page, pageSize := parsePage(c)
	query := h.db.Model(&model.AttendanceDetailLog{})
	var logs []model.AttendanceDetailLog
	total, err := paginate(query.Order("operated_at DESC"), page, pageSize, &logs)
	if err != nil {
		fail(c, 500, "list attendance detail logs failed")
		return
	}
	ok(c, pageResult{Items: logs, Page: page, PageSize: pageSize, Total: total})
}
