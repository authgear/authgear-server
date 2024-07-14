package accountmanagement

import (
	"time"

	"github.com/authgear/authgear-server/pkg/util/crypto"
	"github.com/authgear/authgear-server/pkg/util/rand"
)

type Token struct {
	AppID     string     `json:"app_id,omitempty"`
	UserID    string     `json:"user_id,omitempty"`
	State     string     `json:"state,omitempty"`
	TokenHash string     `json:"token_hash,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	ExpireAt  *time.Time `json:"expire_at,omitempty"`
}

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
