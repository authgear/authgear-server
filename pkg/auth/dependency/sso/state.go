package sso

// State is an opaque value used by the client to maintain
// state between the request and callback.
// See https://tools.ietf.org/html/rfc6749#section-4.1.1
type State struct {
	UserID      string            `json:"user_id,omitempty"`
	Extra       map[string]string `json:"extra,omitempty"`
	Action      string            `json:"action,omitempty"`
	HashedNonce string            `json:"hashed_nonce"`
	APIClientID string            `json:"api_client_id"`
}
