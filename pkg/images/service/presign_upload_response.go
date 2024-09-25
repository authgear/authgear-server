package service

import (
	"net/http"
)

type HeaderField struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type PresignUploadResponse struct {
	Key     string        `json:"key"`
	URL     string        `json:"url"`
	Method  string        `json:"method"`
	Headers []HeaderField `json:"headers"`
}

func NewPresignUploadResponse(r *http.Request, key string) PresignUploadResponse {
	var headers []HeaderField
	for key, values := range r.Header {
		for _, value := range values {
			headers = append(headers, HeaderField{
				Name:  key,
				Value: value,
			})
		}
	}

	return PresignUploadResponse{
		Key:     key,
		URL:     r.URL.String(),
		Method:  r.Method,
		Headers: headers,
	}
}
