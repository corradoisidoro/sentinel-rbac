package errors

import "errors"

var (
	ErrUserAlreadyExists     = errors.New("user already exists")
	ErrUserNotFound          = errors.New("user not found")
	ErrInvalidInput          = errors.New("invalid input")
	ErrInternal              = errors.New("internal error")
	ErrInvalidPassword       = errors.New("invalid password")
	ErrFailedToGenerateToken = errors.New("failed to generate token")
)
