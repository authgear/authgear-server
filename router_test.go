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
		t.Fatalf("Simple map failed %v", resp.Body.String())
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
		t.Fatalf("Empty route should not mappped")
	}
}

func TestRouterCheckAuth(t *testing.T) {
	mockResp := handlers.Response{}
	type exampleResp struct {
		Auth string `json:"auth"`
	}
	mockResp.Result = exampleResp{
		Auth: "ok",
	}
	mockHander := MockHander{
		outputs: mockResp,
	}
	r := NewRouter()
	r.Map("mock:map", mockHander.handle)
	r.Preprocess(CheckAuth)

	// Positive test
	var mockJSON = `{
	"action": "mock:map",
	"access_token": "validToken"
}`

	req, _ := http.NewRequest(
		"POST",
		"http://ourd.dev/api/v1",
		strings.NewReader(mockJSON),
	)
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Body.String() != "{\"result\":{\"auth\":\"ok\"}}" {
		t.Fatalf("CheckAuth failed: %v", resp.Body.String())
	}

	// Negative test
	var mockJSON2 = `{
	"action": "mock:map",
	"access_token": "InValidToken"
}`

	req2, _ := http.NewRequest(
		"POST",
		"http://ourd.dev/api/v1",
		strings.NewReader(mockJSON2),
	)
	req2.Header.Set("Content-Type", "application/json")
	resp2 := httptest.NewRecorder()
	r.ServeHTTP(resp2, req2)

	if resp2.Body.String() != "Unauthorized request" {
		t.Fatalf(
			"CheckAuth failed to reject unauthorized request: %v",
			resp2.Body.String())
	}
}
