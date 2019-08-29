package auth

import (
	"time"
)

type SessionTokenKind string

const (
	SessionTokenKindAccessToken  SessionTokenKind = "access-token"
	SessionTokenKindRefreshToken SessionTokenKind = "refresh-token"
)

// Session represents a session of user logged in with a principal.
type Session struct {
	ID          string `json:"id"`
	ClientID    string `json:"client_id"`
	UserID      string `json:"user_id"`
	PrincipalID string `json:"principal_id"`

	CreatedAt  time.Time `json:"created_at"`
	AccessedAt time.Time `json:"accessed_at"`

	AccessToken          string    `json:"access_token"`
	RefreshToken         string    `json:"refresh_token,omitempty"`
	AccessTokenCreatedAt time.Time `json:"access_token_created_at"`
}
