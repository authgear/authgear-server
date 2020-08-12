package handler

import (
	"io"
	"mime"
	"net/http"
	"strings"

	"github.com/authgear/authgear-server/pkg/lib/api/apierrors"
)

const BodyMaxSize = 1024 * 1024 * 10

func IsJSONContentType(contentType string) bool {
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return false
	}
	if mediaType != "application/json" {
		return false
	}
	// No params is good
	if len(params) == 0 {
		return true
	}
	// Contains unknown params
	if len(params) > 1 {
		return false
	}
	// The sole param must be charset=utf-8
	charset := params["charset"]
	return strings.ToLower(charset) == "utf-8"
}

func ParseJSONBody(r *http.Request, w http.ResponseWriter, parse func(io.Reader, interface{}) error, payload interface{}) error {
	if !IsJSONContentType(r.Header.Get("Content-Type")) {
		return apierrors.NewBadRequest("request content type is invalid")
	}
	body := http.MaxBytesReader(w, r.Body, BodyMaxSize)
	defer body.Close()
	return parse(body, payload)
}
