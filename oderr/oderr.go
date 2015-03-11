// Package oderr contains information of errors used in ourd.
package oderr

import (
	"encoding/json"
	"fmt"
)

// ErrCode is error code being assigned to router.Payload when
// Handler signifies an error
type ErrCode int

// UnknownErr represents an unknown error
const UnknownErr ErrCode = 1

// ErrCode signifying authentication error
const (
	_ ErrCode = 100 + iota
	UserIDDuplicatedErr
	UserIDNotFoundErr
	AuthenticationInfoIncorrectErr
	InvalidAccessTokenErr
)

// ErrCode signifying invalid request
const (
	_ ErrCode = 200 + iota
	RequestInvalidErr
	MissingDatabaseIDErr
)

// ErrCode signifying internal error
const (
	_ ErrCode = 300 + iota
	PersistentStorageErr
)

// Error specifies the interfaces required by an error in ourd
type Error interface {
	Code() ErrCode
	Message() string
	error
	json.Marshaler
}

// genericError is an intuitive implementation of Error that contains
// an code and error message.
type genericError struct {
	code    ErrCode
	message string
}

// New creates an Error to be returned as Response's result
func New(code ErrCode, message string) Error {
	return &genericError{code, message}
}

// NewFmt creates an Error with the specified fmt and args
func NewFmt(code ErrCode, format string, args ...interface{}) Error {
	return New(code, fmt.Sprintf(format, args...))
}

func (e *genericError) Code() ErrCode {
	return e.code
}

func (e *genericError) Message() string {
	return e.message
}

func (e *genericError) Error() string {
	return fmt.Sprintf("%v: %v", e.code, e.message)
}

func (e *genericError) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Code    ErrCode `json:"code"`
		Message string  `json:"message"`
	}{e.Code(), e.Message()})
}
