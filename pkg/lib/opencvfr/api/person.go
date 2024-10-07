package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/lib/opencvfr/openapi"
)

type PersonHTTPClient interface {
	Get(path string, query url.Values) (respBody []byte, err error)
	Post(path string, body io.Reader, expectedStatus int) (respBody []byte, err error)
	Patch(path string, body io.Reader) (respBody []byte, err error)
	Delete(path string, targetID string) (err error)
}

type PersonService struct {
	HTTPClient PersonHTTPClient
}

func (ps *PersonService) Create(reqBody *openapi.CreatePersonSchema) (p *openapi.PersonSchema, err error) {
	path := "/person"

	rbb, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	body, err := ps.HTTPClient.Post(path, bytes.NewBuffer(rbb), http.StatusCreated)
	if err != nil {
		return nil, fmt.Errorf("failed to create person - req: %v, err: %w", reqBody, err)
	}

	p = &openapi.PersonSchema{}
	err = p.UnmarshalJSON(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse POST %v response body: %w", path, err)
	}

	return p, nil
}
