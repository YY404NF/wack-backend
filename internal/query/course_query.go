package query

import (
	"sort"
	"strconv"
	"time"

	"gorm.io/gorm"

	"wack-backend/internal/model"
)

type CourseCalendarItem struct {
	ID                  uint64   `json:"id"`
	CourseGroupID       uint64   `json:"course_group_id"`
	CourseID            uint64   `json:"course_id"`
	SessionNo           int      `json:"session_no"`
	Term                string   `json:"term"`
	WeekNo              int      `json:"week_no"`
	Weekday             int      `json:"weekday"`
	Section             int      `json:"section"`
	BuildingName        string   `json:"building_name"`
	RoomName            string   `json:"room_name"`
	CourseName          string   `json:"course_name"`
	TeacherName         string   `json:"teacher_name"`
	HasAttendanceRecord bool     `json:"has_attendance_record"`
	AttendanceRate      float64  `json:"attendance_rate"`
	StudentCount        int64    `json:"student_count"`
	RecordCount         int64    `json:"record_count"`
	LateCount           int64    `json:"late_count"`
	AbsentCount         int64    `json:"absent_count"`
	LeaveCount          int64    `json:"leave_count"`
	ClassNames          []string `gorm:"-" json:"class_names"`
	ClassIDs            []uint64 `gorm:"-" json:"class_ids"`
	Grades              []int    `gorm:"-" json:"grades"`
	MajorNames          []string `gorm:"-" json:"major_names"`
}

type CourseCalendarOutlineItem struct {
	CourseGroupID uint64   `json:"course_group_id"`
	CourseID      uint64   `json:"course_id"`
	Weekday       int      `json:"weekday"`
	Section       int      `json:"section"`
	CourseName    string   `json:"course_name"`
	TeacherName   string   `json:"teacher_name"`
	WeekNos       []int    `gorm:"-" json:"week_nos"`
	Locations     []string `gorm:"-" json:"locations"`
	ClassNames    []string `gorm:"-" json:"class_names"`
}

type CourseListItem struct {
	model.Course
	ClassNames []string `gorm:"-" json:"class_names"`
	ClassIDs   []uint64 `gorm:"-" json:"class_ids"`
}

type CourseGroupListItem struct {
	model.CourseGroup
	ClassNames   []string `gorm:"-" json:"class_names"`
	ClassIDs     []uint64 `gorm:"-" json:"class_ids"`
	StudentCount int64    `gorm:"-" json:"student_count"`
	LessonCount  int64    `gorm:"-" json:"lesson_count"`
}

type CourseGroupStudentItem struct {
	ID            uint64    `json:"id"`
	TermID        uint64    `json:"term_id"`
	CourseGroupID uint64    `json:"course_group_id"`
	StudentID     uint64    `json:"student_id"`
	ClassID       *uint64   `json:"class_id"`
	Status        int       `json:"status"`
	StudentNo     string    `json:"student_no"`
	StudentName   string    `json:"student_name"`
	ClassName     *string   `json:"class_name"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type AvailableCourseGroupClassItem struct {
	ID           uint64 `json:"id"`
	ClassName    string `json:"class_name"`
	Grade        int    `json:"grade"`
	MajorName    string `json:"major_name"`
	StudentCount int64  `json:"student_count"`
}

type AvailableCourseGroupStudentItem struct {
	ID          uint64 `json:"id"`
	ClassID     uint64 `json:"class_id"`
	StudentNo   string `json:"student_no"`
	StudentName string `json:"student_name"`
	ClassName   string `json:"class_name"`
	Grade       int    `json:"grade"`
	MajorName   string `json:"major_name"`
}

type courseCalendarClassRow struct {
	CourseID  uint64
	ClassID   uint64
	ClassName string
	Grade     int
	MajorName string
}

type courseCalendarGroupClassRow struct {
	CourseGroupID uint64
	ClassID       uint64
	ClassName     string
	Grade         int
	MajorName     string
}

type courseCalendarOutlineRow struct {
	CourseGroupID uint64
	CourseID      uint64
	Weekday       int
	Section       int
	WeekNo        int
	BuildingName  string
	RoomName      string
	CourseName    string
	TeacherName   string
}

type courseGroupClassRow struct {
	CourseGroupID uint64
	ClassID       uint64
	ClassName     string
}

type courseGroupCountRow struct {
	CourseGroupID uint64
	Count         int64
}

type courseCountRow struct {
	CourseID uint64
	Count    int64
}

type CourseQuery struct {
	db *gorm.DB
}

func NewCourseQuery(db *gorm.DB) *CourseQuery {
	return &CourseQuery{db: db}
}

func (q *CourseQuery) courseListBaseQuery(term, grade, teacher, keyword, className, studentCount string) (*gorm.DB, error) {
	queryDB := q.db.Table("course").
		Joins("JOIN term ON term.id = course.term_id").
		Where("course.status = 1")
	if term != "" {
		queryDB = queryDB.Where("term.name = ?", term)
	}
	if grade != "" {
		queryDB = queryDB.Where("CAST(course.grade AS TEXT) LIKE ?", "%"+grade+"%")
	}
	if teacher != "" {
		queryDB = queryDB.Where("course.teacher_name LIKE ?", "%"+teacher+"%")
	}
	if keyword != "" {
		queryDB = queryDB.Where("course.course_name LIKE ?", "%"+keyword+"%")
	}
	if className != "" {
		queryDB = queryDB.Where(`
			EXISTS (
				SELECT 1
				FROM course_group_student
				JOIN course_group ON course_group.id = course_group_student.course_group_id
				JOIN class ON class.id = course_group_student.class_id
				WHERE course_group.course_id = course.id
				  AND course_group.status = 1
				  AND course_group_student.status = 1
				  AND class.status = 1
				  AND class.class_name LIKE ?
			)
		`, "%"+className+"%")
	}
	if studentCount != "" {
		studentCountValue, err := strconv.Atoi(studentCount)
		if err != nil {
			return nil, nil
		}
		queryDB = queryDB.Where(`
			(
				SELECT COUNT(DISTINCT course_group_student.student_id)
				FROM course_group_student
				JOIN course_group ON course_group.id = course_group_student.course_group_id
				WHERE course_group.course_id = course.id
				  AND course_group.status = 1
				  AND course_group_student.status = 1
			) = ?
		`, studentCountValue)
	}
	return queryDB, nil
}

func (q *CourseQuery) CourseGroups(courseID uint64) ([]CourseGroupListItem, error) {
	var groups []CourseGroupListItem
	if err := q.db.Model(&model.CourseGroup{}).
		Where("course_id = ? AND status = 1", courseID).
		Order("id DESC").
		Find(&groups).Error; err != nil {
		return nil, err
	}
	if len(groups) == 0 {
		return groups, nil
	}

	groupIDs := make([]uint64, 0, len(groups))
	for _, item := range groups {
		groupIDs = append(groupIDs, item.ID)
	}

	var classRows []courseGroupClassRow
	if err := q.db.Table("course_group_student").
		Select("course_group_student.course_group_id, class.id AS class_id, class.class_name").
		Joins("JOIN class ON class.id = course_group_student.class_id").
		Where("course_group_student.course_group_id IN ? AND course_group_student.status = 1 AND course_group_student.class_id IS NOT NULL", groupIDs).
		Order("class.grade DESC, class.class_name ASC").
		Scan(&classRows).Error; err != nil {
		return nil, err
	}

	var studentCountRows []courseGroupCountRow
	if err := q.db.Table("course_group_student").
		Select("course_group_id, COUNT(*) AS count").
		Where("course_group_id IN ? AND status = 1", groupIDs).
		Group("course_group_id").
		Scan(&studentCountRows).Error; err != nil {
		return nil, err
	}

	var lessonCountRows []courseGroupCountRow
	if err := q.db.Table("course_group_lesson").
		Select("course_group_id, COUNT(*) AS count").
		Where("course_group_id IN ? AND status = 1", groupIDs).
		Group("course_group_id").
		Scan(&lessonCountRows).Error; err != nil {
		return nil, err
	}

	classNamesByGroupID := make(map[uint64][]string, len(groups))
	classIDsByGroupID := make(map[uint64][]uint64, len(groups))
	for _, row := range classRows {
		classNamesByGroupID[row.CourseGroupID] = append(classNamesByGroupID[row.CourseGroupID], row.ClassName)
		classIDsByGroupID[row.CourseGroupID] = append(classIDsByGroupID[row.CourseGroupID], row.ClassID)
	}

	studentCountByGroupID := make(map[uint64]int64, len(studentCountRows))
	for _, row := range studentCountRows {
		studentCountByGroupID[row.CourseGroupID] = row.Count
	}

	lessonCountByGroupID := make(map[uint64]int64, len(lessonCountRows))
	for _, row := range lessonCountRows {
		lessonCountByGroupID[row.CourseGroupID] = row.Count
	}

	for index := range groups {
		groups[index].ClassNames = dedupeStrings(classNamesByGroupID[groups[index].ID])
		groups[index].ClassIDs = dedupeUint64s(classIDsByGroupID[groups[index].ID])
		groups[index].StudentCount = studentCountByGroupID[groups[index].ID]
		groups[index].LessonCount = lessonCountByGroupID[groups[index].ID]
	}

	return groups, nil
}

func (q *CourseQuery) CourseGroup(courseID, groupID uint64) (model.CourseGroup, error) {
	var group model.CourseGroup
	err := q.db.Where("course_id = ? AND id = ? AND status = 1", courseID, groupID).First(&group).Error
	return group, err
}

func (q *CourseQuery) CourseGroupStudents(groupID uint64) ([]CourseGroupStudentItem, error) {
	var students []CourseGroupStudentItem
	err := q.db.Table("course_group_student").
		Select(`course_group_student.id, course_group_student.term_id, course_group_student.course_group_id,
			course_group_student.student_id, course_group_student.class_id, course_group_student.status,
			course_group_student.created_at, course_group_student.updated_at,
			student.student_no, student.student_name, class.class_name`).
		Joins("JOIN student ON student.id = course_group_student.student_id").
		Joins("LEFT JOIN class ON class.id = course_group_student.class_id").
		Where("course_group_student.course_group_id = ? AND course_group_student.status = 1", groupID).
		Order("class.class_name IS NULL, class.class_name ASC, student.student_no ASC, student.id ASC").
		Scan(&students).Error
	return students, err
}

func (q *CourseQuery) CourseGroupLessons(groupID uint64) ([]model.CourseGroupLesson, error) {
	var lessons []model.CourseGroupLesson
	err := q.db.Where("course_group_id = ? AND status = 1", groupID).
		Order("week_no ASC, weekday ASC, section ASC, id ASC").
		Find(&lessons).Error
	return lessons, err
}

func (q *CourseQuery) AvailableCourseGroupClasses(groupID uint64, className string, page, pageSize int) ([]AvailableCourseGroupClassItem, int64, error) {
	query := q.db.Table("class").
		Select("class.id, class.class_name, class.grade, class.major_name, COUNT(student.id) AS student_count").
		Joins("LEFT JOIN student ON student.class_id = class.id AND student.status = 1").
		Where("class.status = 1").
		Where("class.id NOT IN (?)",
			q.db.Table("course_group_student").
				Select("class_id").
				Where("course_group_id = ? AND status = 1 AND class_id IS NOT NULL", groupID),
		).
		Group("class.id")
	if className != "" {
		query = query.Where("class.class_name LIKE ?", "%"+className+"%")
	}
	var items []AvailableCourseGroupClassItem
	query = query.Group("class.id")

	var total int64
	if err := q.db.Table("(?) AS available_classes", query).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("class.id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&items).Error
	return items, total, err
}

func (q *CourseQuery) AvailableCourseGroupStudents(groupID uint64, studentNo, studentName, className string, page, pageSize int) ([]AvailableCourseGroupStudentItem, int64, error) {
	query := q.db.Table("student").
		Select("student.id, student.class_id, student.student_no, student.student_name, class.class_name, class.grade, class.major_name").
		Joins("JOIN class ON class.id = student.class_id").
		Where("student.status = 1 AND class.status = 1").
		Where("student.id NOT IN (?)",
			q.db.Table("course_group_student").
				Select("student_id").
				Where("course_group_id = ? AND status = 1", groupID),
		)
	if studentNo != "" {
		query = query.Where("student.student_no LIKE ?", "%"+studentNo+"%")
	}
	if studentName != "" {
		query = query.Where("student.student_name LIKE ?", "%"+studentName+"%")
	}
	if className != "" {
		query = query.Where("class.class_name LIKE ?", "%"+className+"%")
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []AvailableCourseGroupStudentItem
	err := query.Order("class.id DESC, student.student_no ASC, student.id ASC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&items).Error
	return items, total, err
}

func (q *CourseQuery) ListCourses(term, grade, teacher, keyword, className, studentCount string, page, pageSize int) ([]CourseListItem, int64, error) {
	queryDB, err := q.courseListBaseQuery(term, grade, teacher, keyword, className, studentCount)
	if err != nil {
		return nil, 0, err
	}
	if queryDB == nil {
		return []CourseListItem{}, 0, nil
	}
	var total int64
	if err := queryDB.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var courses []CourseListItem
	if err := queryDB.
		Select(`course.id,
			course.term_id,
			course.grade,
			term.name AS term,
			course.course_name,
			course.teacher_name,
			course.status,
			0 AS student_count,
			course.created_at,
			course.updated_at`).
		Order("course.id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&courses).Error; err != nil {
		return nil, 0, err
	}
	if len(courses) == 0 {
		return courses, total, nil
	}

	courseIDs := make([]uint64, 0, len(courses))
	for _, item := range courses {
		courseIDs = append(courseIDs, item.ID)
	}

	var classRows []courseCalendarClassRow
	if err := q.db.Table("course_group_student").
		Select("course_group.course_id, class.id AS class_id, class.class_name, class.grade, class.major_name").
		Joins("JOIN course_group ON course_group.id = course_group_student.course_group_id").
		Joins("JOIN class ON class.id = course_group_student.class_id").
		Where("course_group.course_id IN ? AND course_group.status = 1 AND course_group_student.status = 1 AND course_group_student.class_id IS NOT NULL", courseIDs).
		Order("class.grade DESC, class.class_name ASC").
		Scan(&classRows).Error; err != nil {
		return nil, 0, err
	}

	var studentCountRows []courseCountRow
	if err := q.db.Table("course_group_student").
		Select("course_group.course_id, COUNT(DISTINCT course_group_student.student_id) AS count").
		Joins("JOIN course_group ON course_group.id = course_group_student.course_group_id").
		Where("course_group.course_id IN ? AND course_group.status = 1 AND course_group_student.status = 1", courseIDs).
		Group("course_group.course_id").
		Scan(&studentCountRows).Error; err != nil {
		return nil, 0, err
	}

	classNamesByCourseID := make(map[uint64][]string, len(courseIDs))
	classIDsByCourseID := make(map[uint64][]uint64, len(courseIDs))
	for _, row := range classRows {
		classNamesByCourseID[row.CourseID] = append(classNamesByCourseID[row.CourseID], row.ClassName)
		classIDsByCourseID[row.CourseID] = append(classIDsByCourseID[row.CourseID], row.ClassID)
	}

	studentCountByCourseID := make(map[uint64]int64, len(studentCountRows))
	for _, row := range studentCountRows {
		studentCountByCourseID[row.CourseID] = row.Count
	}

	for index := range courses {
		courses[index].ClassNames = dedupeStrings(classNamesByCourseID[courses[index].ID])
		courses[index].ClassIDs = dedupeUint64s(classIDsByCourseID[courses[index].ID])
		courses[index].StudentCount = int(studentCountByCourseID[courses[index].ID])
	}

	return courses, total, nil
}

func (q *CourseQuery) LocateCoursePage(term, grade, teacher, keyword, className, studentCount string, focusCourseID uint64, pageSize int) (FocusPageResult, error) {
	base, err := q.courseListBaseQuery(term, grade, teacher, keyword, className, studentCount)
	if err != nil {
		return FocusPageResult{}, err
	}
	if base == nil {
		return FocusPageResult{}, nil
	}

	var target struct {
		ID uint64 `gorm:"column:id"`
	}
	if err := base.Select("course.id AS id").Where("course.id = ?", focusCourseID).Take(&target).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return FocusPageResult{}, nil
		}
		return FocusPageResult{}, err
	}

	var rowNo int64
	if err := q.db.Table("(?) AS focused_courses", base.Select("course.id AS id")).
		Where("id >= ?", target.ID).
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

func (q *CourseQuery) CourseCalendar(weekNo, term string) ([]CourseCalendarItem, error) {
	query := q.db.Table("course_group_lesson").
		Joins("JOIN course_group ON course_group.id = course_group_lesson.course_group_id").
		Joins("JOIN course ON course.id = course_group.course_id").
		Joins("JOIN term ON term.id = course_group.term_id").
		Where("course_group_lesson.status = 1 AND course_group.status = 1")
	if weekNo != "" {
		query = query.Where("course_group_lesson.week_no = ?", weekNo)
	}
	if term != "" {
		query = query.Where("term.name = ?", term)
	}
	var items []CourseCalendarItem
	err := query.Select(`course_group_lesson.id,
			course_group.id AS course_group_id,
			course_group.course_id,
			ROW_NUMBER() OVER (
				PARTITION BY course_group.id
				ORDER BY course_group_lesson.week_no, course_group_lesson.weekday, course_group_lesson.section, course_group_lesson.id
			) AS session_no,
			term.name AS term,
			course_group_lesson.week_no,
			course_group_lesson.weekday,
			course_group_lesson.section,
			course_group_lesson.building_name,
			course_group_lesson.room_name,
			course.course_name,
			course.teacher_name,
			EXISTS (
				SELECT 1
				FROM attendance_record
				WHERE attendance_record.course_group_lesson_id = course_group_lesson.id
			) AS has_attendance_record,
			(
				SELECT COUNT(1)
				FROM course_group_student AS cgs
				JOIN student ON student.id = cgs.student_id AND student.status = 1
				WHERE cgs.course_group_id = course_group.id
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
			) AS leave_count,
			COALESCE((
				SELECT CASE
					WHEN SUM(CASE WHEN attendance_record.attendance_status IN (0, 1, 2) THEN 1 ELSE 0 END) = 0 THEN 1
					ELSE CAST(SUM(CASE WHEN attendance_record.attendance_status IN (0, 1) THEN 1 ELSE 0 END) AS REAL)
						/ SUM(CASE WHEN attendance_record.attendance_status IN (0, 1, 2) THEN 1 ELSE 0 END)
				END
				FROM attendance_record
				WHERE attendance_record.course_group_lesson_id = course_group_lesson.id
			), 0) AS attendance_rate`).
		Order("course_group_lesson.week_no, course_group_lesson.weekday, course_group_lesson.section, session_no").
		Scan(&items).Error
	if err != nil || len(items) == 0 {
		return items, err
	}

	classNamesByCourseGroupID, classIDsByCourseGroupID, gradesByCourseGroupID, majorNamesByCourseGroupID, err := q.courseCalendarGroupClassMaps(courseGroupIDsFromCalendarItems(items))
	if err != nil {
		return nil, err
	}

	for index := range items {
		items[index].ClassNames = dedupeStrings(classNamesByCourseGroupID[items[index].CourseGroupID])
		items[index].ClassIDs = dedupeUint64s(classIDsByCourseGroupID[items[index].CourseGroupID])
		items[index].Grades = dedupeInts(gradesByCourseGroupID[items[index].CourseGroupID])
		items[index].MajorNames = dedupeStrings(majorNamesByCourseGroupID[items[index].CourseGroupID])
	}

	return items, nil
}

func (q *CourseQuery) CourseCalendarOutline(term string) ([]CourseCalendarOutlineItem, error) {
	query := q.db.Table("course_group_lesson").
		Joins("JOIN course_group ON course_group.id = course_group_lesson.course_group_id").
		Joins("JOIN course ON course.id = course_group.course_id").
		Joins("JOIN term ON term.id = course_group.term_id").
		Where("course_group_lesson.status = 1 AND course_group.status = 1")
	if term != "" {
		query = query.Where("term.name = ?", term)
	}

	var rows []courseCalendarOutlineRow
	if err := query.Select(`
			course_group.id AS course_group_id,
			course_group.course_id,
			course_group_lesson.weekday,
			course_group_lesson.section,
			course_group_lesson.week_no,
			course_group_lesson.building_name,
			course_group_lesson.room_name,
			course.course_name,
			course.teacher_name`).
		Order("course_group_lesson.weekday, course_group_lesson.section, course.course_name, course_group.id, course_group_lesson.week_no").
		Scan(&rows).Error; err != nil || len(rows) == 0 {
		return nil, err
	}

	classNamesByCourseGroupID, _, _, _, err := q.courseCalendarGroupClassMaps(courseGroupIDsFromOutlineRows(rows))
	if err != nil {
		return nil, err
	}

	grouped := make(map[string]*CourseCalendarOutlineItem, len(rows))
	order := make([]string, 0, len(rows))
	for _, row := range rows {
		key := strconv.FormatUint(row.CourseGroupID, 10) + ":" + strconv.Itoa(row.Weekday) + ":" + strconv.Itoa(row.Section)
		item, exists := grouped[key]
		if !exists {
			item = &CourseCalendarOutlineItem{
				CourseGroupID: row.CourseGroupID,
				CourseID:      row.CourseID,
				Weekday:       row.Weekday,
				Section:       row.Section,
				CourseName:    row.CourseName,
				TeacherName:   row.TeacherName,
				ClassNames:    dedupeStrings(classNamesByCourseGroupID[row.CourseGroupID]),
			}
			grouped[key] = item
			order = append(order, key)
		}
		item.WeekNos = append(item.WeekNos, row.WeekNo)
		item.Locations = append(item.Locations, row.BuildingName+"-"+row.RoomName)
	}

	items := make([]CourseCalendarOutlineItem, 0, len(order))
	for _, key := range order {
		item := grouped[key]
		item.WeekNos = dedupeInts(item.WeekNos)
		sort.Ints(item.WeekNos)
		item.Locations = dedupeStrings(item.Locations)
		sort.Strings(item.Locations)
		items = append(items, *item)
	}
	return items, nil
}

func courseGroupIDsFromCalendarItems(items []CourseCalendarItem) []uint64 {
	ids := make([]uint64, 0, len(items))
	seen := make(map[uint64]struct{}, len(items))
	for _, item := range items {
		if _, exists := seen[item.CourseGroupID]; exists {
			continue
		}
		seen[item.CourseGroupID] = struct{}{}
		ids = append(ids, item.CourseGroupID)
	}
	return ids
}

func courseGroupIDsFromOutlineRows(rows []courseCalendarOutlineRow) []uint64 {
	ids := make([]uint64, 0, len(rows))
	seen := make(map[uint64]struct{}, len(rows))
	for _, row := range rows {
		if _, exists := seen[row.CourseGroupID]; exists {
			continue
		}
		seen[row.CourseGroupID] = struct{}{}
		ids = append(ids, row.CourseGroupID)
	}
	return ids
}

func (q *CourseQuery) courseCalendarGroupClassMaps(courseGroupIDs []uint64) (map[uint64][]string, map[uint64][]uint64, map[uint64][]int, map[uint64][]string, error) {
	classNamesByCourseGroupID := make(map[uint64][]string, len(courseGroupIDs))
	classIDsByCourseGroupID := make(map[uint64][]uint64, len(courseGroupIDs))
	gradesByCourseGroupID := make(map[uint64][]int, len(courseGroupIDs))
	majorNamesByCourseGroupID := make(map[uint64][]string, len(courseGroupIDs))
	if len(courseGroupIDs) == 0 {
		return classNamesByCourseGroupID, classIDsByCourseGroupID, gradesByCourseGroupID, majorNamesByCourseGroupID, nil
	}

	var classRows []courseCalendarGroupClassRow
	if err := q.db.Table("course_group_student").
		Select("DISTINCT course_group.id AS course_group_id, class.id AS class_id, class.class_name, class.grade, class.major_name").
		Joins("JOIN course_group ON course_group.id = course_group_student.course_group_id").
		Joins("JOIN class ON class.id = course_group_student.class_id").
		Where("course_group.id IN ? AND course_group.status = 1 AND course_group_student.status = 1 AND course_group_student.class_id IS NOT NULL", courseGroupIDs).
		Order("class.grade DESC, class.class_name ASC").
		Scan(&classRows).Error; err != nil {
		return nil, nil, nil, nil, err
	}

	for _, row := range classRows {
		classNamesByCourseGroupID[row.CourseGroupID] = append(classNamesByCourseGroupID[row.CourseGroupID], row.ClassName)
		classIDsByCourseGroupID[row.CourseGroupID] = append(classIDsByCourseGroupID[row.CourseGroupID], row.ClassID)
		gradesByCourseGroupID[row.CourseGroupID] = append(gradesByCourseGroupID[row.CourseGroupID], row.Grade)
		majorNamesByCourseGroupID[row.CourseGroupID] = append(majorNamesByCourseGroupID[row.CourseGroupID], row.MajorName)
	}

	return classNamesByCourseGroupID, classIDsByCourseGroupID, gradesByCourseGroupID, majorNamesByCourseGroupID, nil
}

func dedupeStrings(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func dedupeUint64s(values []uint64) []uint64 {
	if len(values) == 0 {
		return []uint64{}
	}
	seen := make(map[uint64]struct{}, len(values))
	result := make([]uint64, 0, len(values))
	for _, value := range values {
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	sort.Slice(result, func(left, right int) bool {
		return result[left] < result[right]
	})
	return result
}

func dedupeInts(values []int) []int {
	if len(values) == 0 {
		return []int{}
	}
	seen := make(map[int]struct{}, len(values))
	result := make([]int, 0, len(values))
	for _, value := range values {
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	sort.Ints(result)
	return result
}
