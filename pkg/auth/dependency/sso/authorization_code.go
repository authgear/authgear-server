package sso

// SkygearAuthorizationCode is a OAuth authorization code like value
// that can be used to exchange access token.
type SkygearAuthorizationCode struct {
	Action              string `json:"action"`
	CodeChallenge       string `json:"code_challenge"`
	UserID              string `json:"user_id"`
	PrincipalID         string `json:"principal_id,omitempty"`
	SessionCreateReason string `json:"session_create_reason,omitempty"`
}
