package handler

import (
	"encoding/json"
	"net/http"

	skyerr "github.com/skygeario/skygear-server/pkg/core/xskyerr"
)

type RequestPayload interface {
	Validate() error
}

type EmptyRequestPayload struct{}

func (p EmptyRequestPayload) Validate() error {
	return nil
}

const BodyMaxSize = 1024 * 1024 * 10

func DecodeJSONBody(r *http.Request, w http.ResponseWriter, payload interface{}) error {
	if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
		return skyerr.NewBadRequest("request content type is invalid")
	}
	body := http.MaxBytesReader(w, r.Body, BodyMaxSize)
	defer body.Close()
	if err := json.NewDecoder(body).Decode(payload); err != nil {
		return skyerr.NewBadRequest("failed to decode the request payload")
	}
	return nil
}
