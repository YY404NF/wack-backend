package model

import "time"

const (
	RoleAdmin   = 1
	RoleStudent = 2

	UserStatusActive  = 1
	UserStatusFrozen  = 2
	AttendanceUnset   = 0
	AttendancePresent = 1
	AttendanceLate    = 2
	AttendanceAbsent  = 3
	AttendanceOnLeave = 4
)

type User struct {
	StudentID    string    `gorm:"column:student_id;primaryKey;size:32" json:"student_id"`
	PasswordHash string    `gorm:"column:password_hash;size:255;not null" json:"-"`
	RealName     string    `gorm:"column:real_name;size:50;not null" json:"real_name"`
	Role         int       `gorm:"column:role;not null;index:idx_role_status" json:"role"`
	Status       int       `gorm:"column:status;not null;default:1;index:idx_role_status" json:"status"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (User) TableName() string { return "user" }

type Class struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ClassCode string    `gorm:"column:class_code;size:50;not null;uniqueIndex" json:"class_code"`
	ClassName string    `gorm:"column:class_name;size:100;not null" json:"class_name"`
	Grade     int       `gorm:"column:grade;not null;index:idx_grade_major" json:"grade"`
	MajorName string    `gorm:"column:major_name;size:100;not null;index:idx_grade_major" json:"major_name"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (Class) TableName() string { return "class" }

type UserClass struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	StudentID string    `gorm:"column:student_id;size:32;not null;uniqueIndex:uk_student_class" json:"student_id"`
	ClassID   uint64    `gorm:"column:class_id;not null;uniqueIndex:uk_student_class;index:idx_user_class_class_id" json:"class_id"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (UserClass) TableName() string { return "user_class" }

type StudentFreeTime struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Term      string    `gorm:"column:term;size:20;not null;uniqueIndex:uk_term_student_time" json:"term"`
	StudentID string    `gorm:"column:student_id;size:32;not null;uniqueIndex:uk_term_student_time;index:idx_student_term" json:"student_id"`
	Weekday   int       `gorm:"column:weekday;not null;uniqueIndex:uk_term_student_time" json:"weekday"`
	Section   int       `gorm:"column:section;not null;uniqueIndex:uk_term_student_time" json:"section"`
	FreeWeeks string    `gorm:"column:free_weeks;size:100;not null" json:"free_weeks"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (StudentFreeTime) TableName() string { return "student_free_time" }

type Course struct {
	ID                     uint64    `gorm:"column:id;primaryKey" json:"id"`
	Term                   string    `gorm:"column:term;size:20;not null;index:idx_term_teacher;index:idx_term_course_name" json:"term"`
	CourseName             string    `gorm:"column:course_name;size:100;not null;index:idx_term_course_name" json:"course_name"`
	TeacherName            string    `gorm:"column:teacher_name;size:50;not null;index:idx_term_teacher" json:"teacher_name"`
	AttendanceStudentCount int       `gorm:"column:attendance_student_count;not null;default:0" json:"attendance_student_count"`
	CreatedAt              time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt              time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (Course) TableName() string { return "course" }

type CourseStudent struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	CourseID  uint64    `gorm:"column:course_id;not null;uniqueIndex:uk_course_student" json:"course_id"`
	StudentID string    `gorm:"column:student_id;size:32;not null;uniqueIndex:uk_course_student;index:idx_student_id" json:"student_id"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (CourseStudent) TableName() string { return "course_student" }

type CourseClass struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	CourseID  uint64    `gorm:"column:course_id;not null;uniqueIndex:uk_course_class" json:"course_id"`
	ClassID   uint64    `gorm:"column:class_id;not null;uniqueIndex:uk_course_class;index:idx_course_class_class_id" json:"class_id"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (CourseClass) TableName() string { return "course_class" }

type CourseSession struct {
	ID           uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	CourseID     uint64    `gorm:"column:course_id;not null;uniqueIndex:uk_course_session_no;uniqueIndex:uk_course_week_time" json:"course_id"`
	SessionNo    int       `gorm:"column:session_no;not null;uniqueIndex:uk_course_session_no" json:"session_no"`
	WeekNo       int       `gorm:"column:week_no;not null;uniqueIndex:uk_course_week_time;index:idx_weekday_section_week" json:"week_no"`
	Weekday      int       `gorm:"column:weekday;not null;uniqueIndex:uk_course_week_time;index:idx_weekday_section_week" json:"weekday"`
	Section      int       `gorm:"column:section;not null;uniqueIndex:uk_course_week_time;index:idx_weekday_section_week" json:"section"`
	BuildingName string    `gorm:"column:building_name;size:50;not null" json:"building_name"`
	RoomName     string    `gorm:"column:room_name;size:50;not null" json:"room_name"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (CourseSession) TableName() string { return "course_session" }

type AttendanceCheck struct {
	ID              uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	CourseSessionID uint64    `gorm:"column:course_session_id;not null;uniqueIndex:uk_course_session" json:"course_session_id"`
	StartedBy       string    `gorm:"column:started_by;size:32;not null;index:idx_started_by_time" json:"started_by"`
	StartedAt       time.Time `gorm:"column:started_at;not null;index:idx_started_by_time" json:"started_at"`
	CreatedAt       time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (AttendanceCheck) TableName() string { return "attendance_check" }

type AttendanceDetail struct {
	ID                uint64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	AttendanceCheckID uint64     `gorm:"column:attendance_check_id;not null;uniqueIndex:uk_check_student" json:"attendance_check_id"`
	StudentID         string     `gorm:"column:student_id;size:32;not null;uniqueIndex:uk_check_student;index:idx_student_status" json:"student_id"`
	Status            int        `gorm:"column:status;not null;default:0;index:idx_student_status;index:idx_status" json:"status"`
	StatusSetBy       *string    `gorm:"column:status_set_by;size:32;index:idx_status_set_by" json:"status_set_by"`
	StatusSetAt       *time.Time `gorm:"column:status_set_at" json:"status_set_at"`
	UpdatedAt         time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	CreatedAt         time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (AttendanceDetail) TableName() string { return "attendance_detail" }

type AttendanceDetailLog struct {
	ID                 uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	AttendanceDetailID uint64    `gorm:"column:attendance_detail_id;not null" json:"attendance_detail_id"`
	AttendanceCheckID  uint64    `gorm:"column:attendance_check_id;not null;index:idx_check_operated_at" json:"attendance_check_id"`
	StudentID          string    `gorm:"column:student_id;size:32;not null;index:idx_student_operated_at" json:"student_id"`
	OperatorID         string    `gorm:"column:operator_id;size:32;not null;index:idx_operator_operated_at" json:"operator_id"`
	OldStatus          *int      `gorm:"column:old_status" json:"old_status"`
	NewStatus          int       `gorm:"column:new_status;not null" json:"new_status"`
	OperationType      string    `gorm:"column:operation_type;size:50;not null" json:"operation_type"`
	OperatedAt         time.Time `gorm:"column:operated_at;not null;index:idx_check_operated_at;index:idx_student_operated_at;index:idx_operator_operated_at" json:"operated_at"`
	CreatedAt          time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (AttendanceDetailLog) TableName() string { return "attendance_detail_log" }

type AdminOperationLog struct {
	ID          uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	OperatorID  string    `gorm:"column:operator_id;size:32;not null;index:idx_operator_time" json:"operator_id"`
	TargetTable string    `gorm:"column:target_table;size:50;not null;index:idx_target" json:"target_table"`
	TargetID    uint64    `gorm:"column:target_id;not null;index:idx_target" json:"target_id"`
	ActionType  string    `gorm:"column:action_type;size:50;not null" json:"action_type"`
	OldValue    *string   `gorm:"column:old_value;type:text" json:"old_value"`
	NewValue    *string   `gorm:"column:new_value;type:text" json:"new_value"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime;index:idx_operator_time" json:"created_at"`
}

func (AdminOperationLog) TableName() string { return "admin_operation_log" }
