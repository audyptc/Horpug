package domain

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrAccountInactive    = errors.New("account is inactive")
)
