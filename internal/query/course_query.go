package query

import (
	"sort"
	"time"

	"gorm.io/gorm"

	"wack-backend/internal/model"
)

type CourseStudentItem struct {
	ID        uint64    `json:"id"`
	CourseID  uint64    `json:"course_id"`
	StudentID string    `json:"student_id"`
	RealName  string    `json:"real_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CourseCalendarItem struct {
	model.CourseSession
	Term        string   `json:"term"`
	CourseName  string   `json:"course_name"`
	TeacherName string   `json:"teacher_name"`
	ClassNames  []string `gorm:"-" json:"class_names"`
	ClassIDs    []uint64 `gorm:"-" json:"class_ids"`
	Grades      []int    `gorm:"-" json:"grades"`
	MajorNames  []string `gorm:"-" json:"major_names"`
}

type CourseListItem struct {
	model.Course
	ClassNames []string `gorm:"-" json:"class_names"`
	ClassIDs   []uint64 `gorm:"-" json:"class_ids"`
}

type courseCalendarClassRow struct {
	CourseID   uint64
	ClassID    uint64
	ClassName  string
	Grade      int
	MajorName  string
}

type CourseQuery struct {
	db *gorm.DB
}

func NewCourseQuery(db *gorm.DB) *CourseQuery {
	return &CourseQuery{db: db}
}

func (q *CourseQuery) CourseStudents(courseID uint64) ([]CourseStudentItem, error) {
	var students []CourseStudentItem
	err := q.db.Table("course_student").
		Select("course_student.id, course_student.course_id, course_student.student_id, course_student.real_name, course_student.created_at, course_student.updated_at").
		Where("course_student.course_id = ?", courseID).
		Find(&students).Error
	return students, err
}

func (q *CourseQuery) ListCourses(term, teacher, keyword string, page, pageSize int) ([]CourseListItem, int64, error) {
	queryDB := q.db.Model(&model.Course{})
	if term != "" {
		queryDB = queryDB.Where("term = ?", term)
	}
	if teacher != "" {
		queryDB = queryDB.Where("teacher_name LIKE ?", "%"+teacher+"%")
	}
	if keyword != "" {
		queryDB = queryDB.Where("course_name LIKE ?", "%"+keyword+"%")
	}
	var total int64
	if err := queryDB.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var courses []CourseListItem
	if err := queryDB.Order("id DESC").Offset((page-1)*pageSize).Limit(pageSize).Find(&courses).Error; err != nil {
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
	if err := q.db.Table("course_class").
		Select("course_class.course_id, class.id AS class_id, class.class_name, class.grade, class.major_name").
		Joins("JOIN class ON class.id = course_class.class_id").
		Where("course_class.course_id IN ?", courseIDs).
		Order("class.grade DESC, class.class_name ASC").
		Scan(&classRows).Error; err != nil {
		return nil, 0, err
	}

	classNamesByCourseID := make(map[uint64][]string, len(courseIDs))
	classIDsByCourseID := make(map[uint64][]uint64, len(courseIDs))
	for _, row := range classRows {
		classNamesByCourseID[row.CourseID] = append(classNamesByCourseID[row.CourseID], row.ClassName)
		classIDsByCourseID[row.CourseID] = append(classIDsByCourseID[row.CourseID], row.ClassID)
	}

	for index := range courses {
		courses[index].ClassNames = dedupeStrings(classNamesByCourseID[courses[index].ID])
		courses[index].ClassIDs = dedupeUint64s(classIDsByCourseID[courses[index].ID])
	}

	return courses, total, nil
}

func (q *CourseQuery) CourseCalendar(weekNo, term string) ([]CourseCalendarItem, error) {
	query := q.db.Model(&model.CourseSession{}).
		Joins("JOIN course ON course.id = course_session.course_id")
	if weekNo != "" {
		query = query.Where("week_no = ?", weekNo)
	}
	if term != "" {
		query = query.Where("course.term = ?", term)
	}
	var items []CourseCalendarItem
	err := query.Select("course_session.*, course.term, course.course_name, course.teacher_name").
		Order("week_no, weekday, section, session_no").
		Scan(&items).Error
	if err != nil || len(items) == 0 {
		return items, err
	}

	courseIDs := make([]uint64, 0, len(items))
	seenCourseIDs := make(map[uint64]struct{}, len(items))
	for _, item := range items {
		if _, exists := seenCourseIDs[item.CourseID]; exists {
			continue
		}
		seenCourseIDs[item.CourseID] = struct{}{}
		courseIDs = append(courseIDs, item.CourseID)
	}

	var classRows []courseCalendarClassRow
	if err := q.db.Table("course_class").
		Select("course_class.course_id, class.id AS class_id, class.class_name, class.grade, class.major_name").
		Joins("JOIN class ON class.id = course_class.class_id").
		Where("course_class.course_id IN ?", courseIDs).
		Order("class.grade DESC, class.class_name ASC").
		Scan(&classRows).Error; err != nil {
		return nil, err
	}

	classNamesByCourseID := make(map[uint64][]string, len(courseIDs))
	classIDsByCourseID := make(map[uint64][]uint64, len(courseIDs))
	gradesByCourseID := make(map[uint64][]int, len(courseIDs))
	majorNamesByCourseID := make(map[uint64][]string, len(courseIDs))
	for _, row := range classRows {
		classNamesByCourseID[row.CourseID] = append(classNamesByCourseID[row.CourseID], row.ClassName)
		classIDsByCourseID[row.CourseID] = append(classIDsByCourseID[row.CourseID], row.ClassID)
		gradesByCourseID[row.CourseID] = append(gradesByCourseID[row.CourseID], row.Grade)
		majorNamesByCourseID[row.CourseID] = append(majorNamesByCourseID[row.CourseID], row.MajorName)
	}

	for index := range items {
		items[index].ClassNames = dedupeStrings(classNamesByCourseID[items[index].CourseID])
		items[index].ClassIDs = dedupeUint64s(classIDsByCourseID[items[index].CourseID])
		items[index].Grades = dedupeInts(gradesByCourseID[items[index].CourseID])
		items[index].MajorNames = dedupeStrings(majorNamesByCourseID[items[index].CourseID])
	}

	return items, nil
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
