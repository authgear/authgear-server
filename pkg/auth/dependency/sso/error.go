package sso

import "fmt"

type ErrorCode int

const (
	// InvalidRequest - "invalid_request"
	// The request is malformed, a required parameter is missing or a parameter has an invalid value.
	InvalidRequest ErrorCode = iota
	// InvalidClient - "invalid_client"
	// Client authentication failed.
	InvalidClient
	// InvalidGrant - "invalid_grant"
	// Invalid authorization grant, grant invalid, grant expired, or grant revoked.
	InvalidGrant
	// UnauthorizedClient - "unauthorized_client"
	// Client is not authorized to use the grant.
	UnauthorizedClient
	// UnsupportedGrantType - "unsupported_grant_type"
	// Authorization grant is not supported by the Authorization Server.
	UnsupportedGrantType
	// InvalidScope - "invalid_scope"
	// The scope is malformed or invalid.
	InvalidScope
	// UnsupportedTokenType - "unsupported_token_type"
	// The Authorization Server does not support revocation of the presented token type.
	UnsupportedTokenType
)

func (e ErrorCode) String() string {
	names := [...]string{
		"invalid_request",
		"invalid_client",
		"invalid_grant",
		"unauthorized_client",
		"unsupported_grant_type",
		"invalid_scope",
	}

	if e < InvalidRequest || e > InvalidScope {
		return "undefined"
	}

	return names[e]
}

func errorCodeFromString(input string) (e ErrorCode) {
	errors := [...]ErrorCode{
		InvalidRequest,
		InvalidClient,
		InvalidGrant,
		UnauthorizedClient,
		UnsupportedGrantType,
		InvalidScope,
	}
	for _, v := range errors {
		if input == v.String() {
			e = v
			return
		}
	}

	return
}

type ErrorResp struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
	ErrorURI         string `json:"error_uri,omitempty"`
}

type Error interface {
	Code() ErrorCode
	Error() string
}

type ssoError struct {
	code    ErrorCode
	message string
}

func RespToError(errorResp ErrorResp) Error {
	return ssoError{
		code:    errorCodeFromString(errorResp.Error),
		message: errorResp.ErrorDescription,
	}
}

func (e ssoError) Code() ErrorCode {
	return e.code
}

func (e ssoError) Error() string {
	return fmt.Sprintf("%v: %v", e.code, e.message)
}
