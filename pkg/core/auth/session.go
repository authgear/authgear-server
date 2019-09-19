package auth

import (
	"time"
)

type SessionTokenKind string

const (
	SessionTokenKindAccessToken  SessionTokenKind = "access-token"
	SessionTokenKindRefreshToken SessionTokenKind = "refresh-token"
)

type PrincipalType string

const (
	PrincipalTypePassword    PrincipalType = "password"
	PrincipalTypeOAuth       PrincipalType = "oauth"
	PrincipalTypeCustomToken PrincipalType = "custom_token"
)

// Session represents a session of user logged in with a principal.
type Session struct {
	ID       string `json:"id"`
	ClientID string `json:"client_id"`

	UserID string `json:"user_id"`

	PrincipalID        string        `json:"principal_id"`
	PrincipalType      PrincipalType `json:"principal_type"`
	PrincipalUpdatedAt time.Time     `json:"principal_updated_at"`

	AuthenticatorID         string                  `json:"authenticator_id,omitempty"`
	AuthenticatorType       AuthenticatorType       `json:"authenticator_type,omitempty"`
	AuthenticatorOOBChannel AuthenticatorOOBChannel `json:"authenticator_oob_channel,omitempty"`
	AuthenticatorUpdatedAt  *time.Time              `json:"authenticator_updated_at,omitempty"`

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

type SessionCreateReason string

const (
	SessionCreateReasonSignup = "signup"
	SessionCreateReasonLogin  = "login"
)
