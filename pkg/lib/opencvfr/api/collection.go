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

type CollectionHTTPClient interface {
	Get(path string, query url.Values) (respBody []byte, err error)
	Post(path string, body io.Reader, expectedStatus int) (respBody []byte, err error)
	Patch(path string, body io.Reader) (respBody []byte, err error)
	Delete(path string, targetID string) (err error)
}

type CollectionService struct {
	HTTPClient CollectionHTTPClient
}

func (cs *CollectionService) Create(reqBody *openapi.CreateCollectionSchema) (c *openapi.CollectionSchema, err error) {
	path := "/collection"

	rbb, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	body, err := cs.HTTPClient.Post(path, bytes.NewBuffer(rbb), http.StatusCreated)
	if err != nil {
		return nil, fmt.Errorf("failed to create collection - req: %v, err: %w", reqBody, err)
	}

	c = &openapi.CollectionSchema{}
	err = c.UnmarshalJSON(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse POST %v response body: %w", path, err)
	}

	return c, nil
}

func (cs *CollectionService) Get(id string) (c *openapi.CollectionSchema, err error) {
	path := "/collection/" + id

	body, err := cs.HTTPClient.Get(path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection - id=%s, err: %w", id, err)
	}

	c = &openapi.CollectionSchema{}
	err = c.UnmarshalJSON(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse GET %v response body: %w", path, err)
	}

	return c, nil
}

func (cs *CollectionService) Delete(id string) (err error) {
	path := "/collection"

	err = cs.HTTPClient.Delete(path, id)
	if err != nil {
		return fmt.Errorf("failed to delete collection - id=%s, err: %w", id, err)
	}

	return nil
}

func (cs *CollectionService) Update(reqBody *openapi.UpdateCollectionSchema) (c *openapi.CollectionSchema, err error) {
	path := "/collection"

	rbb, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	body, err := cs.HTTPClient.Patch(path, bytes.NewBuffer(rbb))
	if err != nil {
		return nil, fmt.Errorf("failed to update collection - req: %v, err: %w", reqBody, err)
	}

	c = &openapi.CollectionSchema{}
	err = c.UnmarshalJSON(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PATCH %v response body: %w", path, err)
	}

	return c, nil
}
