package cimd

import (
	"encoding/json"
	"errors"
	"net/url"
	"slices"
	"strings"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

// CIMDDocumentInvalid is the one apierrors.Kind shared by every
// ErrDocument* sentinel below, so any of them -- individually or
// wrapped, e.g. via errors.Join -- can be recognized generically with
// apierrors.IsAPIError/IsKind instead of a bespoke plain-error check.
// Each sentinel is still distinguishable by identity (errors.Is) or by
// its apierrors.Cause (apierrors.IsAPIErrorWithCondition +
// (*APIError).HasCause).
//
// This Kind is never itself rendered into an HTTP response: Service
// (EnsureClientResolved) always collapses a validation failure into its
// own uniform CIMDUnresolvable/ErrUnresolvable() before returning past
// this package, per docs/specs/cimd.md § Authgear as an SSRF/Probing
// Oracle. Using apierrors here is about making these errors catchable
// and self-describing, not about exposing a richer error to callers.
var CIMDDocumentInvalid = apierrors.Invalid.WithReason("CIMDDocumentInvalid")

var (
	ErrDocumentNotJSONObject                      = CIMDDocumentInvalid.NewWithCause("cimd: document is not a JSON object", apierrors.StringCause("NotJSONObject"))
	ErrDocumentClientIDMismatch                   = CIMDDocumentInvalid.NewWithCause("cimd: client_id does not equal the request URL", apierrors.StringCause("ClientIDMismatch"))
	ErrDocumentRedirectURIsMissing                = CIMDDocumentInvalid.NewWithCause("cimd: redirect_uris is required", apierrors.StringCause("RedirectURIsMissing"))
	ErrDocumentRedirectURIInvalid                 = CIMDDocumentInvalid.NewWithCause("cimd: invalid redirect_uri", apierrors.StringCause("RedirectURIInvalid"))
	ErrDocumentGrantTypeUnsupported               = CIMDDocumentInvalid.NewWithCause("cimd: no supported grant_type", apierrors.StringCause("GrantTypeUnsupported"))
	ErrDocumentResponseTypeInconsistent           = CIMDDocumentInvalid.NewWithCause("cimd: response_types is inconsistent with grant_types", apierrors.StringCause("ResponseTypeInconsistent"))
	ErrDocumentApplicationTypeUnsupported         = CIMDDocumentInvalid.NewWithCause("cimd: unsupported application_type", apierrors.StringCause("ApplicationTypeUnsupported"))
	ErrDocumentTokenEndpointAuthMethodNotAccepted = CIMDDocumentInvalid.NewWithCause("cimd: token_endpoint_auth_method is not accepted", apierrors.StringCause("TokenEndpointAuthMethodNotAccepted"))
	ErrDocumentURIFieldNotHTTPS                   = CIMDDocumentInvalid.NewWithCause("cimd: uri field must use https", apierrors.StringCause("URIFieldNotHTTPS"))
)

// rawDocument is the wire shape. Every field the spec says to reject or
// ignore -- client_secret, client_secret_expires_at, jwks_uri,
// software_statement, and any unknown property -- is simply absent from
// this struct, so encoding/json drops it. There is no need to name them
// and no DisallowUnknownFields: spec § Validation says "Unrecognized
// properties are ignored (the spec explicitly allows additional
// properties)", and spec §4.1 says credential material is "always ignored",
// not "rejected". A document carrying a client_secret is therefore VALID
// and its secret is discarded.
type rawDocument struct {
	ClientID                *string  `json:"client_id"`
	ClientName              *string  `json:"client_name"`
	RedirectURIs            []string `json:"redirect_uris"`
	GrantTypes              []string `json:"grant_types"`
	ResponseTypes           []string `json:"response_types"`
	ApplicationType         *string  `json:"application_type"`
	LogoURI                 *string  `json:"logo_uri"`
	ClientURI               *string  `json:"client_uri"`
	TOSURI                  *string  `json:"tos_uri"`
	PolicyURI               *string  `json:"policy_uri"`
	TokenEndpointAuthMethod *string  `json:"token_endpoint_auth_method"`
}

// Document is a rawDocument after defaults have been applied and every
// field has passed validation. Field-for-field it is the CIMD counterpart
// of dcr.NormalizedRegistration, and for the same reason: the caller
// (Part 3) copies it straight onto oauthclient.NewClientOptions.
type Document struct {
	ClientName   *string
	RedirectURIs []string
	// GrantTypes holds only the grant types Authgear implements, in the
	// order the document listed them -- anything else the document
	// declared has been dropped by Rule 5, so this is what the client can
	// actually do here, not a verbatim copy of what it published.
	GrantTypes    []string
	ResponseTypes []string
	// ApplicationType is always "web" or "native" -- never nil, never
	// anything else.
	ApplicationType string
	LogoURI         *string
	ClientURI       *string
	TOSURI          *string
	PolicyURI       *string
}

var cimdSupportedGrantTypes = map[string]bool{
	"authorization_code": true,
	"refresh_token":      true,
}

var cimdSupportedResponseTypes = map[string]bool{
	"code": true,
}

// ParseAndValidate implements every rule in
// docs/specs/cimd.md#accepted-metadata-fields. requestURL is the exact
// client_id string that was fetched -- NOT a re-serialized url.URL, because
// the client_id equality check below is byte-for-byte.
//
// It returns the specific sentinel for whichever rule failed. Service
// collapses these into its single unresolvable error, so spec § Error
// Handling's "a document that fails any MUST-level check is treated as if
// the fetch had failed" holds at that boundary rather than here.
//
// allowInsecureHTTP comes from
// OAuthClientIDMetadataDocumentFeatureConfig.IsInsecureHTTPAllowed() and
// relaxes rule 8's https requirement on the document's URI fields only.
// Like ParseCIMDClientID's equivalent parameter, it is explicit rather than
// read from config in here, so this function stays pure and every call
// site states its posture.
func ParseAndValidate(requestURL string, body []byte, allowInsecureHTTP bool) (*Document, error) {
	var raw rawDocument
	// Guard against a JSON array/string/number/null top level, which
	// unmarshals into a struct with an error, and against a valid object
	// whose every field is absent (caught by the redirect_uris check below
	// anyway).
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, errors.Join(ErrDocumentNotJSONObject, err)
	}

	// Rule 1: token_endpoint_auth_method, if present, MUST be "none".
	// Checked first, mirroring dcr.ValidateAndNormalize. Rejects
	// private_key_jwt and every client_secret_* variant.
	if raw.TokenEndpointAuthMethod != nil && *raw.TokenEndpointAuthMethod != "none" {
		return nil, ErrDocumentTokenEndpointAuthMethodNotAccepted
	}

	// Rule 2: client_id MUST be present and MUST equal requestURL
	// byte-for-byte. No normalization, no case folding, no trailing-slash
	// tolerance -- this single comparison is the whole reason one host
	// cannot vouch for another's identity.
	if raw.ClientID == nil || *raw.ClientID != requestURL {
		return nil, ErrDocumentClientIDMismatch
	}

	// Rule 3: application_type, if present, MUST be "web" or "native".
	applicationType := "web"
	if raw.ApplicationType != nil {
		applicationType = *raw.ApplicationType
	}
	if applicationType != "web" && applicationType != "native" {
		return nil, ErrDocumentApplicationTypeUnsupported
	}

	// Rule 4: redirect_uris MUST be present and non-empty; each entry
	// validated. application_type is NOT consulted -- see
	// validateCIMDRedirectURI.
	if len(raw.RedirectURIs) == 0 {
		return nil, ErrDocumentRedirectURIsMissing
	}
	for _, ru := range raw.RedirectURIs {
		if err := validateCIMDRedirectURI(ru); err != nil {
			return nil, err
		}
	}

	// Rule 5: grant_types entries Authgear does not implement are IGNORED.
	// A document that declared grant types and had ALL of them dropped
	// keeps the ErrDocumentGrantTypeUnsupported it has always returned;
	// everything else is left to Rules 6 and 7, exactly as before.
	//
	// Ignoring rather than rejecting is the whole point: a CIMD document is
	// ONE self-published description of a client, used against every
	// authorization server that client talks to, so it advertises every
	// grant the client can perform anywhere -- not the subset any single
	// server implements, and unlike a DCR request it cannot be tailored to
	// the server receiving it. Claude's document
	// (https://claude.ai/oauth/mcp-oauth-client-metadata) declares
	// urn:ietf:params:oauth:grant-type:jwt-bearer for MCP Enterprise
	// Managed Auth's identity-assertion exchange (RFC 7523) alongside the
	// authorization_code and refresh_token it uses everywhere else.
	// Rejecting the whole document over that one entry made every Claude
	// connector unresolvable, and it protected nothing: a grant Authgear
	// omits from grant_types_supported is never requested, and if one were
	// requested anyway the token endpoint answers unsupported_grant_type
	// on its own (see handler_token.go's validateRequestWithoutTx) --
	// the persisted grant_types are not what gates it.
	//
	// This is deliberately a pure relaxation: no document that resolved
	// before this rule existed stops resolving because of it. In
	// particular an explicitly empty grant_types is NOT an error here --
	// it never was, and Rule 7 keeps deciding it -- so the emptiness check
	// below is scoped to a list that had entries to lose.
	rawGrantTypes := raw.GrantTypes
	if rawGrantTypes == nil {
		rawGrantTypes = []string{"authorization_code", "refresh_token"}
	}
	grantTypes := make([]string, 0, len(rawGrantTypes))
	for _, gt := range rawGrantTypes {
		if cimdSupportedGrantTypes[gt] {
			grantTypes = append(grantTypes, gt)
		}
	}
	if len(grantTypes) == 0 && len(rawGrantTypes) > 0 {
		return nil, ErrDocumentGrantTypeUnsupported
	}

	// Rule 6: response_types MUST be a subset of ["code"].
	responseTypes := raw.ResponseTypes
	if responseTypes == nil {
		responseTypes = []string{"code"}
	}
	for _, rt := range responseTypes {
		if !cimdSupportedResponseTypes[rt] {
			return nil, ErrDocumentResponseTypeInconsistent
		}
	}

	// Rule 7: response_types MUST be consistent with grant_types. Same rule
	// as DCR (dcr/validate.go): contains(grantTypes, "authorization_code")
	// == contains(responseTypes, "code"). Checked against the FILTERED
	// grant types, so a document whose authorization_code was never
	// declared -- or that declared only refresh_token next to grants
	// Authgear does not implement -- is held to the same consistency it
	// always was: it needs a response_types that omits "code" too.
	hasAuthorizationCode := slices.Contains(grantTypes, "authorization_code")
	hasCode := slices.Contains(responseTypes, "code")
	if hasAuthorizationCode != hasCode {
		return nil, ErrDocumentResponseTypeInconsistent
	}

	// Rule 8: logo_uri, client_uri, tos_uri, policy_uri, if present, MUST
	// be https:// -- except http:// is accepted when allowInsecureHTTP is
	// set. Nothing else about these fields is relaxed.
	for _, uri := range []*string{raw.LogoURI, raw.ClientURI, raw.TOSURI, raw.PolicyURI} {
		if uri != nil {
			if err := validateCIMDHTTPSURI(*uri, allowInsecureHTTP); err != nil {
				return nil, err
			}
		}
	}

	// Rule 9: client_name is optional with no constraint. The
	// "Client <clientID>" fallback is NOT applied here; it is computed on
	// read by oauthclient.Client.DisplayName().
	return &Document{
		ClientName:      raw.ClientName,
		RedirectURIs:    raw.RedirectURIs,
		GrantTypes:      grantTypes,
		ResponseTypes:   responseTypes,
		ApplicationType: applicationType,
		LogoURI:         raw.LogoURI,
		ClientURI:       raw.ClientURI,
		TOSURI:          raw.TOSURI,
		PolicyURI:       raw.PolicyURI,
	}, nil
}

// validateCIMDRedirectURI implements spec § Accepted Metadata Fields --
// redirect_uris. Each entry must be an absolute URI with no fragment, and
// must be one of:
//
//   - https://
//   - a loopback http:// URI -- host localhost, 127.0.0.1 or [::1], ANY port
//   - a custom (non-http, non-https) URI scheme
//
// Unlike dcr's redirect URI rule, application_type is NOT a parameter. DCR
// gates http://localhost on application_type: native; CIMD cannot, because
// the MCP Authorization spec's own reference CIMD document uses
// http://127.0.0.1:3000/callback and http://localhost:3000/callback while
// omitting application_type entirely (so it defaults to "web"). Gating
// loopback the DCR way would reject the reference example outright.
func validateCIMDRedirectURI(raw string) error {
	u, err := url.Parse(raw)
	if err != nil || !u.IsAbs() {
		return ErrDocumentRedirectURIInvalid
	}
	if u.Fragment != "" || u.RawFragment != "" {
		return ErrDocumentRedirectURIInvalid
	}
	switch strings.ToLower(u.Scheme) {
	case "https":
		if u.Host == "" {
			return ErrDocumentRedirectURIInvalid
		}
		return nil
	case "http":
		// url.URL.Hostname() strips the brackets from "[::1]:3000", so this
		// matches "http://[::1]:3000/callback" too. RFC 8252 §7.3 treats
		// both IPv4 and IPv6 loopback as loopback, and an IPv6-only
		// developer machine has no 127.0.0.1 to listen on.
		switch u.Hostname() {
		case "localhost", "127.0.0.1", "::1":
			return nil
		default:
			return ErrDocumentRedirectURIInvalid
		}
	default:
		// Any custom scheme (com.example.app:/callback, myapp://cb, ...).
		return nil
	}
}

// validateCIMDHTTPSURI implements rule 8. Mirrors dcr's validateHTTPSURI
// except for the allowInsecureHTTP relaxation.
func validateCIMDHTTPSURI(raw string, allowInsecureHTTP bool) error {
	u, err := url.Parse(raw)
	if err != nil {
		return ErrDocumentURIFieldNotHTTPS
	}
	switch {
	case strings.EqualFold(u.Scheme, "https"):
		return nil
	case allowInsecureHTTP && strings.EqualFold(u.Scheme, "http"):
		return nil
	default:
		return ErrDocumentURIFieldNotHTTPS
	}
}
