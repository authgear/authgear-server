package handlers

import (
	"testing"
)

func TestHomeHandler(t *testing.T) {
	req := Payload{}

	resp := HomeHandler(req)
	var s statusResponse

	switch pt := resp.Result.(type) {
	default:
		t.Fatalf("unexpected type %T", pt)
	case statusResponse:
		s = resp.Result.(statusResponse)
	}

	if s.Status != "OK" {
		t.Fatalf("got response %v, want `OK`", s.Status)
	}
}

func TestLoginHandler(t *testing.T) {

	req := Payload{
		Data: make(map[string]interface{}),
	}
	req.Data["email"] = "rick.mak@gmail.com"

	resp := LoginHandler(req)
	var s authResponse

	switch pt := resp.Result.(type) {
	default:
		t.Fatalf("unexpected type %T", pt)
	case authResponse:
		s = resp.Result.(authResponse)
	}

	if s.UserID != "rickmak-oursky" {
		t.Fatalf("UserId mismatch, expecting `rickmak-oursky`, got %v",
			s.UserID)
	}
}
