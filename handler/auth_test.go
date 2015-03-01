package handler

import (
	"testing"

	"github.com/oursky/ourd/router"
)

func TestHomeHandler(t *testing.T) {
	req := router.Payload{}
	resp := router.Response{}

	HomeHandler(&req, &resp)
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

	req := router.Payload{
		Data: make(map[string]interface{}),
	}
	req.Data["email"] = "rick.mak@gmail.com"
	resp := router.Response{}

	LoginHandler(&req, &resp)
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
