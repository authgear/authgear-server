package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/opencvfr/openapi"
)

type LivenessHTTPClient interface {
	Post(path string, body io.Reader, expectedStatus int) (respBody []byte, err error)
}

type LivenessService struct {
	HTTPClient LivenessHTTPClient
}

func (ss *LivenessService) Check(reqBody *openapi.LivenessSchema) (r *openapi.LivenessResultSchema, err error) {
	path := "/liveness"

	rbb, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	body, err := ss.HTTPClient.Post(path, bytes.NewBuffer(rbb), http.StatusOK)
	if err != nil {
		return nil, fmt.Errorf("failed to check liveness - req: %v, err: %w", reqBody.ToLoggingFormat(), err)
	}

	r = &openapi.LivenessResultSchema{}
	err = r.UnmarshalJSON(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse POST %v response body: %w", path, err)
	}

	return r, nil
}
