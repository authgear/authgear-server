package handler

import (
	"encoding/json"
	"io"
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

const BodyMaxSize = 1024 * 1024 * 10

func DecodeJSONBody(r *http.Request, w http.ResponseWriter, payload interface{}) error {
	return ParseJSONBody(r, w, func(r io.Reader, p interface{}) error {
		if err := json.NewDecoder(r).Decode(payload); err != nil {
			return skyerr.NewBadRequest("failed to decode the request payload")
		}
		return nil
	}, payload)
}

func ParseJSONBody(r *http.Request, w http.ResponseWriter, parse func(io.Reader, interface{}) error, payload interface{}) error {
	if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
		return skyerr.NewBadRequest("request content type is invalid")
	}
	body := http.MaxBytesReader(w, r.Body, BodyMaxSize)
	defer body.Close()
	return parse(body, payload)
}
