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
