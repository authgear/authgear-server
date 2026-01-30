package api

import (
	"context"
	"encoding/json"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

type Response struct {
	Result interface{}
	Error  error
}

func (r *Response) EncodeToJSON(ctx context.Context) ([]byte, error) {
	return json.Marshal(struct {
		Result interface{}         `json:"result,omitempty"`
		Error  *apierrors.APIError `json:"error,omitempty"`
	}{r.Result, apierrors.AsAPIErrorWithContext(ctx, r.Error)})
}
