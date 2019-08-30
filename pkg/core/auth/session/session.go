package session

import (
	"time"
)

type TokenKind string

const (
	TokenKindAccessToken TokenKind = "access-token"
)

// Session represents a session of user logged in with a principal.
type Session struct {
	ID          string `json:"id"`
	UserID      string `json:"user_id"`
	PrincipalID string `json:"principal_id"`

	CreatedAt  time.Time `json:"created_at"`
	AccessedAt time.Time `json:"accessed_at"`

	AccessToken          string    `json:"access_token"`
	AccessTokenCreatedAt time.Time `json:"access_token_created_at"`
}
