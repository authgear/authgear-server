package mockoidc

import (
	"encoding/json"

	"github.com/golang-jwt/jwt/v5"
)

type User interface {
	ID() string

	Userinfo([]string) ([]byte, error)

	Claims([]string, *IDTokenClaims) (jwt.Claims, error)
}

type MockUser struct {
	Subject           string
	Email             string
	EmailVerified     bool
	PreferredUsername string
	Phone             string
}

func DefaultUser() *MockUser {
	return &MockUser{
		Subject:           "mock",
		Email:             "mock@example.com",
		PreferredUsername: "mock",
		Phone:             "+85295000001",
	}
}

type mockUserinfo struct {
	Email             string `json:"email,omitempty"`
	PreferredUsername string `json:"preferred_username,omitempty"`
	Phone             string `json:"phone_number,omitempty"`
}

func (u *MockUser) ID() string {
	return u.Subject
}

func (u *MockUser) Userinfo(scope []string) ([]byte, error) {
	info := &mockUserinfo{
		Email:             u.Email,
		PreferredUsername: u.PreferredUsername,
		Phone:             u.Phone,
	}

	return json.Marshal(info)
}

type mockClaims struct {
	*IDTokenClaims
	Email             string `json:"email,omitempty"`
	PreferredUsername string `json:"preferred_username,omitempty"`
	Phone             string `json:"phone_number,omitempty"`
}

func (u *MockUser) Claims(scope []string, claims *IDTokenClaims) (jwt.Claims, error) {
	return &mockClaims{
		IDTokenClaims:     claims,
		Email:             u.Email,
		PreferredUsername: u.PreferredUsername,
		Phone:             u.Phone,
	}, nil
}
