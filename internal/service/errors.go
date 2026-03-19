package service

import "errors"

var (
	ErrSystemAlreadyInitialized   = errors.New("system already initialized")
	ErrInvalidCredentials         = errors.New("invalid credentials")
	ErrInvalidInput               = errors.New("invalid input")
	ErrUserFrozen                 = errors.New("user is frozen")
	ErrUserNotFound               = errors.New("user not found")
	ErrClassNotFound              = errors.New("class not found")
	ErrCourseNotFound             = errors.New("course not found")
	ErrCourseGroupNotFound        = errors.New("course group not found")
	ErrCourseGroupLessonNotFound  = errors.New("course group lesson not found")
	ErrStudentNotFound            = errors.New("student not found")
	ErrTermNotFound               = errors.New("term not found")
	ErrFreeTimeNotFound           = errors.New("free time not found")
	ErrAttendanceSessionNotFound  = errors.New("attendance session not found")
	ErrAttendanceRecordNotFound   = errors.New("attendance record not found")
	ErrAttendanceDeadlinePassed   = errors.New("attendance entry deadline passed")
	ErrAttendanceSessionSubmitted = errors.New("attendance session already submitted")
	ErrAttendanceRecordLocked     = errors.New("attendance record already locked")
	ErrOldPasswordIncorrect       = errors.New("old password is incorrect")
	ErrAdminResetOwnPassword      = errors.New("admin cannot reset own password here")
	ErrAdminFreezeSelf            = errors.New("admin cannot freeze self")
	ErrAdminRemoveOwnRole         = errors.New("admin cannot remove own admin role")
)
