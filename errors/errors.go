package errors

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	// ErrNotfound represents a "record not found" in DB
	ErrNotfound = NewNotFound("record not found")
	// ErrDuplicated represents when a record already exists in DB
	ErrDuplicated = NewConflict("record already exists")
)

// Error represents a custom error structure.
type Error struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Details []Detail `json:"details,omitempty"`
}

// Detail contains detail information of wrong input field.
type Detail struct {
	Field       string
	Description string
}

// New just calls to errors.New. This function is just to avoid a dependency on the standard library.
func New(text string) error {
	return errors.New(text)
}

// As just calls to errors.As. This function is just to avoid a dependency on the standard library.
func As(err error, target any) bool {
	return errors.As(err, target)
}

// Is just calls to errors.Is. This function is just to avoid a dependency on the standard library.
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// Join just calls to errors.Join. This function is just to avoid a dependency on the standard library.
func Join(errs ...error) error {
	return errors.Join(errs...)
}

func NewInternal(message string) *Error {
	return &Error{
		Code:    http.StatusInternalServerError,
		Message: message,
	}
}

func NewWrongInput(message string, details ...Detail) *Error {
	return &Error{
		Code:    http.StatusBadRequest,
		Message: message,
		Details: details,
	}
}

func NewNotFound(message string) *Error {
	return &Error{
		Code:    http.StatusNotFound,
		Message: message,
	}
}

func NewConflict(message string) *Error {
	return &Error{
		Code:    http.StatusConflict,
		Message: message,
	}
}

func (e *Error) AddDetail(detail Detail) {
	e.Details = append(e.Details, detail)
}

func (e Error) Error() string {
	var res string
	res = e.Message
	for _, d := range e.Details {
		res += fmt.Sprintf("\n- Field: %s, Description: %s", d.Field, d.Description)
	}
	return res
}
