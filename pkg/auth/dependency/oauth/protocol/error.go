package protocol

import (
	"fmt"
	"sort"
	"strings"
)

type ErrorResponse map[string]string

func NewErrorResponse(err, description string) ErrorResponse {
	resp := ErrorResponse{}
	resp.Error(err)
	resp.ErrorDescription(description)
	return resp
}

func (r ErrorResponse) Error(v string)            { r["error"] = v }
func (r ErrorResponse) ErrorDescription(v string) { r["error_description"] = v }
func (r ErrorResponse) State(v string)            { r["state"] = v }

// ToWWWAuthenticateHeader transform OAuth error response into a value for
// HTTP WWW-Authenticate header.
// Note that the caller should ensure the response keys & values do not
// require escaping.
func (r ErrorResponse) ToWWWAuthenticateHeader() string {
	keys := make([]string, 0, len(r))
	for k := range r {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	fields := make([]string, len(keys))
	for i, key := range keys {
		fields[i] = fmt.Sprintf(`%s="%s"`, key, r[key])
	}
	return "Bearer " + strings.Join(fields, ", ")
}

type OAuthProtocolError struct {
	Response ErrorResponse
}

func NewError(err, description string) error {
	return &OAuthProtocolError{NewErrorResponse(err, description)}
}

func (e *OAuthProtocolError) Error() string { return e.Response["error_description"] }
