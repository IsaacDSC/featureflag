package errorutils

import "fmt"

type NotFoundError struct {
	msg string
}

func NewNotFoundError(msg string) *NotFoundError {
	return &NotFoundError{msg: msg}
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("not found %s", e.msg)
}

func (e NotFoundError) GetStatusCode() int {
	return 404
}
