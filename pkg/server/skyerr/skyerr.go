// Copyright 2015-present Oursky Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package skyerr contains information of errors used in skygear.
package skyerr

import (
	"encoding/json"
	"fmt"
)

// ErrorCode is an integer representation of an error condition
// occurred within the system.
//go:generate stringer -type=ErrorCode
type ErrorCode int

// A list of all expected errors.
//
// Naming convention:
// * Try not to end an error name with "Error"
// * "NotAccepted" refers to information that seems valid but still not accepted for some reason
// * "Bad" refers to information that is malformed or in a corrupted format
// * "Invalid" refers to information that is not correct
const (
	// NotAuthenticated is for operations that requires authentication
	// but the request is not properly authenticated.
	NotAuthenticated ErrorCode = 101 + iota

	// PermissionDenied occurs when the requested resource or operation
	// exists, but the request is not allowed for some reason.
	PermissionDenied

	// AccessKeyNotAccepted occurs when the request contains access key
	// (API key), but the access key is not accepted.
	AccessKeyNotAccepted

	// AccessTokenNotAccepted occurs when the request contains access token
	// but the access token is not accepted.
	AccessTokenNotAccepted

	// InvalidCredentials occurs when the information supplied by a user
	// to get authenticated is incorrect.
	InvalidCredentials

	// InvalidSignature is returned by an operation that requires a signature
	// and the provided signature is not valid.
	InvalidSignature

	// BadRequest is an error when the server does not understand the request.
	//
	// The same error is used for requests that does not conform to HTTP
	// protocol.
	// The same error may be used for requests that are missing arguments.
	BadRequest

	// InvalidArgument is an error when the server understand the request,
	// but the supplied argument is not valid
	InvalidArgument

	// Duplicated is an error that occurs when a resource to be saved is
	// a duplicate of an existing resource
	Duplicated

	// ResourceNotFound is returned because the requested resource
	// is not found, and this is unlikely due to a failure.
	//
	// The same error is used for operations that require a critical resource
	// to be available, and that resource is specified in the request.
	ResourceNotFound

	// NotSupported occurs when the server understands the request,
	// but the feature is not available due to a known limitation.
	//
	// Use this when the feature is not likely to be implemented in the near
	// future.
	NotSupported

	// NotImplemented occurs when the server understands the request,
	// but the feature is not implemented yet.
	//
	// Use this when the feature is likely to be implemented in the near
	// future.
	NotImplemented

	// ConstraintViolated occurs when a resource cannot be saved because
	// doing so would violate a constraint.
	ConstraintViolated

	// IncompatibleSchema occurs if because the saving record is incompatible
	// with the existing schema.
	IncompatibleSchema

	// AtomicOperationFailure occurs when a batch operation failed because
	// it failed partially, and the batch operation is required to be atomic
	AtomicOperationFailure

	// PartialOperationFailure occurs when a batch operation failed because
	// it failed partially, and the batch operation is not required to be atomic
	PartialOperationFailure

	// UndefinedOperation is an operation that is not known to the system
	UndefinedOperation

	// PluginUnavailable occurs when the configured plugin is not available at
	// the moment
	PluginUnavailable

	// PluginTimeout occurs when an operation carried by a plugin is timed out
	PluginTimeout

	// RecordQueryInvalid is returned when information contained in a record
	// query
	// is not valid. Examples include referencing keypath that is invalid, and
	// unsupported comparison.
	RecordQueryInvalid

	// PluginInitializing occurs when any of the plugins are initializing
	PluginInitializing

	// ResponseTimeout occurs when an operation is taking too long to produce
	// a response
	ResponseTimeout

	// DeniedArgument occurs when the a request involves an argument
	// that the user is not allowed to specify. This might occur
	// when modifying a field in a record that the user has no write access.
	DeniedArgument

	// RecordQueryDenied is returned when the user is not allowed to
	// perform the query.
	// Examples include referencing a field that is disallowed by Field ACL.
	RecordQueryDenied

	// NotConfigured is returned when a feature requires extra configuration
	// and the required configuration is missing.
	NotConfigured

	// Error codes for expected error condition should be placed
	// above this line.
)

// A list of unexpected errors.
const (

	// UnexpectedError is for an error that is not likely to happen or
	// an error that cannot be classified into any other error type.
	//
	// Refrain from using this error code.
	UnexpectedError ErrorCode = 10000 + iota
	UnexpectedAuthInfoNotFound
	UnexpectedUnableToOpenDatabase
	UnexpectedPushNotificationNotConfigured
	InternalQueryInvalid
	UnexpectedUserNotFound

	// Error codes for unexpected error condition should be placed
	// above this line.
)

// Error specifies the interfaces required by an error in skygear
type Error interface {
	Name() string
	Code() ErrorCode
	Message() string
	Info() map[string]interface{}
	error
	json.Marshaler
}

// genericError is an intuitive implementation of Error that contains
// an code and error message.
type genericError struct {
	code    ErrorCode
	message string
	info    map[string]interface{}
}

// NewError returns an error suitable to be returned to the client
func NewError(code ErrorCode, message string) Error {
	return NewErrorWithInfo(code, message, nil)
}

// NewErrorf returns an Error
func NewErrorf(code ErrorCode, message string, a ...interface{}) Error {
	return NewError(code, fmt.Sprintf(message, a...))
}

// NewErrorWithInfo returns an Error
func NewErrorWithInfo(code ErrorCode, message string, info map[string]interface{}) Error {
	return &genericError{
		code:    code,
		message: message,
		info:    info,
	}
}

// NewInvalidArgument is a convenient function to returns an invalid argument
// error with a list of arguments that are invalid.
func NewInvalidArgument(message string, arguments []string) Error {
	return &genericError{
		code:    InvalidArgument,
		message: message,
		info: map[string]interface{}{
			"arguments": arguments,
		},
	}
}

// NewDeniedArgument is a convenient function to returns a denied argument
// error with a list of arguments that are denied.
func NewDeniedArgument(message string, arguments []string) Error {
	return &genericError{
		code:    DeniedArgument,
		message: message,
		info: map[string]interface{}{
			"arguments": arguments,
		},
	}
}

func newNotFoundErr(code ErrorCode, message string) Error {
	return NewError(code, message)
}

// MakeError returns an Error interface with the specified error. If the
// specified error already implements the Error interface, the specified error
// is returned.
//
// For specified error of other kinds, the returned error always have code
// `UnexpectedError`.
func MakeError(err error) Error {
	if skyError, ok := err.(Error); ok {
		return skyError
	}
	return NewError(UnexpectedError, err.Error())
}

// NewRequestJSONInvalidErr returns new RequestJSONInvalid Error
func NewRequestJSONInvalidErr(err error) Error {
	return NewError(BadRequest, err.Error())
}

// NewResourceFetchFailureErr returns a new ResourceFetchFailure Error
func NewResourceFetchFailureErr(kind string, id interface{}) Error {
	return NewError(UnexpectedError, fmt.Sprintf("failed to fetch %v id = %v", kind, id))
}

func NewResourceSaveFailureErr(kind string, id interface{}) Error {
	var message string
	if id != nil {
		message = fmt.Sprintf("failed to save %v id = %v", kind, id)
	} else {
		message = "failed to save " + kind
	}

	return NewError(UnexpectedError, message)
}

// NewResourceSaveFailureErrWithStringID returns a new ResourceSaveFailure Error
// with the specified kind and string id in the error message
func NewResourceSaveFailureErrWithStringID(kind string, id string) Error {
	var iID interface{}
	if id != "" {
		iID = id
	}
	return NewResourceSaveFailureErr(kind, iID)
}

func newResourceDeleteFailureErr(kind string, id interface{}) Error {
	var message string
	if id != nil {
		message = fmt.Sprintf("failed to delete %v id = %v", kind, id)
	} else {
		message = "failed to delete " + kind
	}

	return NewError(UnexpectedError, message)
}

// NewResourceDeleteFailureErrWithStringID returns a new ResourceDeleteFailure Error
func NewResourceDeleteFailureErrWithStringID(kind string, id string) Error {
	var iID interface{}
	if id != "" {
		iID = id
	}
	return newResourceDeleteFailureErr(kind, iID)
}

func (e *genericError) Name() string {
	return fmt.Sprintf("%v", e.code)
}

func (e *genericError) Code() ErrorCode {
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
		Name    string                 `json:"name"`
		Code    ErrorCode              `json:"code"`
		Message string                 `json:"message"`
		Info    map[string]interface{} `json:"info,omitempty"`
	}{e.Name(), e.Code(), e.Message(), e.Info()})
}
