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

func (ps *PersonService) Get(id string) (p *openapi.PersonSchema, err error) {
	path := "/person/" + id

	body, err := ps.HTTPClient.Get(path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get person - id=%s, err: %w", id, err)
	}

	p = &openapi.PersonSchema{}
	err = p.UnmarshalJSON(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse GET %v response body: %w", path, err)
	}

	return p, nil
}

func (ps *PersonService) Delete(id string) (err error) {
	path := "/person"

	err = ps.HTTPClient.Delete(path, id)
	if err != nil {
		return fmt.Errorf("failed to delete person (id=%s), err: %w", id, err)
	}

	return nil
}

func (ps *PersonService) Update(reqBody *openapi.UpdatePersonSchema) (p *openapi.PersonSchema, err error) {
	path := "/person"

	rbb, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	body, err := ps.HTTPClient.Patch(path, bytes.NewBuffer(rbb))
	if err != nil {
		return nil, fmt.Errorf("failed to update person - req: %v, err: %w", reqBody, err)
	}

	p = &openapi.PersonSchema{}
	err = p.UnmarshalJSON(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PATCH %v response body: %w", path, err)
	}

	return p, nil
}

// List searches persons by id or name, then return a slice of persons.
//
// Behavior was not documented, but experimenting shows
// 1. skip/take does not affect count
// 2. it is case-insensitive substring match by ID/Name
//
// For example, for "Alice S"
//
// matches - "Alice S", "Alice", "S", "A", "Al", "li", "ce", " ", " S", "alice", "alice s"
//
// non-matches - "e S", "AliceS"
func (ps *PersonService) List(params *openapi.ListPersonsQuery) (p *openapi.ListPersonsSchema, err error) {
	path := "/persons"

	if params == nil {
		params = &openapi.ListPersonsQuery{}
	}

	query := params.ToQuery()

	body, err := ps.HTTPClient.Get(path, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list person - params: %v, err: %w", params, err)
	}

	p = &openapi.ListPersonsSchema{}
	err = p.UnmarshalJSON(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse GET %v response body: %w", path, err)
	}

	return p, nil
}
