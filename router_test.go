package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/oursky/ourd/handlers"
)

type MockHander struct {
	outputs handlers.Response
}

func (m *MockHander) handle(p *handlers.Payload, r *handlers.Response) {
	r.Result = m.outputs.Result
	return
}

func TestRouterMap(t *testing.T) {
	mockResp := handlers.Response{}
	type exampleResp struct {
		Username string `json:"username"`
	}
	mockResp.Result = exampleResp{
		Username: "example",
	}
	mockHander := MockHander{
		outputs: mockResp,
	}
	r := NewRouter()
	r.Map("mock:map", mockHander.handle)
	var mockJSON = `{
	"action": "mock:map"
}`

	req, _ := http.NewRequest(
		"POST",
		"http://ourd.dev/api/v1",
		strings.NewReader(mockJSON),
	)
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Body.String() != "{\"result\":{\"username\":\"example\"}}" {
		t.Errorf("Simple map failed %v", resp.Body.String())
	}
}

func TestRouterMapMissing(t *testing.T) {
	mockHander := MockHander{
		outputs: handlers.Response{},
	}
	r := NewRouter()
	r.Map("mock:map", mockHander.handle)
	var mockJSON = `{
	"action": "missing"
}`

	req, _ := http.NewRequest(
		"POST",
		"http://ourd.dev/api/v1",
		strings.NewReader(mockJSON),
	)
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Body.String() != "Unmatched Route" {
		t.Errorf("Empty route should not mappped")
	}
}
