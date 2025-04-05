package store

import "errors"

var (
	ErrUserNotFound        = errors.New("user not found")
	ErrUserAlreadyExists   = errors.New("user already exists")
	ErrLinkAlreadyExists   = errors.New("link already exists")
	ErrDatabaseOperation   = errors.New("database operation failed")
	ErrLastInsertIDFailed  = errors.New("failed to get last insert ID")
	ErrUserRetrievalFailed = errors.New("failed to retrieve created user")
	ErrDataScanFailed      = errors.New("failed to scan rows")
	ErrNoRowsAffected      = errors.New("no rows were affected by the operation")
)
