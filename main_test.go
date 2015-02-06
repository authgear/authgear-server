package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestAuthHandler(t *testing.T) {
	var authJson = `{
	"action": "auth:login",
	"api_key": "oursky",
	"email": "rickmak@gmail.com",
	"password": "validpassword"
}`

	req, _ := http.NewRequest(
		"POST",
		"http://ourd.dev/api/v1",
		strings.NewReader(authJson),
	)
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	LoginHandler(resp, req)

	var result ResponseJson
	if err := json.Unmarshal(resp.Body.Bytes(), &result); err != nil {
		t.Errorf("Repsonse is not a valid json: %v", err.Error())
	}
	if result.UserId != "rickmak-oursky" {
		t.Errorf("UserId mismatch, expecting `rickmak-oursky`, got %v", result.UserId)
	}
}
