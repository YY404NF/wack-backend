package query

import (
	"time"

	"gorm.io/gorm"

	"wack-backend/internal/model"
)

type AttendanceDashboardSummary struct {
	Present int64 `json:"present"`
	Late    int64 `json:"late"`
	Absent  int64 `json:"absent"`
	Leave   int64 `json:"leave"`
	Unset   int64 `json:"unset"`
}

type AttendanceResultItem struct {
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

type AttendanceDetailItem struct {
	ID                uint64     `json:"id"`
	AttendanceCheckID uint64     `json:"attendance_check_id"`
	StudentID         string     `json:"student_id"`
	RealName          string     `json:"real_name"`
	Status            int        `json:"status"`
	StatusSetByUserID *uint64    `json:"status_set_by_user_id"`
	StatusSetAt       *time.Time `json:"status_set_at"`
}

type AttendanceCheckPayload struct {
	ID                 uint64    `json:"id"`
	CourseSessionID    uint64    `json:"course_session_id"`
	StartedByUserID    uint64    `json:"started_by_user_id"`
	StartedByStudentID string    `json:"started_by_student_id"`
	StartedAt          time.Time `json:"started_at"`
}

type AvailableCourseItem struct {
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

type SessionWithCourse struct {
	model.CourseSession
	CourseName        string     `json:"course_name"`
	TeacherName       string     `json:"teacher_name"`
	AttendanceCheckID *uint64    `json:"attendance_check_id"`
	StartedAt         *time.Time `json:"started_at"`
}

type AttendanceDetailLogItem struct {
	ID                 uint64    `json:"id"`
	AttendanceDetailID uint64    `json:"attendance_detail_id"`
	AttendanceCheckID  uint64    `json:"attendance_check_id"`
	StudentID          string    `json:"student_id"`
	RealName           string    `json:"real_name"`
	OperatorUserID     uint64    `json:"operator_user_id"`
	OperatorStudentID  string    `json:"operator_student_id"`
	OldStatus          *int      `json:"old_status"`
	NewStatus          int       `json:"new_status"`
	OperationType      string    `json:"operation_type"`
	OperatedAt         time.Time `json:"operated_at"`
	CreatedAt          time.Time `json:"created_at"`
}

type AttendanceQuery struct {
	db *gorm.DB
}

func NewAttendanceQuery(db *gorm.DB) *AttendanceQuery {
	return &AttendanceQuery{db: db}
}

func (q *AttendanceQuery) DashboardSummary(weekNo, term, courseID string) (AttendanceDashboardSummary, error) {
	result := AttendanceDashboardSummary{}
	base := q.db.Table("attendance_detail").
		Joins("JOIN attendance_check ON attendance_check.id = attendance_detail.attendance_check_id").
		Joins("JOIN course_session ON course_session.id = attendance_check.course_session_id").
		Joins("JOIN course ON course.id = course_session.course_id")
	if weekNo != "" {
		base = base.Where("course_session.week_no = ?", weekNo)
	}
	if term != "" {
		base = base.Where("course.term = ?", term)
	}
	if courseID != "" {
		base = base.Where("course.id = ?", courseID)
	}
	statuses := map[int]*int64{
		1: &result.Present,
		2: &result.Late,
		3: &result.Absent,
		4: &result.Leave,
		0: &result.Unset,
	}
	for status, target := range statuses {
		var count int64
		if err := base.Where("attendance_detail.status = ?", status).Count(&count).Error; err != nil {
			return AttendanceDashboardSummary{}, err
		}
		*target = count
	}
	return result, nil
}

func (q *AttendanceQuery) AttendanceResults(weekNo, courseID, status string, page, pageSize int) ([]AttendanceResultItem, int64, error) {
	query := q.db.Table("attendance_detail").
		Select("attendance_check.id AS attendance_check_id, attendance_detail.id AS attendance_detail_id, course.id AS course_id, course.course_name, course.teacher_name, course_session.week_no, course_session.session_no, attendance_detail.student_id, attendance_detail.status").
		Joins("JOIN attendance_check ON attendance_check.id = attendance_detail.attendance_check_id").
		Joins("JOIN course_session ON course_session.id = attendance_check.course_session_id").
		Joins("JOIN course ON course.id = course_session.course_id")
	if courseID != "" {
		query = query.Where("course.id = ?", courseID)
	}
	if weekNo != "" {
		query = query.Where("course_session.week_no = ?", weekNo)
	}
	if status != "" {
		query = query.Where("attendance_detail.status = ?", status)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []AttendanceResultItem
	if err := query.Order("attendance_detail.id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Scan(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (q *AttendanceQuery) AvailableSessions(weekday, weekNo int) ([]SessionWithCourse, error) {
	var sessions []SessionWithCourse
	err := q.db.Table("course_session").
		Select("course_session.*, course.course_name, course.teacher_name, attendance_check.id AS attendance_check_id, attendance_check.started_at").
		Joins("JOIN course ON course.id = course_session.course_id").
		Joins("LEFT JOIN attendance_check ON attendance_check.course_session_id = course_session.id").
		Where("course_session.weekday = ? AND course_session.week_no = ?", weekday, weekNo).
		Order("course_session.section ASC").
		Scan(&sessions).Error
	return sessions, err
}

func (q *AttendanceQuery) AttendanceCheckDetails(checkID uint64) ([]AttendanceDetailItem, error) {
	var details []AttendanceDetailItem
	err := q.db.Table("attendance_detail").
		Select("attendance_detail.id, attendance_detail.attendance_check_id, attendance_detail.student_id, attendance_detail.real_name, attendance_detail.status, attendance_detail.status_set_by_user_id, attendance_detail.status_set_at").
		Where("attendance_check_id = ?", checkID).
		Order("attendance_detail.id ASC").
		Scan(&details).Error
	return details, err
}

func (q *AttendanceQuery) AttendanceDetailLogs(detailID uint64) ([]AttendanceDetailLogItem, error) {
	var logs []AttendanceDetailLogItem
	err := q.db.Table("attendance_detail_log").
		Select("attendance_detail_log.id, attendance_detail_log.attendance_detail_id, attendance_detail_log.attendance_check_id, attendance_detail_log.student_id, attendance_detail_log.real_name, attendance_detail_log.operator_user_id, operator_user.student_id AS operator_student_id, attendance_detail_log.old_status, attendance_detail_log.new_status, attendance_detail_log.operation_type, attendance_detail_log.operated_at, attendance_detail_log.created_at").
		Joins("JOIN user AS operator_user ON operator_user.id = attendance_detail_log.operator_user_id").
		Where("attendance_detail_log.attendance_detail_id = ?", detailID).
		Order("attendance_detail_log.operated_at DESC").
		Scan(&logs).Error
	return logs, err
}
