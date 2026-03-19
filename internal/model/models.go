package model

import "time"

const (
	RoleAdmin        = 1
	RoleStudent      = 2
	RoleCommissioner = 3

	UserStatusActive  = 1
	UserStatusFrozen  = 2
	AttendanceUnset   = 0
	AttendancePresent = 1
	AttendanceLate    = 2
	AttendanceAbsent  = 3
	AttendanceOnLeave = 4
)

type User struct {
	ID             uint64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	LoginID        string     `gorm:"column:login_id;size:32;not null;uniqueIndex" json:"login_id"`
	PasswordHash   string     `gorm:"column:password_hash;size:255;not null" json:"-"`
	RealName       string     `gorm:"column:real_name;size:50;not null" json:"real_name"`
	Role           int        `gorm:"column:role;not null;index:idx_role_status" json:"role"`
	ManagedClassID *uint64    `gorm:"column:managed_class_id;index:idx_managed_class_id" json:"managed_class_id"`
	Status         int        `gorm:"column:status;not null;default:1;index:idx_role_status" json:"status"`
	LastLoginAt    *time.Time `gorm:"column:last_login_at" json:"last_login_at"`
	CreatedAt      time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (User) TableName() string { return "user" }

type Term struct {
	ID            uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name          string    `gorm:"column:name;size:20;not null;uniqueIndex" json:"name"`
	TermStartDate string    `gorm:"column:term_start_date;size:10;not null" json:"term_start_date"`
	CreatedAt     time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (Term) TableName() string { return "term" }

type Class struct {
	ID           uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ClassName    string    `gorm:"column:class_name;size:100;not null;index:idx_grade_major" json:"class_name"`
	Grade        int       `gorm:"column:grade;not null;index:idx_grade_major" json:"grade"`
	MajorName    string    `gorm:"column:major_name;size:100;not null;index:idx_grade_major" json:"major_name"`
	Status       int       `gorm:"column:status;not null;default:1;index:idx_status" json:"status"`
	StudentCount int64     `gorm:"-" json:"student_count"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (Class) TableName() string { return "class" }

type Student struct {
	ID          uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	StudentNo   string    `gorm:"column:student_no;size:32;not null;uniqueIndex" json:"student_no"`
	StudentName string    `gorm:"column:student_name;size:50;not null;index:idx_student_name" json:"student_name"`
	ClassID     *uint64   `gorm:"column:class_id;index:idx_student_class_id;index:idx_student_class_status" json:"class_id"`
	Status      int       `gorm:"column:status;not null;default:1;index:idx_student_status;index:idx_student_class_status" json:"status"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (Student) TableName() string { return "student" }

type UserFreeTime struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	TermID    uint64    `gorm:"column:term_id;not null;uniqueIndex:uk_term_user_time;index:idx_term_user" json:"term_id"`
	UserID    uint64    `gorm:"column:user_id;not null;uniqueIndex:uk_term_user_time;index:idx_term_user" json:"user_id"`
	Weekday   int       `gorm:"column:weekday;not null;uniqueIndex:uk_term_user_time" json:"weekday"`
	Section   int       `gorm:"column:section;not null;uniqueIndex:uk_term_user_time" json:"section"`
	FreeWeeks string    `gorm:"column:free_weeks;size:100;not null" json:"free_weeks"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (UserFreeTime) TableName() string { return "user_free_time" }

type SystemSetting struct {
	ID                   uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	CurrentTermStartDate string    `gorm:"column:current_term_start_date;size:10;not null" json:"current_term_start_date"`
	CurrentSchedule      string    `gorm:"column:current_schedule;size:20;not null" json:"current_schedule"`
	CreatedAt            time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt            time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (SystemSetting) TableName() string { return "system_setting" }

type Course struct {
	ID           uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	TermID       uint64    `gorm:"column:term_id;not null;index:idx_term_grade;index:idx_term_course_name;index:idx_term_teacher_name;index:idx_term_status" json:"term_id"`
	Grade        int       `gorm:"column:grade;not null;index:idx_term_grade" json:"grade"`
	CourseName   string    `gorm:"column:course_name;size:100;not null;index:idx_term_course_name" json:"course_name"`
	TeacherName  string    `gorm:"column:teacher_name;size:50;not null;index:idx_term_teacher_name" json:"teacher_name"`
	Status       int       `gorm:"column:status;not null;default:1;index:idx_term_status" json:"status"`
	Term         string    `gorm:"-" json:"term"`
	StudentCount int       `gorm:"-" json:"student_count"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (Course) TableName() string { return "course" }

type CourseGroup struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	TermID    uint64    `gorm:"column:term_id;not null;index:idx_term_course_id;index:idx_term_course_status" json:"term_id"`
	CourseID  uint64    `gorm:"column:course_id;not null;index:idx_term_course_id;index:idx_term_course_status;index:idx_course_id" json:"course_id"`
	Status    int       `gorm:"column:status;not null;default:1;index:idx_term_course_status;index:idx_status" json:"status"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (CourseGroup) TableName() string { return "course_group" }

type CourseGroupStudent struct {
	ID            uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	TermID        uint64    `gorm:"column:term_id;not null;uniqueIndex:uk_term_group_student;index:idx_term_group_id;index:idx_term_student_id;index:idx_term_class_id;index:idx_term_group_class_id;index:idx_term_group_status;index:idx_term_student_status" json:"term_id"`
	CourseGroupID uint64    `gorm:"column:course_group_id;not null;uniqueIndex:uk_term_group_student;index:idx_term_group_id;index:idx_term_group_class_id;index:idx_term_group_status" json:"course_group_id"`
	StudentID     uint64    `gorm:"column:student_id;not null;uniqueIndex:uk_term_group_student;index:idx_term_student_id;index:idx_term_student_status" json:"student_id"`
	ClassID       *uint64   `gorm:"column:class_id;index:idx_term_class_id;index:idx_term_group_class_id" json:"class_id"`
	Status        int       `gorm:"column:status;not null;default:1;index:idx_term_group_status;index:idx_term_student_status" json:"status"`
	CreatedAt     time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (CourseGroupStudent) TableName() string { return "course_group_student" }

type CourseGroupLesson struct {
	ID            uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	TermID        uint64    `gorm:"column:term_id;not null;uniqueIndex:uk_term_group_lesson_time;index:idx_time_room;index:idx_term_group_id;index:idx_term_group_status" json:"term_id"`
	CourseGroupID uint64    `gorm:"column:course_group_id;not null;uniqueIndex:uk_term_group_lesson_time;index:idx_term_group_id;index:idx_course_group_id;index:idx_term_group_status" json:"course_group_id"`
	WeekNo        int       `gorm:"column:week_no;not null;uniqueIndex:uk_term_group_lesson_time;index:idx_time_room" json:"week_no"`
	Weekday       int       `gorm:"column:weekday;not null;uniqueIndex:uk_term_group_lesson_time;index:idx_time_room" json:"weekday"`
	Section       int       `gorm:"column:section;not null;uniqueIndex:uk_term_group_lesson_time;index:idx_time_room" json:"section"`
	BuildingName  string    `gorm:"column:building_name;size:50;not null;index:idx_time_room" json:"building_name"`
	RoomName      string    `gorm:"column:room_name;size:50;not null;index:idx_time_room" json:"room_name"`
	Status        int       `gorm:"column:status;not null;default:1;index:idx_term_group_status;index:idx_status" json:"status"`
	CreatedAt     time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (CourseGroupLesson) TableName() string { return "course_group_lesson" }

type AttendanceRecord struct {
	ID                  uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	TermID              uint64    `gorm:"column:term_id;not null;index:idx_term_student;index:idx_term_class;index:idx_term_course;index:idx_term_student_status;index:idx_term_class_status;index:idx_term_course_status;index:idx_term_lesson_id;index:idx_term_updated_by_user_id" json:"term_id"`
	CourseID            uint64    `gorm:"column:course_id;not null;index:idx_term_course;index:idx_term_course_status" json:"course_id"`
	CourseGroupLessonID uint64    `gorm:"column:course_group_lesson_id;not null;uniqueIndex:uk_lesson_student;index:idx_term_lesson_id" json:"course_group_lesson_id"`
	StudentID           uint64    `gorm:"column:student_id;not null;uniqueIndex:uk_lesson_student;index:idx_term_student;index:idx_term_student_status" json:"student_id"`
	ClassID             *uint64   `gorm:"column:class_id;index:idx_term_class;index:idx_term_class_status" json:"class_id"`
	AttendanceStatus    int       `gorm:"column:attendance_status;not null;default:0;index:idx_term_student_status;index:idx_term_class_status;index:idx_term_course_status" json:"attendance_status"`
	UpdatedByUserID     *uint64   `gorm:"column:updated_by_user_id;index:idx_term_updated_by_user_id" json:"updated_by_user_id"`
	UpdatedAt           time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	CreatedAt           time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (AttendanceRecord) TableName() string { return "attendance_record" }

type AttendanceRecordLog struct {
	ID                  uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	TermID              uint64    `gorm:"column:term_id;not null;index:idx_term_id;index:idx_term_attendance_record_id;index:idx_term_operated_by_user_id" json:"term_id"`
	AttendanceRecordID  uint64    `gorm:"column:attendance_record_id;not null;index:idx_term_attendance_record_id" json:"attendance_record_id"`
	OperatedByUserID    uint64    `gorm:"column:operated_by_user_id;not null;index:idx_term_operated_by_user_id" json:"operated_by_user_id"`
	OldAttendanceStatus int       `gorm:"column:old_attendance_status;not null" json:"old_attendance_status"`
	NewAttendanceStatus int       `gorm:"column:new_attendance_status;not null" json:"new_attendance_status"`
	CreatedAt           time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (AttendanceRecordLog) TableName() string { return "attendance_record_log" }
