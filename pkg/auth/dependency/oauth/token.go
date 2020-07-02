package oauth

import (
	"github.com/authgear/authgear-server/pkg/core/crypto"
	"github.com/authgear/authgear-server/pkg/core/rand"
)

const (
	tokenAlphabet string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

func GenerateToken() string {
	token := rand.StringWithAlphabet(32, tokenAlphabet, rand.SecureRand)
	return token
}

func HashToken(token string) string {
	return crypto.SHA256String(token)
}
