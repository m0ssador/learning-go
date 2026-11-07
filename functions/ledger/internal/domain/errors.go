package domain

import "errors"

var (
	ErrBudgetExceeded = errors.New("budget exceeded")
	ErrNotFound       = errors.New("not found")
)

type ValidationError struct {
	msg string
}

func (e *ValidationError) Error() string {
	return e.msg
}

func invalid(msg string) error {
	return &ValidationError{msg: msg}
}
