package handler

import (
	"encoding/json"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/api/apierrors"
)

type APIResponse struct {
	Result interface{}
	Error  error
}

func (r APIResponse) MarshalJSON() ([]byte, error) {
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

func WriteResponse(rw http.ResponseWriter, response APIResponse) {
	httpStatus := http.StatusOK
	encoder := json.NewEncoder(rw)
	err := apierrors.AsAPIError(response.Error)

	if err != nil {
		httpStatus = err.Code
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(httpStatus)
	encoder.Encode(response)

	if err != nil && err.Code >= 500 && err.Code < 600 {
		// delegate logging to panic recovery
		panic(HandledError{response.Error})
	}
}
