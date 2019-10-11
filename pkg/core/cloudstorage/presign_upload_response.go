package cloudstorage

import (
	"net/http"
)

type HeaderField struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type PresignUploadResponse struct {
	AssetID string        `json:"asset_id"`
	URL     string        `json:"url"`
	Method  string        `json:"method"`
	Headers []HeaderField `json:"headers"`
}

func NewPresignUploadResponse(assetID string, r *http.Request) PresignUploadResponse {
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
		AssetID: assetID,
		URL:     r.URL.String(),
		Method:  r.Method,
		Headers: headers,
	}
}
