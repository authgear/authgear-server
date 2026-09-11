package model

// OAuthClientResolutionDocument is the CIMD client metadata document as
// fetched: no defaults applied where the document omitted a field, and
// nothing dropped -- GrantTypes in particular is the document's declared
// list before Rule 5 filters out entries Authgear does not implement,
// unlike the persisted client's own GrantTypes, which is what survived.
// This is the CIMD counterpart of OAuthClientRegistrationRequest, and for
// the same reason: a resolution that succeeds can still have silently
// dropped or ignored parts of the document (an unimplemented grant_type,
// the ignored token_endpoint_auth_method) that the persisted client alone
// cannot show. Lives in this package rather than nonblocking because it is
// plain data referenced from outside the event system too
// (pkg/lib/cimd/service.go builds one directly from the parsed document).
//
// The document is the client's own published, publicly-fetched metadata --
// never a secret (client_secret is dropped before it ever reaches Go code;
// see cimd.rawDocument's own doc comment) -- so recording it in full
// carries none of the SSRF/probing-oracle sensitivity that governs
// oauth.client.resolution.failed's uniform "unavailable" reason.
type OAuthClientResolutionDocument struct {
	ClientName      string   `json:"client_name,omitempty"`
	RedirectURIs    []string `json:"redirect_uris,omitempty"`
	GrantTypes      []string `json:"grant_types,omitempty"`
	ResponseTypes   []string `json:"response_types,omitempty"`
	ApplicationType string   `json:"application_type,omitempty"`
	// TokenEndpointAuthMethod is recorded even though CIMD ignores it, and
	// so can never cause a failure: being ignored is what makes it
	// invisible everywhere else -- the persisted client always resolves as
	// "none" regardless of what the document declared.
	TokenEndpointAuthMethod string `json:"token_endpoint_auth_method,omitempty"`
	ClientURI               string `json:"client_uri,omitempty"`
	LogoURI                 string `json:"logo_uri,omitempty"`
	TOSURI                  string `json:"tos_uri,omitempty"`
	PolicyURI               string `json:"policy_uri,omitempty"`
}
