package query

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

type AttendanceDashboardSummary struct {
	Present int64 `json:"present"`
	Late    int64 `json:"late"`
	Absent  int64 `json:"absent"`
	Leave   int64 `json:"leave"`
}

type OverviewCourseRankingItem struct {
	CourseID       uint64  `json:"course_id"`
	Rank           int     `json:"rank"`
	CourseName     string  `json:"course_name"`
	TeacherName    string  `json:"teacher_name"`
	Grade          int     `json:"grade"`
	ArrivedCount   int64   `json:"arrived_count"`
	TotalCount     int64   `json:"total_count"`
	AttendanceRate float64 `json:"attendance_rate"`
}

type OverviewClassRankingItem struct {
	ClassID        uint64  `json:"class_id"`
	Rank           int     `json:"rank"`
	ClassName      string  `json:"class_name"`
	MajorName      string  `json:"major_name"`
	Grade          int     `json:"grade"`
	ArrivedCount   int64   `json:"arrived_count"`
	TotalCount     int64   `json:"total_count"`
	AttendanceRate float64 `json:"attendance_rate"`
}

type OverviewStudentRankingItem struct {
	StudentRefID   uint64  `json:"student_ref_id"`
	Rank           int     `json:"rank"`
	StudentID      string  `json:"student_id"`
	RealName       string  `json:"real_name"`
	ClassName      string  `json:"class_name"`
	ArrivedCount   int64   `json:"arrived_count"`
	TotalCount     int64   `json:"total_count"`
	AttendanceRate float64 `json:"attendance_rate"`
}

type OverviewRecentSessionItem struct {
	CourseID            uint64  `json:"course_id"`
	CourseGroupID       uint64  `json:"course_group_id"`
	CourseGroupLessonID uint64  `json:"course_group_lesson_id"`
	CourseName          string  `json:"course_name"`
	TeacherName         string  `json:"teacher_name"`
	WeekNo              int     `json:"week_no"`
	Weekday             int     `json:"weekday"`
	Section             int     `json:"section"`
	BuildingName        string  `json:"building_name"`
	RoomName            string  `json:"room_name"`
	ClassSummary        string  `json:"class_summary"`
	StudentCount        int64   `json:"student_count"`
	RecordCount         int64   `json:"record_count"`
	PresentCount        int64   `json:"present_count"`
	LateCount           int64   `json:"late_count"`
	AbsentCount         int64   `json:"absent_count"`
	LeaveCount          int64   `json:"leave_count"`
	AttendanceRate      float64 `json:"attendance_rate"`
}

type OverviewRecentAbnormalItem struct {
	AttendanceRecordID  uint64 `json:"attendance_record_id"`
	CourseID            uint64 `json:"course_id"`
	CourseGroupID       uint64 `json:"course_group_id"`
	CourseGroupLessonID uint64 `json:"course_group_lesson_id"`
	StudentRefID        uint64 `json:"student_ref_id"`
	StudentID           string `json:"student_id"`
	RealName            string `json:"real_name"`
	ClassName           string `json:"class_name"`
	CourseName          string `json:"course_name"`
	TeacherName         string `json:"teacher_name"`
	Grade               int    `json:"grade"`
	Status              int    `json:"status"`
	WeekNo              int    `json:"week_no"`
	Weekday             int    `json:"weekday"`
	Section             int    `json:"section"`
}

type AdminOverviewData struct {
	Term                   string                       `json:"term"`
	CourseRankings         []OverviewCourseRankingItem  `json:"course_rankings"`
	CourseRankingsTotal    int64                        `json:"course_rankings_total"`
	CourseRankingsHasMore  bool                         `json:"course_rankings_has_more"`
	CourseRankingsMinRate  float64                      `json:"course_rankings_min_rate"`
	ClassRankings          []OverviewClassRankingItem   `json:"class_rankings"`
	ClassRankingsTotal     int64                        `json:"class_rankings_total"`
	ClassRankingsHasMore   bool                         `json:"class_rankings_has_more"`
	ClassRankingsMinRate   float64                      `json:"class_rankings_min_rate"`
	StudentRankings        []OverviewStudentRankingItem `json:"student_rankings"`
	StudentRankingsTotal   int64                        `json:"student_rankings_total"`
	StudentRankingsHasMore bool                         `json:"student_rankings_has_more"`
	StudentRankingsMinRate float64                      `json:"student_rankings_min_rate"`
	RecentSessions         []OverviewRecentSessionItem  `json:"recent_sessions"`
	RecentSessionsTotal    int64                        `json:"recent_sessions_total"`
	RecentSessionsHasMore  bool                         `json:"recent_sessions_has_more"`
	RecentSessionsMinRate  float64                      `json:"recent_sessions_min_rate"`
	RecentAbnormalStudents []OverviewRecentAbnormalItem `json:"recent_abnormal_students"`
	RecentAbnormalTotal    int64                        `json:"recent_abnormal_students_total"`
	RecentAbnormalHasMore  bool                         `json:"recent_abnormal_students_has_more"`
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

type AttendanceSessionSummaryItem struct {
	CourseGroupLessonID uint64 `json:"course_group_lesson_id"`
	TermID              uint64 `json:"term_id"`
	Term                string `json:"term"`
	LessonDate          string `json:"lesson_date"`
	CourseID            uint64 `json:"course_id"`
	CourseName          string `json:"course_name"`
	TeacherName         string `json:"teacher_name"`
	WeekNo              int    `json:"week_no"`
	Weekday             int    `json:"weekday"`
	Section             int    `json:"section"`
	BuildingName        string `json:"building_name"`
	RoomName            string `json:"room_name"`
	ClassSummary        string `json:"class_summary"`
	SessionNo           int    `json:"session_no"`
	StudentCount        int64  `json:"student_count"`
	RecordCount         int64  `json:"record_count"`
	PresentCount        int64  `json:"present_count"`
	LateCount           int64  `json:"late_count"`
	AbsentCount         int64  `json:"absent_count"`
	LeaveCount          int64  `json:"leave_count"`
}

type AttendanceSessionSummaryListInput struct {
	Term             string
	Keyword          string
	LessonDate       string
	LessonDateFrom   string
	LessonDateTo     string
	CourseName       string
	TeacherName      string
	WeekNo           string
	Weekday          string
	Section          string
	ClassID          string
	ClassName        string
	Status           string
	IncludeUnchecked bool
	Page             int
	PageSize         int
}

type AttendanceRecordItem struct {
	ID                  uint64     `json:"id"`
	AttendanceRecordID  *uint64    `json:"attendance_record_id"`
	CourseGroupLessonID uint64     `json:"course_group_lesson_id"`
	StudentID           string     `json:"student_id"`
	RealName            string     `json:"real_name"`
	ClassID             *uint64    `json:"class_id"`
	ClassName           string     `json:"class_name"`
	Status              *int       `json:"status"`
	StatusSetByUserID   *uint64    `json:"status_set_by_user_id"`
	StatusSetAt         *time.Time `json:"status_set_at"`
	OperatorName        string     `json:"operator_name"`
	OperatedAt          *time.Time `json:"operated_at"`
}

type AttendanceClassGroupItem struct {
	ClassID      *uint64 `json:"class_id"`
	ClassName    string  `json:"class_name"`
	StudentCount int64   `json:"student_count"`
}

type AvailableCourseItem struct {
	CourseGroupLessonID uint64 `json:"course_group_lesson_id"`
	CourseID            uint64 `json:"course_id"`
	CourseName          string `json:"course_name"`
	TeacherName         string `json:"teacher_name"`
	WeekNo              int    `json:"week_no"`
	Weekday             int    `json:"weekday"`
	Section             int    `json:"section"`
	BuildingName        string `json:"building_name"`
	RoomName            string `json:"room_name"`
	CanEnter            bool   `json:"can_enter"`
	AvailabilityStatus  string `json:"availability_status"`
	EnterDeadline       string `json:"enter_deadline"`
}

type SessionWithCourse struct {
	TermID       uint64 `json:"term_id"`
	ID           uint64 `json:"id"`
	CourseID     uint64 `json:"course_id"`
	SessionNo    int    `json:"session_no"`
	WeekNo       int    `json:"week_no"`
	Weekday      int    `json:"weekday"`
	Section      int    `json:"section"`
	BuildingName string `json:"building_name"`
	RoomName     string `json:"room_name"`
	CourseName   string `json:"course_name"`
	TeacherName  string `json:"teacher_name"`
}

type AttendanceRecordLogItem struct {
	ID                  uint64    `json:"id"`
	AttendanceRecordID  uint64    `json:"attendance_record_id"`
	CourseGroupLessonID uint64    `json:"course_group_lesson_id"`
	StudentID           string    `json:"student_id"`
	RealName            string    `json:"real_name"`
	OperatorUserID      uint64    `json:"operator_user_id"`
	OperatorLoginID     string    `json:"operator_login_id"`
	OldStatus           *int      `json:"old_status"`
	NewStatus           int       `json:"new_status"`
	OperationType       string    `json:"operation_type"`
	OperatedAt          time.Time `json:"operated_at"`
	CreatedAt           time.Time `json:"created_at"`
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
		0: &result.Present,
		1: &result.Late,
		2: &result.Absent,
		3: &result.Leave,
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

func (q *AttendanceQuery) OverviewCourseRankings(termName string) ([]OverviewCourseRankingItem, error) {
	var items []OverviewCourseRankingItem
	err := q.db.Raw(`
		SELECT
			course.id AS course_id,
			course.course_name,
			course.teacher_name,
			course.grade,
			SUM(CASE WHEN attendance_record.attendance_status IN (0, 1) THEN 1 ELSE 0 END) AS arrived_count,
			SUM(CASE WHEN attendance_record.attendance_status IN (0, 1, 2) THEN 1 ELSE 0 END) AS total_count,
			CASE
				WHEN SUM(CASE WHEN attendance_record.attendance_status IN (0, 1, 2) THEN 1 ELSE 0 END) = 0 THEN 1
				ELSE CAST(SUM(CASE WHEN attendance_record.attendance_status IN (0, 1) THEN 1 ELSE 0 END) AS REAL)
					/ SUM(CASE WHEN attendance_record.attendance_status IN (0, 1, 2) THEN 1 ELSE 0 END)
			END AS attendance_rate
		FROM attendance_record
		JOIN course ON course.id = attendance_record.course_id
		JOIN term ON term.id = attendance_record.term_id
		WHERE term.name = ?
		GROUP BY course.id, course.course_name, course.teacher_name, course.grade
		ORDER BY attendance_rate DESC, total_count DESC, course.course_name ASC
	`, termName).Scan(&items).Error
	return items, err
}

func (q *AttendanceQuery) OverviewClassRankings(termName string) ([]OverviewClassRankingItem, error) {
	var items []OverviewClassRankingItem
	err := q.db.Raw(`
		SELECT
			class.id AS class_id,
			class.class_name,
			class.major_name,
			class.grade,
			SUM(CASE WHEN attendance_record.attendance_status IN (0, 1) THEN 1 ELSE 0 END) AS arrived_count,
			SUM(CASE WHEN attendance_record.attendance_status IN (0, 1, 2) THEN 1 ELSE 0 END) AS total_count,
			CASE
				WHEN SUM(CASE WHEN attendance_record.attendance_status IN (0, 1, 2) THEN 1 ELSE 0 END) = 0 THEN 1
				ELSE CAST(SUM(CASE WHEN attendance_record.attendance_status IN (0, 1) THEN 1 ELSE 0 END) AS REAL)
					/ SUM(CASE WHEN attendance_record.attendance_status IN (0, 1, 2) THEN 1 ELSE 0 END)
			END AS attendance_rate
		FROM attendance_record
		JOIN class ON class.id = attendance_record.class_id
		JOIN term ON term.id = attendance_record.term_id
		WHERE term.name = ?
		GROUP BY class.id, class.class_name, class.major_name, class.grade
		ORDER BY attendance_rate DESC, total_count DESC, class.class_name ASC
	`, termName).Scan(&items).Error
	return items, err
}

func (q *AttendanceQuery) OverviewStudentRankings(termName string) ([]OverviewStudentRankingItem, error) {
	var items []OverviewStudentRankingItem
	err := q.db.Raw(`
		SELECT
			student.id AS student_ref_id,
			student.student_no AS student_id,
			student.student_name AS real_name,
			COALESCE(class.class_name, '其他学生') AS class_name,
			SUM(CASE WHEN attendance_record.attendance_status IN (0, 1) THEN 1 ELSE 0 END) AS arrived_count,
			SUM(CASE WHEN attendance_record.attendance_status IN (0, 1, 2) THEN 1 ELSE 0 END) AS total_count,
			CASE
				WHEN SUM(CASE WHEN attendance_record.attendance_status IN (0, 1, 2) THEN 1 ELSE 0 END) = 0 THEN 1
				ELSE CAST(SUM(CASE WHEN attendance_record.attendance_status IN (0, 1) THEN 1 ELSE 0 END) AS REAL)
					/ SUM(CASE WHEN attendance_record.attendance_status IN (0, 1, 2) THEN 1 ELSE 0 END)
			END AS attendance_rate
		FROM attendance_record
		JOIN student ON student.id = attendance_record.student_id
		LEFT JOIN class ON class.id = student.class_id
		JOIN term ON term.id = attendance_record.term_id
		WHERE term.name = ?
		GROUP BY student.id, student.student_no, student.student_name
		ORDER BY attendance_rate DESC, total_count DESC, student.student_no ASC
	`, termName).Scan(&items).Error
	return items, err
}

func (q *AttendanceQuery) OverviewRecentSessions(termName string) ([]OverviewRecentSessionItem, error) {
	var items []OverviewRecentSessionItem
	err := q.db.Raw(`
		SELECT
			course.id AS course_id,
			course_group.id AS course_group_id,
			course_group_lesson.id AS course_group_lesson_id,
			course.course_name,
			course.teacher_name,
			course_group_lesson.week_no,
			course_group_lesson.weekday,
			course_group_lesson.section,
			course_group_lesson.building_name,
			course_group_lesson.room_name,
			COALESCE((
				SELECT GROUP_CONCAT(class_name, '、')
				FROM (
					SELECT DISTINCT class.class_name AS class_name
					FROM course_group_student
					JOIN class ON class.id = course_group_student.class_id
					WHERE course_group_student.course_group_id = course_group.id
					  AND course_group_student.class_id IS NOT NULL
					  AND course_group_student.status = 1
					ORDER BY class.class_name
				)
			), '其他学生') AS class_summary,
			(
				SELECT COUNT(1)
				FROM course_group_student AS cgs
				JOIN student ON student.id = cgs.student_id AND student.status = 1
				WHERE cgs.course_group_id = course_group.id
				  AND cgs.status = 1
			) AS student_count,
			COUNT(attendance_record.id) AS record_count,
			SUM(CASE WHEN attendance_record.attendance_status = 0 THEN 1 ELSE 0 END) AS present_count,
			SUM(CASE WHEN attendance_record.attendance_status = 1 THEN 1 ELSE 0 END) AS late_count,
			SUM(CASE WHEN attendance_record.attendance_status = 2 THEN 1 ELSE 0 END) AS absent_count,
			SUM(CASE WHEN attendance_record.attendance_status = 3 THEN 1 ELSE 0 END) AS leave_count,
			CASE
				WHEN SUM(CASE WHEN attendance_record.attendance_status IN (0, 1, 2) THEN 1 ELSE 0 END) = 0 THEN 1
				ELSE CAST(SUM(CASE WHEN attendance_record.attendance_status IN (0, 1) THEN 1 ELSE 0 END) AS REAL)
					/ SUM(CASE WHEN attendance_record.attendance_status IN (0, 1, 2) THEN 1 ELSE 0 END)
			END AS attendance_rate
		FROM attendance_record
		JOIN course_group_lesson ON course_group_lesson.id = attendance_record.course_group_lesson_id
		JOIN course_group ON course_group.id = course_group_lesson.course_group_id
		JOIN course ON course.id = course_group.course_id
		JOIN term ON term.id = attendance_record.term_id
		WHERE term.name = ?
		GROUP BY
			course_group_lesson.id,
			course.course_name,
			course.teacher_name,
			course_group_lesson.week_no,
			course_group_lesson.weekday,
			course_group_lesson.section,
			course_group_lesson.building_name,
			course_group_lesson.room_name,
			course_group.id
		ORDER BY
			course_group_lesson.week_no DESC,
			course_group_lesson.weekday DESC,
			course_group_lesson.section DESC,
			course_group_lesson.id DESC
	`, termName).Scan(&items).Error
	return items, err
}

func (q *AttendanceQuery) OverviewRecentAbnormalStudents(termName string) ([]OverviewRecentAbnormalItem, error) {
	var items []OverviewRecentAbnormalItem
	err := q.db.Raw(`
		SELECT
			attendance_record.id AS attendance_record_id,
			course.id AS course_id,
			course_group.id AS course_group_id,
			attendance_record.course_group_lesson_id AS course_group_lesson_id,
			student.id AS student_ref_id,
			student.student_no AS student_id,
			student.student_name AS real_name,
			COALESCE(class.class_name, '其他学生') AS class_name,
			course.course_name,
			course.teacher_name,
			course.grade,
			attendance_record.attendance_status AS status,
			course_group_lesson.week_no,
			course_group_lesson.weekday,
			course_group_lesson.section
		FROM attendance_record
		JOIN student ON student.id = attendance_record.student_id
		LEFT JOIN class ON class.id = attendance_record.class_id
		JOIN course_group_lesson ON course_group_lesson.id = attendance_record.course_group_lesson_id
		JOIN course_group ON course_group.id = course_group_lesson.course_group_id
		JOIN course ON course.id = attendance_record.course_id
		JOIN term ON term.id = attendance_record.term_id
		WHERE term.name = ?
		  AND attendance_record.attendance_status IN (1, 2, 3)
		ORDER BY
			course_group_lesson.week_no DESC,
			course_group_lesson.weekday DESC,
			course_group_lesson.section DESC,
			attendance_record.id DESC
	`, termName).Scan(&items).Error
	return items, err
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

func (q *AttendanceQuery) AttendanceSessionSummaries(input AttendanceSessionSummaryListInput) ([]AttendanceSessionSummaryItem, int64, error) {
	lessonDateExpr := "date(term.term_start_date, printf('+%d days', (course_group_lesson.week_no - 1) * 7 + (course_group_lesson.weekday - 1)))"
	base := q.db.Table("course_group_lesson").
		Select(`
			course_group_lesson.id AS course_group_lesson_id,
			term.id AS term_id,
			term.name AS term,
			` + lessonDateExpr + ` AS lesson_date,
			course.id AS course_id,
			course.course_name,
			course.teacher_name,
			course_group_lesson.week_no,
			course_group_lesson.weekday,
			course_group_lesson.section,
			course_group_lesson.building_name,
			course_group_lesson.room_name,
			COALESCE((
				SELECT GROUP_CONCAT(DISTINCT COALESCE(class.class_name, '其他学生'))
				FROM course_group_student AS cgs
				LEFT JOIN class ON class.id = cgs.class_id
				JOIN student ON student.id = cgs.student_id AND student.status = 1
				WHERE cgs.course_group_id = course_group_lesson.course_group_id
				  AND cgs.status = 1
			), '') AS class_summary,
			ROW_NUMBER() OVER (
				PARTITION BY course.id
				ORDER BY course_group_lesson.week_no, course_group_lesson.weekday, course_group_lesson.section, course_group_lesson.id
			) AS session_no,
			(
				SELECT COUNT(1)
				FROM course_group_student AS cgs
				JOIN student ON student.id = cgs.student_id AND student.status = 1
				WHERE cgs.course_group_id = course_group_lesson.course_group_id
				  AND cgs.status = 1
			) AS student_count,
			(
				SELECT COUNT(1)
				FROM attendance_record
				WHERE attendance_record.course_group_lesson_id = course_group_lesson.id
			) AS record_count,
			(
				SELECT COUNT(1)
				FROM attendance_record
				WHERE attendance_record.course_group_lesson_id = course_group_lesson.id
				  AND attendance_record.attendance_status = 0
			) AS present_count,
			(
				SELECT COUNT(1)
				FROM attendance_record
				WHERE attendance_record.course_group_lesson_id = course_group_lesson.id
				  AND attendance_record.attendance_status = 1
			) AS late_count,
			(
				SELECT COUNT(1)
				FROM attendance_record
				WHERE attendance_record.course_group_lesson_id = course_group_lesson.id
				  AND attendance_record.attendance_status = 2
			) AS absent_count,
			(
				SELECT COUNT(1)
				FROM attendance_record
				WHERE attendance_record.course_group_lesson_id = course_group_lesson.id
				  AND attendance_record.attendance_status = 3
			) AS leave_count
		`).
		Joins("JOIN course_group ON course_group.id = course_group_lesson.course_group_id").
		Joins("JOIN course ON course.id = course_group.course_id").
		Joins("JOIN term ON term.id = course.term_id").
		Where("course_group_lesson.status = 1 AND course_group.status = 1 AND course.status = 1")
	if value := strings.TrimSpace(input.Term); value != "" {
		base = base.Where("term.name = ?", value)
	}
	if value := strings.TrimSpace(input.LessonDate); value != "" {
		base = base.Where(fmt.Sprintf("%s = ?", lessonDateExpr), value)
	}
	if value := strings.TrimSpace(input.LessonDateFrom); value != "" {
		base = base.Where(fmt.Sprintf("%s >= ?", lessonDateExpr), value)
	}
	if value := strings.TrimSpace(input.LessonDateTo); value != "" {
		base = base.Where(fmt.Sprintf("%s <= ?", lessonDateExpr), value)
	}
	if value := strings.TrimSpace(input.WeekNo); value != "" {
		base = base.Where("course_group_lesson.week_no = ?", value)
	}
	if value := strings.TrimSpace(input.Weekday); value != "" {
		base = base.Where("course_group_lesson.weekday = ?", value)
	}
	if value := strings.TrimSpace(input.Section); value != "" {
		base = base.Where("course_group_lesson.section = ?", value)
	}
	if value := strings.TrimSpace(input.ClassID); value != "" {
		base = base.Where(`
			EXISTS (
				SELECT 1
				FROM course_group_student
				JOIN student ON student.id = course_group_student.student_id AND student.status = 1
				WHERE course_group_student.course_group_id = course_group_lesson.course_group_id
				  AND course_group_student.class_id = ?
				  AND course_group_student.status = 1
			)
		`, value)
	}
	if value := strings.TrimSpace(input.ClassName); value != "" {
		like := "%" + value + "%"
		base = base.Where(`
			EXISTS (
				SELECT 1
				FROM course_group_student
				JOIN student ON student.id = course_group_student.student_id AND student.status = 1
				LEFT JOIN class ON class.id = course_group_student.class_id
				WHERE course_group_student.course_group_id = course_group_lesson.course_group_id
				  AND course_group_student.status = 1
				  AND COALESCE(class.class_name, '其他学生') LIKE ?
			)
		`, like)
	}
	if value := strings.TrimSpace(input.CourseName); value != "" {
		base = base.Where("course.course_name LIKE ?", "%"+value+"%")
	}
	if value := strings.TrimSpace(input.TeacherName); value != "" {
		base = base.Where("course.teacher_name LIKE ?", "%"+value+"%")
	}
	if value := strings.TrimSpace(input.Keyword); value != "" {
		like := "%" + value + "%"
		base = base.Where("course.course_name LIKE ? OR course.teacher_name LIKE ? OR CAST(course_group_lesson.id AS TEXT) LIKE ?", like, like, like)
	}
	if !input.IncludeUnchecked {
		base = base.Where("EXISTS (SELECT 1 FROM attendance_record WHERE attendance_record.course_group_lesson_id = course_group_lesson.id)")
	}
	if value := strings.TrimSpace(input.Status); value != "" {
		base = base.Where(`
			EXISTS (
				SELECT 1
				FROM attendance_record
				WHERE attendance_record.course_group_lesson_id = course_group_lesson.id
				  AND attendance_record.attendance_status = ?
			)
		`, value)
	}

	countQuery := q.db.Table("(?) AS attendance_sessions", base)
	var total int64
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []AttendanceSessionSummaryItem
	err := q.db.Table("(?) AS attendance_sessions", base).
		Order("term_id DESC, week_no DESC, weekday DESC, section DESC, course_group_lesson_id DESC").
		Offset((input.Page - 1) * input.PageSize).
		Limit(input.PageSize).
		Scan(&items).Error
	return items, total, err
}

func (q *AttendanceQuery) AvailableCourseGroupLessons(termID uint64, weekday, weekNo int) ([]SessionWithCourse, error) {
	var sessions []SessionWithCourse
	err := q.db.Table("course_group_lesson").
		Select(`
			course_group_lesson.term_id,
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
		Where("course_group_lesson.term_id = ? AND course_group_lesson.weekday = ? AND course_group_lesson.week_no = ? AND course_group_lesson.status = 1 AND course_group.status = 1", termID, weekday, weekNo).
		Order("course_group_lesson.section ASC, course_group_lesson.id ASC").
		Scan(&sessions).Error
	return sessions, err
}

func (q *AttendanceQuery) AvailableCourseGroupLessonsForClass(termID uint64, weekday, weekNo int, classID uint64) ([]SessionWithCourse, error) {
	var sessions []SessionWithCourse
	err := q.db.Table("course_group_lesson").
		Select(`
			course_group_lesson.term_id,
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
		Where("course_group_lesson.term_id = ? AND course_group_lesson.weekday = ? AND course_group_lesson.week_no = ? AND course_group_lesson.status = 1 AND course_group.status = 1", termID, weekday, weekNo).
		Where(`
			EXISTS (
				SELECT 1
				FROM course_group_student
				WHERE course_group_student.course_group_id = course_group.id
				  AND course_group_student.class_id = ?
				  AND course_group_student.status = 1
			)
		`, classID).
		Order("course_group_lesson.section ASC, course_group_lesson.id ASC").
		Scan(&sessions).Error
	return sessions, err
}

func (q *AttendanceQuery) AttendanceSessionRecords(sessionID uint64) ([]AttendanceRecordItem, error) {
	items, _, err := q.AttendanceSessionRecordPage(sessionID, "", "", "", "", "", "", 1, 1000000)
	return items, err
}

func (q *AttendanceQuery) attendanceSessionRecordBase(sessionID uint64, studentID, realName, className, status, operatorName, operatedDate string) *gorm.DB {
	base := q.db.Table("course_group_student").
		Joins("JOIN course_group_lesson ON course_group_lesson.course_group_id = course_group_student.course_group_id AND course_group_lesson.id = ?", sessionID).
		Joins("JOIN student ON student.id = course_group_student.student_id").
		Joins("LEFT JOIN attendance_record ON attendance_record.course_group_lesson_id = course_group_lesson.id AND attendance_record.student_id = course_group_student.student_id").
		Joins("LEFT JOIN user AS operator_user ON operator_user.id = attendance_record.updated_by_user_id").
		Joins("LEFT JOIN class ON class.id = course_group_student.class_id").
		Where("course_group_student.status = 1 AND student.status = 1")
	if studentID != "" {
		base = base.Where("student.student_no LIKE ?", "%"+studentID+"%")
	}
	if realName != "" {
		base = base.Where("student.student_name LIKE ?", "%"+realName+"%")
	}
	if className != "" {
		base = base.Where("COALESCE(class.class_name, '') LIKE ?", "%"+className+"%")
	}
	switch status {
	case "unrecorded":
		base = base.Where("attendance_record.id IS NULL")
	case "0", "1", "2", "3":
		base = base.Where("attendance_record.attendance_status = ?", status)
	}
	if operatorName != "" {
		base = base.Where("operator_user.real_name LIKE ?", "%"+operatorName+"%")
	}
	if operatedDate != "" {
		base = base.Where("DATE(attendance_record.updated_at) = ?", operatedDate)
	}
	return base
}

func (q *AttendanceQuery) AttendanceSessionRecordPage(sessionID uint64, studentID, realName, className, status, operatorName, operatedDate string, page, pageSize int) ([]AttendanceRecordItem, int64, error) {
	var records []AttendanceRecordItem
	base := q.attendanceSessionRecordBase(sessionID, studentID, realName, className, status, operatorName, operatedDate)

	var total int64
	if err := q.db.Table("(?) AS attendance_records", base).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := base.
		Select(`
			student.id AS id,
			attendance_record.id AS attendance_record_id,
			course_group_lesson.id AS course_group_lesson_id,
			student.student_no AS student_id,
			student.student_name AS real_name,
			course_group_student.class_id AS class_id,
			COALESCE(class.class_name, '') AS class_name,
			attendance_record.attendance_status AS status,
			attendance_record.updated_by_user_id AS status_set_by_user_id,
			attendance_record.updated_at AS status_set_at,
			COALESCE(operator_user.real_name, '') AS operator_name,
			attendance_record.updated_at AS operated_at
		`).
		Order("COALESCE(class.class_name, '其他学生') ASC, student.student_no ASC, student.id ASC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&records).Error
	return records, total, err
}

func (q *AttendanceQuery) LocateAttendanceSessionRecordPage(sessionID uint64, studentID, realName, className, status, operatorName, operatedDate string, focusStudentID uint64, pageSize int) (FocusPageResult, error) {
	base := q.attendanceSessionRecordBase(sessionID, studentID, realName, className, status, operatorName, operatedDate)

	var target struct {
		ID            uint64 `gorm:"column:id"`
		StudentNo     string `gorm:"column:student_no"`
		SortClassName string `gorm:"column:sort_class_name"`
	}
	if err := base.Select(`
			student.id AS id,
			student.student_no AS student_no,
			COALESCE(class.class_name, '其他学生') AS sort_class_name
		`).
		Where("student.id = ?", focusStudentID).
		Take(&target).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return FocusPageResult{}, nil
		}
		return FocusPageResult{}, err
	}

	var rowNo int64
	if err := q.attendanceSessionRecordBase(sessionID, studentID, realName, className, status, operatorName, operatedDate).
		Where(`
			COALESCE(class.class_name, '其他学生') < ?
			OR (COALESCE(class.class_name, '其他学生') = ? AND student.student_no < ?)
			OR (COALESCE(class.class_name, '其他学生') = ? AND student.student_no = ? AND student.id <= ?)
		`,
			target.SortClassName,
			target.SortClassName, target.StudentNo,
			target.SortClassName, target.StudentNo, target.ID,
		).
		Count(&rowNo).Error; err != nil {
		return FocusPageResult{}, err
	}
	if rowNo <= 0 {
		return FocusPageResult{}, nil
	}

	return FocusPageResult{
		Found:  true,
		Page:   int((rowNo-1)/int64(pageSize)) + 1,
		RowKey: target.ID,
	}, nil
}

func (q *AttendanceQuery) AttendanceSessionRecordsForClass(sessionID uint64, classID uint64) ([]AttendanceRecordItem, error) {
	var records []AttendanceRecordItem
	err := q.db.Table("course_group_student").
		Select(`
			student.id AS id,
			attendance_record.id AS attendance_record_id,
			course_group_lesson.id AS course_group_lesson_id,
			student.student_no AS student_id,
			student.student_name AS real_name,
			course_group_student.class_id AS class_id,
			COALESCE(class.class_name, '') AS class_name,
			attendance_record.attendance_status AS status,
			attendance_record.updated_by_user_id AS status_set_by_user_id,
			attendance_record.updated_at AS status_set_at
		`).
		Joins("JOIN course_group_lesson ON course_group_lesson.course_group_id = course_group_student.course_group_id AND course_group_lesson.id = ?", sessionID).
		Joins("JOIN student ON student.id = course_group_student.student_id").
		Joins("LEFT JOIN attendance_record ON attendance_record.course_group_lesson_id = course_group_lesson.id AND attendance_record.student_id = course_group_student.student_id").
		Joins("JOIN class ON class.id = course_group_student.class_id").
		Where("course_group_student.class_id = ? AND course_group_student.status = 1 AND student.status = 1", classID).
		Order("student.student_no ASC, student.id ASC").
		Scan(&records).Error
	return records, err
}

func (q *AttendanceQuery) AttendanceClassGroups(checkID uint64) ([]AttendanceClassGroupItem, error) {
	var groups []AttendanceClassGroupItem
	if err := q.db.Table("course_group_student").
		Select("course_group_student.class_id AS class_id, COALESCE(class.class_name, '其他学生') AS class_name, COUNT(course_group_student.id) AS student_count").
		Joins("JOIN course_group_lesson ON course_group_lesson.course_group_id = course_group_student.course_group_id AND course_group_lesson.id = ?", checkID).
		Joins("JOIN student ON student.id = course_group_student.student_id AND student.status = 1").
		Joins("LEFT JOIN class ON class.id = course_group_student.class_id").
		Where("course_group_student.status = 1").
		Group("course_group_student.class_id, class.class_name").
		Order("class.class_name ASC").
		Scan(&groups).Error; err != nil {
		return nil, err
	}
	return groups, nil
}

func (q *AttendanceQuery) AttendanceClassGroupsForClass(checkID uint64, classID uint64) ([]AttendanceClassGroupItem, error) {
	var groups []AttendanceClassGroupItem
	err := q.db.Table("course_group_student").
		Select("class.id AS class_id, class.class_name, COUNT(course_group_student.id) AS student_count").
		Joins("JOIN course_group_lesson ON course_group_lesson.course_group_id = course_group_student.course_group_id AND course_group_lesson.id = ?", checkID).
		Joins("JOIN student ON student.id = course_group_student.student_id AND student.status = 1").
		Joins("JOIN class ON class.id = course_group_student.class_id").
		Where("course_group_student.class_id = ? AND course_group_student.status = 1", classID).
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
			CASE
				WHEN attendance_record_log.old_attendance_status IS NULL THEN 'create_record'
				ELSE 'set_status'
			END AS operation_type,
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
