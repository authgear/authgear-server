// Package skyerr contains information of errors used in skygear.
package skyerr

import (
	"encoding/json"
	"fmt"

	"github.com/oursky/skygear/skydb"
)

// Various errors emitted by Skygear handlers
var (
	ErrAuthFailure  = newError("AuthenticationError", 101, "authentication failed")
	ErrInvalidLogin = newError("AuthenticationError", 102, "invalid authentication information")

	ErrWriteDenied = newError("PermissionDenied", 101, "write is not allowed")

	ErrDatabaseOpenFailed            = newError("DatabaseError", 101, "failed to open database")
	ErrDatabaseQueryFailed           = newError("DatabaseError", 102, "failed to query record")
	ErrDatabaseSchemaMigrationFailed = newError("DatabaseError", 103, "failed to migrate record schema")
	ErrDatabaseTxNotSupported        = newError("DatabaseError", 660, "database impl does not support transaction")

	ErrUserNotFound   = newNotFoundErr(101, "user not found")
	ErrDeviceNotFound = newNotFoundErr(102, "device not found")
	ErrRecordNotFound = newNotFoundErr(103, "record not found")

	ErrUserDuplicated = newDuplicatedErr(101, "user duplicated")
)

// ErrCode is error code being assigned to router.Payload when
// Handler signifies an error
type ErrCode uint

// UnknownErr represents an unknown error
const UnknownErr ErrCode = 1

// ErrCode signifying authentication error
const (
	_ ErrCode = 100 + iota
	UserIDDuplicatedErr
	UserIDNotFoundErr
	AuthenticationInfoIncorrectErr
	InvalidAccessTokenErr
	CannotVerifyAPIKey
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

// Error specifies the interfaces required by an error in skygear
type Error interface {
	Type() string
	Code() uint
	Message() string
	Info() map[string]interface{}
	error
	json.Marshaler
}

// genericError is an intuitive implementation of Error that contains
// an code and error message.
type genericError struct {
	t       string
	code    uint
	message string
	info    map[string]interface{}
}

func newError(t string, code uint, message string) Error {
	return &genericError{
		t:       t,
		code:    code,
		message: message,
	}
}

func newNotFoundErr(code uint, message string) Error {
	return newError("ResourceNotFound", code, message)
}

func newDuplicatedErr(code uint, message string) Error {
	return newError("ResourceDuplicated", code, message)
}

// NewUnknownErr returns a new UnknownError
func NewUnknownErr(err error) Error {
	return newError("UnknownError", 1, err.Error())
}

// NewRequestInvalidErr returns a new RequestInvalid Error
func NewRequestInvalidErr(err error) Error {
	return newError("RequestInvalid", 101, err.Error())
}

// NewRequestJSONInvalidErr returns new RequestJSONInvalid Error
func NewRequestJSONInvalidErr(err error) Error {
	return newError("RequestInvalid", 102, err.Error())
}

// NewResourceFetchFailureErr returns a new ResourceFetchFailure Error
func NewResourceFetchFailureErr(kind string, id interface{}) Error {
	return newError("ResourceFetchFailure", 101, fmt.Sprintf("failed to fetch %v id = %v", kind, id))
}

func newResourceSaveFailureErr(kind string, id interface{}) Error {
	var message string
	if id != nil {
		message = fmt.Sprintf("failed to save %v id = %v", kind, id)
	} else {
		message = "failed to save " + kind
	}

	return newError("ResourceSaveFailure", 101, message)
}

// NewResourceSaveFailureErrWithStringID returns a new ResourceSaveFailure Error
// with the specified kind and string id in the error message
func NewResourceSaveFailureErrWithStringID(kind string, id string) Error {
	var iID interface{}
	if id != "" {
		iID = id
	}
	return newResourceSaveFailureErr(kind, iID)
}

func newResourceDeleteFailureErr(kind string, id interface{}) Error {
	var message string
	if id != nil {
		message = fmt.Sprintf("failed to delete %v id = %v", kind, id)
	} else {
		message = "failed to delete " + kind
	}

	return newError("ResourceDeleteFailure", 101, message)
}

// NewResourceDeleteFailureErrWithStringID returns a new ResourceDeleteFailure Error
func NewResourceDeleteFailureErrWithStringID(kind string, id string) Error {
	var iID interface{}
	if id != "" {
		iID = id
	}
	return newResourceDeleteFailureErr(kind, iID)
}

// NewAtomicOperationFailed return a new DatabaseError to be returned
// when atomic operation (like record save/delete) failed due to
// one of the sub-operation failed
func NewAtomicOperationFailedErr(errMap map[skydb.RecordID]error) Error {
	info := map[string]interface{}{}
	for recordID, err := range errMap {
		info[recordID.String()] = err.Error()
	}

	return &genericError{
		t:       "DatabaseError",
		code:    666,
		message: "Atomic Operation rolled back due to one or more errors",
		info:    info,
	}
}

// NewAtomicOperationFailedErrWithCause return a new DatabaseError to be returned
// when atomic operation (like record save/delete) failed due to
// a global operation failed
func NewAtomicOperationFailedErrWithCause(err error) Error {
	return &genericError{
		t:       "DatabaseError",
		code:    667,
		message: "Atomic Operation rolled back due to an error",
		info:    map[string]interface{}{"innerError": err},
	}

}

// New creates an Error to be returned as Response's result
func New(code ErrCode, message string) Error {
	return &genericError{
		code:    uint(code),
		message: message,
	}
}

// NewFmt creates an Error with the specified fmt and args
func NewFmt(code ErrCode, format string, args ...interface{}) Error {
	return New(code, fmt.Sprintf(format, args...))
}

func (e *genericError) Type() string {
	return e.t
}

func (e *genericError) Code() uint {
	return e.code
}

func (e *genericError) Message() string {
	return e.message
}

func (e *genericError) Info() map[string]interface{} {
	return e.info
}

func (e *genericError) Error() string {
	return fmt.Sprintf("%v: %v", e.code, e.message)
}

func (e *genericError) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type    string                 `json:"type"`
		Code    uint                   `json:"code"`
		Message string                 `json:"message"`
		Info    map[string]interface{} `json:"info,omitempty"`
	}{e.Type(), e.Code(), e.Message(), e.Info()})
}
