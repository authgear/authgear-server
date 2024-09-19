package accountmanagement

import (
	"crypto/subtle"
	"time"

	"github.com/authgear/authgear-server/pkg/util/crypto"
	"github.com/authgear/authgear-server/pkg/util/rand"
)

type Token struct {
	AppID     string     `json:"app_id,omitempty"`
	UserID    string     `json:"user_id,omitempty"`
	TokenHash string     `json:"token_hash,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	ExpireAt  *time.Time `json:"expire_at,omitempty"`

	// Adding OAuth
	Alias       string `json:"alias,omitempty"`
	RedirectURI string `json:"redirect_uri,omitempty"`
	State       string `json:"state,omitempty"`

	// Adding Identity
	Identity *IdentityToken `json:"identity_token,omitempty"`
}

type IdentityToken struct {
	IdentityID  string `json:"identity_id,omitempty"`
	PhoneNumber string `json:"phone_number,omitempty"`
	Email       string `json:"email,omitempty"`
}

func (t *Token) CheckUser(userID string) error {
	if subtle.ConstantTimeCompare([]byte(t.UserID), []byte(userID)) == 1 {
		return nil
	}

	return ErrAccountManagementTokenNotBoundToUser
}

func (t *Token) CheckUser_OAuth(userID string) error {
	if subtle.ConstantTimeCompare([]byte(t.UserID), []byte(userID)) == 1 {
		return nil
	}
	return ErrOAuthTokenNotBoundToUser
}

func (t *Token) CheckState(state string) error {
	if t.State == "" {
		// token is not originally bound to state.
		return nil
	}

	if subtle.ConstantTimeCompare([]byte(t.State), []byte(state)) == 1 {
		return nil
	}

	return ErrOAuthStateNotBoundToToken
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
