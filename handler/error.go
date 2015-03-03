package handler

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

// genericError is the simpliest form of error that contains
// an code and error message.
//
// genericError is only intend to be used within this file. Should
// there be needs to examine error details in the future, please
// add an error interface in the router package.
type genericError struct {
	Code    ErrCode `json:"code"`
	Message string  `json:"message"`
}

// NewError creates an error to be returned as Response's result
func NewError(code ErrCode, message string) interface{} {
	return genericError{code, message}
}
