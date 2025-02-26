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
	// ErrInvalidPayload represents an invalid payload error
	ErrInvalidPayload = NewWrongInput("invalid payload")
)

// Error represents a custom error structure.
type Error struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Details []Detail `json:"details,omitempty"`
}

// Detail contains detail information of wrong input field.
type Detail struct {
	Field       string `json:"field"`
	Description string `json:"description"`
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

// NewInternal creates a new internal error with a specific message.
func NewInternal(message string) error {
	return &Error{
		Code:    http.StatusInternalServerError,
		Message: message,
	}
}

// NewWrongInput creates a new wrong input error with specific message and details.
func NewWrongInput(message string, details ...Detail) error {
	return &Error{
		Code:    http.StatusBadRequest,
		Message: message,
		Details: details,
	}
}

// NewNotFound creates a new not found error with specific message.
func NewNotFound(message string) error {
	return &Error{
		Code:    http.StatusNotFound,
		Message: message,
	}
}

// NewConflict creates a new conflict error with specific message.
func NewConflict(message string) error {
	return &Error{
		Code:    http.StatusConflict,
		Message: message,
	}
}

// AddDetail adds a new detail to the error.
func (e *Error) AddDetail(detail Detail) {
	e.Details = append(e.Details, detail)
}

// Error returns the error message with details if any.
func (e Error) Error() string {
	var res string
	res = e.Message
	for _, d := range e.Details {
		res += fmt.Sprintf("\n- Field: %s, Description: %s", d.Field, d.Description)
	}
	return res
}
