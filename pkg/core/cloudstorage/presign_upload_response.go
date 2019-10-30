package cloudstorage

import (
	"net/http"
)

type HeaderField struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type PresignUploadResponse struct {
	AssetName string        `json:"asset_name"`
	URL       string        `json:"url"`
	Method    string        `json:"method"`
	Headers   []HeaderField `json:"headers"`
}

func NewPresignUploadResponse(r *http.Request, assetName string) PresignUploadResponse {
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
		AssetName: assetName,
		URL:       r.URL.String(),
		Method:    r.Method,
		Headers:   headers,
	}
}
