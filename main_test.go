package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHomeHandler(t *testing.T) {
	req, _ := http.NewRequest(
		"GET",
		"http://example.com/",
		nil,
	)
	resp := httptest.NewRecorder()

	HomeHandler(resp, req)
	if resp.Code != 200 {
		t.Errorf("got status code %v, want %v", resp.Code, 200)
	}

	if resp.Body.String() != "Hello Developer" {
		t.Errorf("got body %v, want `Hello Developer`", resp.Body)
	}
}
