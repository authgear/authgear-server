package handler

// genericError is the simpliest form of error that contains
// an code and error message.
//
// genericError is only intend to be used within this file. Should
// there be needs to examine error details in the future, please
// add an error interface in the router package.
type genericError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// NewError creates an error to be returned as Response's result
func NewError(code int, message string) interface{} {
	return genericError{code, message}
}
