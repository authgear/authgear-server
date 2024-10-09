package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/opencvfr/openapi"
)

type SearchHTTPClient interface {
	Post(path string, body io.Reader, expectedStatus int) (respBody []byte, err error)
}

type SearchService struct {
	HTTPClient SearchHTTPClient
}

func (ss *SearchService) Verify(reqBody *openapi.VerifyPersonSchema) (r *openapi.NullableVerifyPersonResultSchema, err error) {
	path := "/verify"

	rbb, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	body, err := ss.HTTPClient.Post(path, bytes.NewBuffer(rbb), http.StatusOK)
	if err != nil {
		return nil, fmt.Errorf("failed to verify person - req: %v, err: %w", reqBody, err)
	}

	vr := &openapi.VerifyPersonResultSchema{}
	err = vr.UnmarshalJSON(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse POST %v response body: %w", path, err)
	}

	r = openapi.NewNullableVerifyPersonResultSchema(vr)
	err = r.UnmarshalJSON(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse POST %v response body: %w", path, err)
	}

	return r, nil
}

func (ss *SearchService) Search(reqBody *openapi.SearchPersonSchema) (r []*openapi.SearchPersonResultSchema, err error) {
	path := "/search"

	rbb, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	body, err := ss.HTTPClient.Post(path, bytes.NewBuffer(rbb), http.StatusOK)
	if err != nil {
		return nil, fmt.Errorf("failed to search person - req: %v, err: %w", reqBody, err)
	}

	sr := []*openapi.SearchPersonResultSchema{}
	err = json.Unmarshal(body, &sr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse POST %v response body: %w", path, err)
	}

	return sr, nil
}

func (ss *SearchService) SearchLiveFace(reqBody *openapi.SearchLiveFaceScheme) (r *openapi.NullableSearchLivePersonResultSchema, err error) {
	path := "/search-live-face"

	rbb, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	body, err := ss.HTTPClient.Post(path, bytes.NewBuffer(rbb), http.StatusOK)
	if err != nil {
		return nil, fmt.Errorf("failed to search live face person - req: %v, err: %w", reqBody, err)
	}

	vr := &openapi.SearchLivePersonResultSchema{}
	err = vr.UnmarshalJSON(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse POST %v response body: %w", path, err)
	}

	r = openapi.NewNullableSearchLivePersonResultSchema(vr)
	err = r.UnmarshalJSON(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse POST %v response body: %w", path, err)
	}

	return r, nil
}
