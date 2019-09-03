package handler

import (
	"encoding/json"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

type RequestPayload interface {
	Validate() error
}

type EmptyRequestPayload struct{}

func (p EmptyRequestPayload) Validate() error {
	return nil
}

func DecodeJSONBody(r *http.Request, payload interface{}) error {
	if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
		return skyerr.NewError(skyerr.BadRequest, "invalid request content type")
	}
	if err := json.NewDecoder(r.Body).Decode(payload); err != nil {
		return skyerr.NewError(skyerr.BadRequest, "fails to decode the request payload")
	}
	return nil
}
