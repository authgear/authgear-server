package cloudstorage

import (
	"fmt"
	"mime"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/uuid"
)

type PresignUploadRequest struct {
	Prefix  string                 `json:"prefix,omitempty"`
	Access  AccessType             `json:"access,omitempty"`
	Headers map[string]interface{} `json:"headers"`
}

func (r *PresignUploadRequest) SetDefaultValue() {
	if r.Access == "" {
		r.Access = AccessTypeDefault
	}
	if _, ok := r.Headers["content-type"]; !ok {
		r.Headers["content-type"] = "application/octet-stream"
	}
}

func (r *PresignUploadRequest) DeriveAssetName() (assetName string, err error) {
	// Derive file extension
	contentType, ok := r.Headers["content-type"].(string)
	var ext string
	if ok {
		exts, err := mime.ExtensionsByType(contentType)
		if err != nil {
			return "", err
		}
		if len(exts) > 0 {
			ext = exts[0]
		}
	}

	assetName = fmt.Sprintf("%s%s%s", r.Prefix, uuid.New(), ext)
	return
}

func (r *PresignUploadRequest) SetCacheControl() {
	if _, ok := r.Headers["cache-control"]; !ok {
		r.Headers["cache-control"] = "max-age: 3600"
	}
}

func (r *PresignUploadRequest) RemoveEmptyHeaders() {
	// Remove any header whose value is empty string
	headers := make(map[string]interface{})
	for key, value := range r.Headers {
		if v, ok := value.(string); ok && v == "" {
			continue
		}
		headers[key] = value
	}
	r.Headers = headers
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
