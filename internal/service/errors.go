package service

import "errors"

var (
	ErrSystemAlreadyInitialized = errors.New("system already initialized")
	ErrInvalidCredentials       = errors.New("invalid credentials")
	ErrUserFrozen               = errors.New("user is frozen")
	ErrUserNotFound             = errors.New("user not found")
	ErrClassNotFound            = errors.New("class not found")
	ErrCourseNotFound           = errors.New("course not found")
	ErrFreeTimeNotFound         = errors.New("free time not found")
	ErrCourseSessionNotFound    = errors.New("course session not found")
	ErrAttendanceCheckNotFound  = errors.New("attendance check not found")
	ErrAttendanceDetailNotFound = errors.New("attendance detail not found")
	ErrAttendanceDeadlinePassed = errors.New("attendance entry deadline passed")
	ErrOldPasswordIncorrect     = errors.New("old password is incorrect")
	ErrAdminResetOwnPassword    = errors.New("admin cannot reset own password here")
	ErrAdminFreezeSelf          = errors.New("admin cannot freeze self")
	ErrAdminRemoveOwnRole       = errors.New("admin cannot remove own admin role")
)
