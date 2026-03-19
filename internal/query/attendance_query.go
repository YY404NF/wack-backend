package query

import (
	"time"

	"gorm.io/gorm"
)

type AttendanceDashboardSummary struct {
	Present int64 `json:"present"`
	Late    int64 `json:"late"`
	Absent  int64 `json:"absent"`
	Leave   int64 `json:"leave"`
	Unset   int64 `json:"unset"`
}

type AttendanceResultItem struct {
	CourseGroupLessonID uint64 `json:"course_group_lesson_id"`
	AttendanceRecordID  uint64 `json:"attendance_record_id"`
	CourseID            uint64 `json:"course_id"`
	TermID              uint64 `json:"term_id"`
	Term                string `json:"term"`
	CourseName          string `json:"course_name"`
	TeacherName         string `json:"teacher_name"`
	WeekNo              int    `json:"week_no"`
	SessionNo           int    `json:"session_no"`
	StudentID           string `json:"student_id"`
	RealName            string `json:"real_name"`
	ClassName           string `json:"class_name"`
	Status              int    `json:"status"`
}

type AttendanceRecordItem struct {
	ID                 uint64     `json:"id"`
	CourseGroupLessonID uint64     `json:"course_group_lesson_id"`
	StudentID          string     `json:"student_id"`
	RealName           string     `json:"real_name"`
	ClassID            *uint64    `json:"class_id"`
	ClassName          string     `json:"class_name"`
	Status             int        `json:"status"`
	StatusSetByUserID  *uint64    `json:"status_set_by_user_id"`
	StatusSetAt        *time.Time `json:"status_set_at"`
}

type AttendanceClassGroupItem struct {
	ClassID      *uint64 `json:"class_id"`
	ClassName    string  `json:"class_name"`
	StudentCount int64   `json:"student_count"`
}

type AvailableCourseItem struct {
	CourseGroupLessonID uint64  `json:"course_group_lesson_id"`
	CourseID            uint64  `json:"course_id"`
	CourseName          string  `json:"course_name"`
	TeacherName         string  `json:"teacher_name"`
	WeekNo              int     `json:"week_no"`
	Weekday             int     `json:"weekday"`
	Section             int     `json:"section"`
	BuildingName        string  `json:"building_name"`
	RoomName            string  `json:"room_name"`
	CanEnter            bool    `json:"can_enter"`
	EnterDeadline       string  `json:"enter_deadline"`
}

type SessionWithCourse struct {
	ID                  uint64  `json:"id"`
	CourseID            uint64  `json:"course_id"`
	SessionNo           int     `json:"session_no"`
	WeekNo              int     `json:"week_no"`
	Weekday             int     `json:"weekday"`
	Section             int     `json:"section"`
	BuildingName        string  `json:"building_name"`
	RoomName            string  `json:"room_name"`
	CourseName          string  `json:"course_name"`
	TeacherName         string  `json:"teacher_name"`
}

type AttendanceRecordLogItem struct {
	ID                 uint64    `json:"id"`
	AttendanceRecordID uint64    `json:"attendance_record_id"`
	CourseGroupLessonID uint64    `json:"course_group_lesson_id"`
	StudentID          string    `json:"student_id"`
	RealName           string    `json:"real_name"`
	OperatorUserID     uint64    `json:"operator_user_id"`
	OperatorLoginID    string    `json:"operator_login_id"`
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
	base := q.db.Table("attendance_record").
		Joins("JOIN course_group_lesson ON course_group_lesson.id = attendance_record.course_group_lesson_id").
		Joins("JOIN course_group ON course_group.id = course_group_lesson.course_group_id").
		Joins("JOIN course ON course.id = course_group.course_id")
	if weekNo != "" {
		base = base.Where("course_group_lesson.week_no = ?", weekNo)
	}
	if term != "" {
		base = base.Where("course.term_id IN (SELECT id FROM term WHERE name = ?)", term)
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
		if err := base.Where("attendance_record.attendance_status = ?", status).Count(&count).Error; err != nil {
			return AttendanceDashboardSummary{}, err
		}
		*target = count
	}
	return result, nil
}

func (q *AttendanceQuery) AttendanceResults(weekNo, courseID, status string, page, pageSize int) ([]AttendanceResultItem, int64, error) {
	query := q.db.Table("attendance_record").
		Select(`
			attendance_record.course_group_lesson_id AS course_group_lesson_id,
			attendance_record.id AS attendance_record_id,
			course.id AS course_id,
			course.term_id AS term_id,
			term.name AS term,
			course.course_name,
			course.teacher_name,
			course_group_lesson.week_no,
			ROW_NUMBER() OVER (
				PARTITION BY course.id
				ORDER BY course_group_lesson.week_no, course_group_lesson.weekday, course_group_lesson.section, course_group_lesson.id
			) AS session_no,
			student.student_no AS student_id,
			student.student_name AS real_name,
			COALESCE(class.class_name, '其他学生') AS class_name,
			attendance_record.attendance_status AS status
		`).
		Joins("JOIN course_group_lesson ON course_group_lesson.id = attendance_record.course_group_lesson_id").
		Joins("JOIN course_group ON course_group.id = course_group_lesson.course_group_id").
		Joins("JOIN course ON course.id = course_group.course_id").
		Joins("JOIN term ON term.id = course.term_id").
		Joins("JOIN student ON student.id = attendance_record.student_id").
		Joins("LEFT JOIN class ON class.id = attendance_record.class_id")
	if courseID != "" {
		query = query.Where("course.id = ?", courseID)
	}
	if weekNo != "" {
		query = query.Where("course_group_lesson.week_no = ?", weekNo)
	}
	if status != "" {
		query = query.Where("attendance_record.attendance_status = ?", status)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []AttendanceResultItem
	if err := query.Order("attendance_record.id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Scan(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (q *AttendanceQuery) AvailableCourseGroupLessons(weekday, weekNo int) ([]SessionWithCourse, error) {
	var sessions []SessionWithCourse
	err := q.db.Table("course_group_lesson").
		Select(`
			course_group_lesson.id,
			course_group.course_id,
			ROW_NUMBER() OVER (
				PARTITION BY course_group.course_id
				ORDER BY course_group_lesson.week_no, course_group_lesson.weekday, course_group_lesson.section, course_group_lesson.id
			) AS session_no,
			course_group_lesson.week_no,
			course_group_lesson.weekday,
			course_group_lesson.section,
			course_group_lesson.building_name,
			course_group_lesson.room_name,
			course.course_name,
			course.teacher_name
		`).
		Joins("JOIN course_group ON course_group.id = course_group_lesson.course_group_id").
		Joins("JOIN course ON course.id = course_group.course_id").
		Where("course_group_lesson.weekday = ? AND course_group_lesson.week_no = ? AND course_group_lesson.status = 1 AND course_group.status = 1", weekday, weekNo).
		Order("course_group_lesson.section ASC, course_group_lesson.id ASC").
		Scan(&sessions).Error
	return sessions, err
}

func (q *AttendanceQuery) AvailableCourseGroupLessonsForClass(weekday, weekNo int, classID uint64) ([]SessionWithCourse, error) {
	var sessions []SessionWithCourse
	err := q.db.Table("course_group_lesson").
		Select(`
			DISTINCT course_group_lesson.id,
			course_group.course_id,
			ROW_NUMBER() OVER (
				PARTITION BY course_group.course_id
				ORDER BY course_group_lesson.week_no, course_group_lesson.weekday, course_group_lesson.section, course_group_lesson.id
			) AS session_no,
			course_group_lesson.week_no,
			course_group_lesson.weekday,
			course_group_lesson.section,
			course_group_lesson.building_name,
			course_group_lesson.room_name,
			course.course_name,
			course.teacher_name
		`).
		Joins("JOIN course_group ON course_group.id = course_group_lesson.course_group_id").
		Joins("JOIN course ON course.id = course_group.course_id").
		Joins("JOIN course_group_student ON course_group_student.course_group_id = course_group.id AND course_group_student.class_id = ? AND course_group_student.status = 1", classID).
		Where("course_group_lesson.weekday = ? AND course_group_lesson.week_no = ? AND course_group_lesson.status = 1 AND course_group.status = 1", weekday, weekNo).
		Order("course_group_lesson.section ASC, course_group_lesson.id ASC").
		Scan(&sessions).Error
	return sessions, err
}

func (q *AttendanceQuery) AttendanceSessionRecords(sessionID uint64) ([]AttendanceRecordItem, error) {
	var records []AttendanceRecordItem
	err := q.db.Table("attendance_record").
		Select(`
			attendance_record.id,
			attendance_record.course_group_lesson_id AS course_group_lesson_id,
			student.student_no AS student_id,
			student.student_name AS real_name,
			attendance_record.class_id,
			COALESCE(class.class_name, '') AS class_name,
			attendance_record.attendance_status AS status,
			attendance_record.updated_by_user_id AS status_set_by_user_id,
			attendance_record.updated_at AS status_set_at
		`).
		Joins("JOIN student ON student.id = attendance_record.student_id").
		Joins("LEFT JOIN class ON class.id = student.class_id").
		Where("attendance_record.course_group_lesson_id = ?", sessionID).
		Order("attendance_record.id ASC").
		Scan(&records).Error
	return records, err
}

func (q *AttendanceQuery) AttendanceSessionRecordsForClass(sessionID uint64, classID uint64) ([]AttendanceRecordItem, error) {
	var records []AttendanceRecordItem
	err := q.db.Table("attendance_record").
		Select(`
			attendance_record.id,
			attendance_record.course_group_lesson_id AS course_group_lesson_id,
			student.student_no AS student_id,
			student.student_name AS real_name,
			attendance_record.class_id,
			COALESCE(class.class_name, '') AS class_name,
			attendance_record.attendance_status AS status,
			attendance_record.updated_by_user_id AS status_set_by_user_id,
			attendance_record.updated_at AS status_set_at
		`).
		Joins("JOIN student ON student.id = attendance_record.student_id AND student.class_id = ?", classID).
		Joins("JOIN class ON class.id = student.class_id").
		Where("attendance_record.course_group_lesson_id = ?", sessionID).
		Order("attendance_record.id ASC").
		Scan(&records).Error
	return records, err
}

func (q *AttendanceQuery) AttendanceClassGroups(checkID uint64) ([]AttendanceClassGroupItem, error) {
	var groups []AttendanceClassGroupItem
	if err := q.db.Table("attendance_record").
		Select("student.class_id AS class_id, COALESCE(class.class_name, '其他学生') AS class_name, COUNT(attendance_record.id) AS student_count").
		Joins("LEFT JOIN student ON student.id = attendance_record.student_id").
		Joins("LEFT JOIN class ON class.id = student.class_id").
		Where("attendance_record.course_group_lesson_id = ?", checkID).
		Group("student.class_id, class.class_name").
		Order("class.class_name ASC").
		Scan(&groups).Error; err != nil {
		return nil, err
	}
	return groups, nil
}

func (q *AttendanceQuery) AttendanceClassGroupsForClass(checkID uint64, classID uint64) ([]AttendanceClassGroupItem, error) {
	var groups []AttendanceClassGroupItem
	err := q.db.Table("attendance_record").
		Select("class.id AS class_id, class.class_name, COUNT(attendance_record.id) AS student_count").
		Joins("JOIN student ON student.id = attendance_record.student_id AND student.class_id = ?", classID).
		Joins("JOIN class ON class.id = student.class_id").
		Where("attendance_record.course_group_lesson_id = ?", checkID).
		Group("class.id, class.class_name").
		Scan(&groups).Error
	return groups, err
}

func (q *AttendanceQuery) CourseGroupLessonBelongsToClass(courseGroupLessonID uint64, classID uint64) (bool, error) {
	var count int64
	err := q.db.Table("course_group_lesson").
		Joins("JOIN course_group ON course_group.id = course_group_lesson.course_group_id").
		Joins("JOIN course_group_student ON course_group_student.course_group_id = course_group.id AND course_group_student.class_id = ? AND course_group_student.status = 1", classID).
		Where("course_group_lesson.id = ? AND course_group_lesson.status = 1 AND course_group.status = 1", courseGroupLessonID).
		Count(&count).Error
	return count > 0, err
}

func (q *AttendanceQuery) AttendanceRecordBelongsToClass(sessionID uint64, recordID uint64, classID uint64) (bool, error) {
	var count int64
	err := q.db.Table("attendance_record").
		Joins("JOIN student ON student.id = attendance_record.student_id AND student.class_id = ?", classID).
		Where("attendance_record.id = ? AND attendance_record.course_group_lesson_id = ?", recordID, sessionID).
		Count(&count).Error
	return count > 0, err
}

func (q *AttendanceQuery) AttendanceRecordLogsByID(recordID uint64) ([]AttendanceRecordLogItem, error) {
	var logs []AttendanceRecordLogItem
	err := q.db.Table("attendance_record_log").
		Select(`
			attendance_record_log.id,
			attendance_record_log.attendance_record_id,
			attendance_record.course_group_lesson_id AS course_group_lesson_id,
			student.student_no AS student_id,
			student.student_name AS real_name,
			attendance_record_log.operated_by_user_id AS operator_user_id,
			operator_user.login_id AS operator_login_id,
			attendance_record_log.old_attendance_status AS old_status,
			attendance_record_log.new_attendance_status AS new_status,
			'set_status' AS operation_type,
			attendance_record_log.created_at AS operated_at,
			attendance_record_log.created_at
		`).
		Joins("JOIN attendance_record ON attendance_record.id = attendance_record_log.attendance_record_id").
		Joins("JOIN student ON student.id = attendance_record.student_id").
		Joins("JOIN user AS operator_user ON operator_user.id = attendance_record_log.operated_by_user_id").
		Where("attendance_record_log.attendance_record_id = ?", recordID).
		Order("attendance_record_log.created_at DESC").
		Scan(&logs).Error
	return logs, err
}

func (q *AttendanceQuery) AttendanceRecordLogs(recordID uint64) ([]AttendanceRecordLogItem, error) {
	return q.AttendanceRecordLogsByID(recordID)
}
