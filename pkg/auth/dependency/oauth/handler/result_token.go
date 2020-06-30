package handler

import (
	"encoding/json"
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/dependency/oauth/protocol"
)

type TokenResult interface {
	WriteResponse(rw http.ResponseWriter, r *http.Request)
	IsInternalError() bool
}

type (
	tokenResultOK struct {
		Response protocol.TokenResponse
	}
	tokenResultError struct {
		InternalError bool
		Response      protocol.ErrorResponse
	}
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
		rw.WriteHeader(http.StatusBadRequest)
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
