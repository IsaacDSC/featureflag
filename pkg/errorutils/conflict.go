package errorutils

import "fmt"

type ConflictError struct {
	msg string
}

func NewConflictError(msg string) *ConflictError {
	return &ConflictError{msg: msg}
}

func (e ConflictError) Error() string {
	return fmt.Sprintf("conflict %s", e.msg)
}

func (e ConflictError) GetStatusCode() int {
	return 409
}
