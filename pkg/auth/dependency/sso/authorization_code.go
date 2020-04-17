package sso

import (
	"github.com/skygeario/skygear-server/pkg/core/crypto"
	"github.com/skygeario/skygear-server/pkg/core/rand"
)

const (
	codeAlphabet string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

// SkygearAuthorizationCode is a OAuth authorization code like value
// that can be used to exchange access token.
type SkygearAuthorizationCode struct {
	CodeHash            string `json:"code_hash"`
	Action              string `json:"action"`
	CodeChallenge       string `json:"code_challenge"`
	UserID              string `json:"user_id"`
	PrincipalID         string `json:"principal_id,omitempty"`
	SessionCreateReason string `json:"session_create_reason,omitempty"`
}

func GenerateCode() string {
	code := rand.StringWithAlphabet(32, codeAlphabet, rand.SecureRand)
	return code
}

func HashCode(code string) string {
	return crypto.SHA256String(code)
}
