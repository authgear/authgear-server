package protocol

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

type OAuthProtocolError struct {
	Response ErrorResponse
}

func NewError(err, description string) error {
	return &OAuthProtocolError{NewErrorResponse(err, description)}
}

func (e *OAuthProtocolError) Error() string { return e.Response["error_description"] }
