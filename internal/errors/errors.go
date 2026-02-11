package errors

import (
	"errors"
)

var (
	// --- Domain & Database Errors ---
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidInput      = errors.New("invalid input")
	ErrInternal          = errors.New("internal server error")
	ErrInvalidRole       = errors.New("invalid role")

	// --- Auth & JWT Errors ---
	ErrInvalidPassword         = errors.New("invalid email or password")
	ErrNoTokenProvided         = errors.New("no token provided")
	ErrTokenExpiredOrInvalid   = errors.New("invalid or expired token")
	ErrUnexpectedSigningMethod = errors.New("unexpected signing method")
	ErrInvalidTokenSubject     = errors.New("invalid token subject")
	ErrInvalidClaims           = errors.New("invalid token claims")
	ErrFailedToGenerateToken   = errors.New("failed to generate token")

	// --- RBAC Errors ---
	ErrPermissionDenied = errors.New("permission denied")
	ErrUnauthorized     = errors.New("unauthorized access")

	// --- Handler Errors ---
	ErrFailedToParseRequestBody = errors.New("failed to parse request body")
)
