package errshandle

import "strings"

func IsDuplicateKeyError(err error) bool {
	return strings.Contains(err.Error(), "UNIQUE constraint failed")
}

func IsForeignKeyError(err error) bool {
	if err == nil {
		return false
	}
	errorMsg := err.Error()
	return strings.Contains(errorMsg, "FOREIGN KEY constraint failed") ||
		strings.Contains(errorMsg, "SQLITE_CONSTRAINT_FOREIGNKEY") ||
		strings.Contains(errorMsg, "no such table:")
}
