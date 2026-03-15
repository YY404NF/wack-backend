package httpserver

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"wack-backend/internal/model"
)

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
	type courseStudent struct {
		ID        uint64    `json:"id"`
		CourseID  uint64    `json:"course_id"`
		UserID    uint64    `json:"user_id"`
		StudentID string    `json:"student_id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}
	var students []courseStudent
	var classes []model.CourseClass
	var sessions []model.CourseSession
	_ = h.db.Table("course_student").
		Select("course_student.id, course_student.course_id, course_student.user_id, user.student_id, course_student.created_at, course_student.updated_at").
		Joins("JOIN user ON user.id = course_student.user_id").
		Where("course_student.course_id = ?", id).
		Find(&students).Error
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
		users, err := h.findUsersByStudentIDs(req.StudentIDs)
		if err != nil {
			return err
		}
		userIDByStudentID := make(map[string]uint64, len(users))
		for _, user := range users {
			userIDByStudentID[user.StudentID] = user.ID
		}
		var relations []model.CourseStudent
		for _, studentID := range req.StudentIDs {
			userID, ok := userIDByStudentID[studentID]
			if !ok {
				return fmt.Errorf("student %s not found", studentID)
			}
			relations = append(relations, model.CourseStudent{CourseID: id, UserID: userID})
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
