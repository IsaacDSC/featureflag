package errorutils

type ValidationError struct {
	msg string
}

func NewValidationError(msg string) *ValidationError {
	return &ValidationError{msg: msg}
}

func (e ValidationError) Error() string {
	return e.msg
}

func (e ValidationError) GetStatusCode() int {
	return 400
}
