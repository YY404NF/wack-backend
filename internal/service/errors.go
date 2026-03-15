package service

import "errors"

var (
	ErrSystemAlreadyInitialized = errors.New("system already initialized")
	ErrInvalidCredentials       = errors.New("invalid credentials")
	ErrUserFrozen               = errors.New("user is frozen")
	ErrUserNotFound             = errors.New("user not found")
	ErrOldPasswordIncorrect     = errors.New("old password is incorrect")
	ErrAdminResetOwnPassword    = errors.New("admin cannot reset own password here")
	ErrAdminFreezeSelf          = errors.New("admin cannot freeze self")
	ErrAdminRemoveOwnRole       = errors.New("admin cannot remove own admin role")
)
