package handler

import (
	"encoding/json"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
)

type (
	tokenResultOK struct {
		Response protocol.TokenResponse
	}
	tokenResultError struct {
		StatusCode    int
		InternalError bool
		Response      protocol.ErrorResponse
	}
	tokenResultEmpty struct{}
)

func (t tokenResultOK) WriteResponse(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	rw.Header().Set("Cache-Control", "no-store")
	rw.Header().Set("Pragma", "no-cache")
	rw.WriteHeader(http.StatusOK)

	encoder := json.NewEncoder(rw)
	err := encoder.Encode(t.Response)
	if err != nil {
		http.Error(rw, err.Error(), 500)
	}
}

func (t tokenResultOK) IsInternalError() bool {
	return false
}

func (t tokenResultError) WriteResponse(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	rw.Header().Set("Cache-Control", "no-store")
	rw.Header().Set("Pragma", "no-cache")
	if t.InternalError {
		rw.WriteHeader(http.StatusInternalServerError)
	} else {
		if t.StatusCode == 0 {
			rw.WriteHeader(http.StatusBadRequest)
		} else {
			rw.WriteHeader(t.StatusCode)
		}
	}

	encoder := json.NewEncoder(rw)
	err := encoder.Encode(t.Response)
	if err != nil {
		http.Error(rw, err.Error(), 500)
	}
}

func (t tokenResultError) IsInternalError() bool {
	return t.InternalError
}

func (t tokenResultEmpty) WriteResponse(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Cache-Control", "no-store")
	rw.Header().Set("Pragma", "no-cache")
	rw.WriteHeader(http.StatusOK)
}

func (t tokenResultEmpty) IsInternalError() bool {
	return false
}
