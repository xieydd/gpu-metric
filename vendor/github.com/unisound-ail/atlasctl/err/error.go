package err

import (
	"fmt"
)

// ErrorType represents the type of error.
type ErrorType uint

const (
	// ErrUnknown indicates a generic error.
	ErrUnknown ErrorType = iota

	// ErrInvalidVolume indicates an invalid tag or invalid use of an existing tag
	ErrInvalidVolume

	// ErrInvalidEnv indicates an invalid environment setting
	ErrInvalidEnv
)

// String is for error interface
func (e ErrorType) String() string {
	switch e {
	case ErrUnknown:
		return "unknown"
	case ErrInvalidVolume:
		return "invalid volume"
	case ErrInvalidEnv:
		return "invalid environment"

	}

	return "unrecognized error type"
}

// Error represents a parser error. The error returned from Parse is of this
// type. The error contains both a Type and Message.
type Error struct {
	// The type of error
	Type ErrorType

	// The error message
	Message string
}

// Error returns the error's message
func (e *Error) Error() string {
	return e.Message
}

// NewError is candy func
func NewError(tp ErrorType, message string) *Error {
	return &Error{
		Type:    tp,
		Message: message,
	}
}

// NewErrorf is a format candy
func NewErrorf(tp ErrorType, format string, args ...interface{}) *Error {
	return NewError(tp, fmt.Sprintf(format, args...))
}

// WrapError is a wapper func
func WrapError(err error) *Error {
	ret, ok := err.(*Error)

	if !ok {
		return NewError(ErrUnknown, err.Error())
	}

	return ret
}
