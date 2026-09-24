package domain

import "errors"

var (
	ErrInvalidCredentials  = errors.New("invalid username or password")
	ErrAccountInactive     = errors.New("account is inactive")
	ErrRefreshTokenInvalid = errors.New("invalid or expired refresh token")
	ErrWrongPassword       = errors.New("current password is incorrect")
	ErrWeakPassword        = errors.New("new password must be at least 8 characters")
	ErrSamePassword        = errors.New("new password must differ from the current one")
	// ErrProtectedAccount means the account's password comes from the
	// ADMIN_PASSWORD setting, which SeedAdmin re-applies on every startup, so
	// a change made here would silently revert.
	ErrProtectedAccount = errors.New("this account's password is set by ADMIN_PASSWORD and can't be changed here")
)
