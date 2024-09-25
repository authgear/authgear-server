package service

import (
	"net/http"
	"strconv"
)

type PresignUploadRequest struct {
	Key     string                 `json:"key"`
	Headers map[string]interface{} `json:"headers"`
}

func (r *PresignUploadRequest) Sanitize() {
	// Remove any header whose value is empty string
	headers := make(map[string]interface{})
	for key, value := range r.Headers {
		if v, ok := value.(string); ok && v == "" {
			continue
		}
		headers[key] = value
	}
	r.Headers = headers

	if _, ok := r.Headers["content-type"]; !ok {
		r.Headers["content-type"] = "application/octet-stream"
	}
	// We do not set cache-control on objects directly.
	// Instead, the GET endpoint will set cache-control header.
}

func (r *PresignUploadRequest) ContentLength() (contentLength int) {
	if s, ok := r.Headers["content-length"].(string); ok {
		i, err := strconv.Atoi(s)
		if err == nil {
			contentLength = i
		}
	}
	return
}

func (r *PresignUploadRequest) HTTPHeader() http.Header {
	httpHeader := http.Header{}
	for key, value := range r.Headers {
		if v, ok := value.(string); ok {
			httpHeader.Add(key, v)
		}
	}
	return httpHeader
}
