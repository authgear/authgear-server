package mockoidc

import (
	"encoding/json"

	"github.com/lestrrat-go/jwx/v2/jwt"
)

type User interface {
	ID() string
	Userinfo(scope []string) ([]byte, error)
	AddClaims(scope []string, token jwt.Token) error
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

func (u *MockUser) ID() string {
	return u.Subject
}

func (u *MockUser) Userinfo(scope []string) ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"email":              u.Email,
		"preferred_username": u.PreferredUsername,
		"phone_number":       u.Phone,
	})
}

func (u *MockUser) AddClaims(scope []string, token jwt.Token) error {
	err := token.Set("email", u.Email)
	if err != nil {
		return err
	}

	err = token.Set("preferred_username", u.PreferredUsername)
	if err != nil {
		return err
	}

	err = token.Set("phone_number", u.Phone)
	if err != nil {
		return err
	}

	return nil
}
