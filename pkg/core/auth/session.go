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

	InitialAccess SessionAccessEvent `json:"initial_access"`
	LastAccess    SessionAccessEvent `json:"last_access"`

	CreatedAt  time.Time `json:"created_at"`
	AccessedAt time.Time `json:"accessed_at"`

	AccessToken          string    `json:"access_token"`
	RefreshToken         string    `json:"refresh_token,omitempty"`
	AccessTokenCreatedAt time.Time `json:"access_token_created_at"`
}

type SessionAccessEvent struct {
	Timestamp time.Time                   `json:"time"`
	Remote    SessionAccessEventConnInfo  `json:"remote,omitempty"`
	UserAgent string                      `json:"user_agent,omitempty"`
	Extra     SessionAccessEventExtraInfo `json:"extra,omitempty"`
}

type SessionAccessEventConnInfo struct {
	RemoteAddr    string `json:"remote_addr,omitempty"`
	XForwardedFor string `json:"x_forwarded_for,omitempty"`
	XRealIP       string `json:"x_real_ip,omitempty"`
	Forwarded     string `json:"forwarded,omitempty"`
}

type SessionAccessEventExtraInfo map[string]interface{}

func (i SessionAccessEventExtraInfo) DeviceName() string {
	deviceName, _ := i["device_name"].(string)
	return deviceName
}
