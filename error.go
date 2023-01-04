package wool

import (
	"fmt"
	"net/http"
)

type Error struct {
	Code      int    `json:"code,omitempty"`
	Message   string `json:"message,omitempty"`
	Data      any    `json:"data,omitempty"`
	Developer string `json:"developer_message,omitempty"`
	Internal  error  `json:"-"`
}

func (e *Error) Error() string {
	if e.Internal == nil {
		return fmt.Sprintf("code=%d, message=%v, data=%v", e.Code, e.Message, e.Data)
	}
	return fmt.Sprintf("code=%d, message=%v, data=%v, internal=%v", e.Code, e.Message, e.Data, e.Internal)
}

func (e *Error) Unwrap() error {
	return e.Internal
}

func NewError(code int, err error, message ...string) *Error {
	e := &Error{Code: code, Message: http.StatusText(code), Internal: err}
	if len(message) > 0 {
		e.Message = message[0]
	}
	return e
}

func NewErrBadRequest(err error, message ...string) *Error {
	return NewError(http.StatusBadRequest, err, message...)
}

func NewErrUnauthorized(err error, message ...string) *Error {
	return NewError(http.StatusUnauthorized, err, message...)
}

func NewErrForbidden(err error, message ...string) *Error {
	return NewError(http.StatusForbidden, err, message...)
}

func NewErrNotFound(err error, message ...string) *Error {
	return NewError(http.StatusNotFound, err, message...)
}

func NewErrMethodNotAllowed(err error, message ...string) *Error {
	return NewError(http.StatusMethodNotAllowed, err, message...)
}

func NewErrConflict(err error, message ...string) *Error {
	return NewError(http.StatusConflict, err, message...)
}

func NewErrRequestEntityTooLarge(err error, message ...string) *Error {
	return NewError(http.StatusRequestEntityTooLarge, err, message...)
}

func NewErrUnprocessableEntity(err error, data any, message ...string) *Error {
	e := NewError(http.StatusUnprocessableEntity, err, message...)
	e.Data = data

	return e
}

func NewErrInternalServerError(err error, message ...string) *Error {
	return NewError(http.StatusInternalServerError, err, message...)
}
