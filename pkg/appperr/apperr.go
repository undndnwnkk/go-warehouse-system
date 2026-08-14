package apperr

import (
	"errors"
	"net/http"
)

type Error struct {
	Code    string
	Message string
	Status  int
	Err     error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

func (e *Error) Unwrap() error {
	return e.Err
}

func New(status int, code, message string) *Error {
	return &Error{
		Status:  status,
		Code:    code,
		Message: message,
	}
}

func Wrap(err error, status int, code, message string) *Error {
	return &Error{
		Status:  status,
		Code:    code,
		Message: message,
		Err:     err,
	}
}

func From(err error) *Error {
	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr
	}

	return &Error{
		Status:  http.StatusInternalServerError,
		Code:    "internal_server_error",
		Message: "internal server error",
		Err:     err,
	}
}
