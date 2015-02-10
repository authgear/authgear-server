package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/oursky/ourd/handlers"
)

type MockHander struct {
	outputs string
}

func (m *MockHander) handle(r handlers.Responser, p handlers.Payloader) {
	r.Write([]byte(m.outputs))
}

func TestRouterMap(t *testing.T) {
	mockHander := MockHander{
		outputs: "outputs",
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

	if resp.Body.String() != "outputs" {
		t.Errorf("Simple map failed")
	}
}

func TestRouterMapMissing(t *testing.T) {
	mockHander := MockHander{
		outputs: "outputs",
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
