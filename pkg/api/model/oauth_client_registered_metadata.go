package model

// OAuthClientRegisteredMetadata is the OAuth client metadata a DCR
// registration returns per RFC 7591 §3.2.1 -- the same metadata
// oauth.client.registered's audit event shows for the client that was
// created. There is exactly one type and one constructor
// (NewOAuthClientRegisteredMetadata) for it, embedded by both
// handler.RegistrationResponse and
// nonblocking.OAuthClientRegisteredEventPayloadClient, so that a field
// present in the HTTP response but missing from the audit event (or vice
// versa) is structurally impossible rather than something a future change
// has to remember to keep in sync by hand.
//
// Always construct this via NewOAuthClientRegisteredMetadata -- never
// populate its fields individually at a call site. Doing so would defeat
// the whole point: the guarantee holds only because every consumer gets
// its values from the one function that knows how to read them off
// *OAuthClient.
type OAuthClientRegisteredMetadata struct {
	ClientName      string   `json:"client_name,omitempty"`
	RedirectURIs    []string `json:"redirect_uris"`
	GrantTypes      []string `json:"grant_types"`
	ResponseTypes   []string `json:"response_types"`
	ApplicationType string   `json:"application_type"`
	// TokenEndpointAuthMethod is always "none": DCR ignores whatever the
	// caller asked for and always registers a public client with no
	// secret. Never omitted -- see RegistrationResponse's own comment on
	// why a client needs to see this even though it never varies.
	TokenEndpointAuthMethod string `json:"token_endpoint_auth_method"`
	ClientURI               string `json:"client_uri,omitempty"`
	LogoURI                 string `json:"logo_uri,omitempty"`
	TOSURI                  string `json:"tos_uri,omitempty"`
	PolicyURI               string `json:"policy_uri,omitempty"`
}

// NewOAuthClientRegisteredMetadata reads the RFC 7591 response metadata off
// a freshly created or resolved *OAuthClient. See
// OAuthClientRegisteredMetadata's own comment for why every consumer must
// go through this constructor rather than building the struct by hand.
func NewOAuthClientRegisteredMetadata(c *OAuthClient) OAuthClientRegisteredMetadata {
	return OAuthClientRegisteredMetadata{
		ClientName:              c.Name,
		RedirectURIs:            c.RedirectURIs,
		GrantTypes:              c.GrantTypes,
		ResponseTypes:           c.ResponseTypes,
		ApplicationType:         derefStringOr(c.ApplicationType, ""),
		TokenEndpointAuthMethod: "none",
		ClientURI:               derefStringOr(c.ClientURI, ""),
		LogoURI:                 derefStringOr(c.LogoURI, ""),
		TOSURI:                  derefStringOr(c.TOSURI, ""),
		PolicyURI:               derefStringOr(c.PolicyURI, ""),
	}
}

func derefStringOr(s *string, fallback string) string {
	if s == nil {
		return fallback
	}
	return *s
}
