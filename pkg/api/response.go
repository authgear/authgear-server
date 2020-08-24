package api

import (
	"encoding/json"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

type Response struct {
	Result interface{}
	Error  error
}

func (r *Response) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Result interface{}         `json:"result,omitempty"`
		Error  *apierrors.APIError `json:"error,omitempty"`
	}{r.Result, apierrors.AsAPIError(r.Error)})
}

// HandledError represents a handled (i.e. API responded with error) unexpected
// error. When encountered this error, panic recovery middleware should log
// the error, without changing the response.
type HandledError struct {
	Error error
}
