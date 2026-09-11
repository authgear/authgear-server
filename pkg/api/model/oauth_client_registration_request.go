package model

// OAuthClientRegistrationRequest is the client metadata a POST
// /oauth2/register caller asked for, recorded as sent: no defaults, no
// normalization, nothing dropped. Shared between nonblocking's
// oauth.client.registered and oauth.client.registration.failed payloads
// rather than given one type per event, because it is the same record
// either way -- "what did the caller actually send" -- and a registration
// that succeeds can still have silently dropped or ignored parts of it (an
// unimplemented grant_type, the ignored token_endpoint_auth_method) that
// the persisted client alone cannot show. Lives in this package rather
// than nonblocking because it is plain data referenced from outside the
// event system too (pkg/lib/oauth/handler builds one directly from the
// decoded request body).
//
// Every field is client-authored and unvalidated: read it as what the
// caller claimed, never as configuration Authgear accepted. Bounded by the
// 1 MB root body limit (middleware.MaxBodySize) and the endpoint's rate
// limits.
type OAuthClientRegistrationRequest struct {
	ClientName      string   `json:"client_name,omitempty"`
	RedirectURIs    []string `json:"redirect_uris,omitempty"`
	GrantTypes      []string `json:"grant_types,omitempty"`
	ResponseTypes   []string `json:"response_types,omitempty"`
	ApplicationType string   `json:"application_type,omitempty"`
	// TokenEndpointAuthMethod is recorded even though DCR ignores it, and
	// so can never cause a failure: being ignored is what makes it
	// invisible everywhere else -- the persisted client always registers
	// as "none" regardless of what was asked for.
	TokenEndpointAuthMethod string `json:"token_endpoint_auth_method,omitempty"`
	ClientURI               string `json:"client_uri,omitempty"`
	LogoURI                 string `json:"logo_uri,omitempty"`
	TOSURI                  string `json:"tos_uri,omitempty"`
	PolicyURI               string `json:"policy_uri,omitempty"`
}
