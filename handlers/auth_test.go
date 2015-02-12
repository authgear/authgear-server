package handlers

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func TestHomeHandler(t *testing.T) {
	req := Payload{}
	resp := httptest.NewRecorder()

	HomeHandler(resp, req)
	if resp.Code != 200 {
		t.Errorf("got status code %v, want %v", resp.Code, 200)
	}

	if resp.Body.String() != "Hello Developer" {
		t.Errorf("got body %v, want `Hello Developer`", resp.Body)
	}
}

func TestLoginHandler(t *testing.T) {

	req := Payload{
		Data: make(map[string]interface{}),
	}
	req.Data["email"] = "rick.mak@gmail.com"

	resp := httptest.NewRecorder()
	LoginHandler(resp, req)

	var result authResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &result); err != nil {
		t.Errorf("Repsonse is not a valid json: %v, %v", err.Error(), resp.Body)
	}
	if result.UserID != "rickmak-oursky" {
		t.Errorf("UserId mismatch, expecting `rickmak-oursky`, got %v", result.UserID)
	}
}
