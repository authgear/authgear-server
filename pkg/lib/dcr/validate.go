package dcr

import (
	"errors"
	"net/url"
)

var (
	ErrDCRRedirectURIsMissing = errors.New("dcr: redirect_uris is required")
	ErrDCRRedirectURIInvalid  = errors.New("dcr: invalid redirect_uri")
	// ErrDCRTokenEndpointAuthMethodNotAccepted is returned when
	// token_endpoint_auth_method is present and is anything other than
	// "none" -- every DCR-registered client is public and uses PKCE, so
	// "none" is accepted (it's simply what every such client already is),
	// but "client_secret_post"/"client_secret_basic" are rejected since
	// Authgear never issues a client_secret via DCR.
	ErrDCRTokenEndpointAuthMethodNotAccepted = errors.New("dcr: token_endpoint_auth_method is not accepted")
	ErrDCRGrantTypeUnsupported               = errors.New("dcr: unsupported grant_type")
	ErrDCRResponseTypeInconsistent           = errors.New("dcr: response_types is inconsistent with grant_types")
	ErrDCRApplicationTypeUnsupported         = errors.New("dcr: unsupported application_type")
	ErrDCRURIFieldNotHTTPS                   = errors.New("dcr: uri field must use https")
)

// RegistrationRequest is the parsed (but not yet validated/normalized)
// POST /oauth2/register request body.
type RegistrationRequest struct {
	ClientName              *string
	RedirectURIs            []string
	GrantTypes              []string // nil means "not provided" (apply default)
	ResponseTypes           []string // nil means "not provided" (apply default)
	ApplicationType         *string
	LogoURI                 *string
	ClientURI               *string
	TOSURI                  *string
	PolicyURI               *string
	TokenEndpointAuthMethod *string // only used for rejecting the request if present and not "none"
}

// NormalizedRegistration is a RegistrationRequest after defaults have been
// applied and every field has passed validation.
type NormalizedRegistration struct {
	ClientName    *string
	RedirectURIs  []string
	GrantTypes    []string
	ResponseTypes []string
	// ApplicationType is always "web" or "native" — never nil, never any
	// other value; ValidateAndNormalize rejects anything else.
	ApplicationType string
	LogoURI         *string
	ClientURI       *string
	TOSURI          *string
	PolicyURI       *string
}

var dcrSupportedGrantTypes = map[string]bool{
	"authorization_code": true,
	"refresh_token":      true,
}

var dcrSupportedResponseTypes = map[string]bool{
	"code": true,
}

// ValidateAndNormalize implements every rule in
// docs/specs/dcr.md#accepted-client-metadata and
// docs/specs/dcr.md#errors. It applies defaults (grant_types,
// response_types, application_type) and returns one of the sentinel
// errors above on the first rule violated — the HTTP handler layer maps
// each to its exact RFC 7591 (error, status) pair.
func ValidateAndNormalize(req *RegistrationRequest) (*NormalizedRegistration, error) {
	if req.TokenEndpointAuthMethod != nil && *req.TokenEndpointAuthMethod != "none" {
		return nil, ErrDCRTokenEndpointAuthMethodNotAccepted
	}

	applicationType := "web"
	if req.ApplicationType != nil {
		applicationType = *req.ApplicationType
	}
	if applicationType != "web" && applicationType != "native" {
		return nil, ErrDCRApplicationTypeUnsupported
	}

	if len(req.RedirectURIs) == 0 {
		return nil, ErrDCRRedirectURIsMissing
	}
	for _, raw := range req.RedirectURIs {
		if err := validateRedirectURI(raw, applicationType); err != nil {
			return nil, err
		}
	}

	grantTypes := req.GrantTypes
	if grantTypes == nil {
		grantTypes = []string{"authorization_code", "refresh_token"}
	}
	for _, gt := range grantTypes {
		if !dcrSupportedGrantTypes[gt] {
			return nil, ErrDCRGrantTypeUnsupported
		}
	}

	responseTypes := req.ResponseTypes
	if responseTypes == nil {
		responseTypes = []string{"code"}
	}
	for _, rt := range responseTypes {
		if !dcrSupportedResponseTypes[rt] {
			return nil, ErrDCRResponseTypeInconsistent
		}
	}

	hasAuthorizationCode := containsString(grantTypes, "authorization_code")
	hasCode := containsString(responseTypes, "code")
	if hasAuthorizationCode != hasCode {
		return nil, ErrDCRResponseTypeInconsistent
	}

	for _, uri := range []*string{req.LogoURI, req.ClientURI, req.TOSURI, req.PolicyURI} {
		if uri != nil {
			if err := validateHTTPSURI(*uri); err != nil {
				return nil, err
			}
		}
	}

	return &NormalizedRegistration{
		ClientName:      req.ClientName,
		RedirectURIs:    req.RedirectURIs,
		GrantTypes:      grantTypes,
		ResponseTypes:   responseTypes,
		ApplicationType: applicationType,
		LogoURI:         req.LogoURI,
		ClientURI:       req.ClientURI,
		TOSURI:          req.TOSURI,
		PolicyURI:       req.PolicyURI,
	}, nil
}

// validateRedirectURI implements the per-application_type redirect URI
// scheme rules from docs/specs/dcr.md's application_type table: web must
// use https://, localhost not allowed; native must use a custom URI
// scheme or http://localhost.
func validateRedirectURI(raw string, applicationType string) error {
	u, err := url.Parse(raw)
	if err != nil || !u.IsAbs() {
		return ErrDCRRedirectURIInvalid
	}
	if u.Fragment != "" {
		return ErrDCRRedirectURIInvalid
	}

	switch applicationType {
	case "web":
		if u.Scheme != "https" {
			return ErrDCRRedirectURIInvalid
		}
	case "native":
		switch u.Scheme {
		case "http":
			if u.Hostname() != "localhost" {
				return ErrDCRRedirectURIInvalid
			}
		case "https":
			return ErrDCRRedirectURIInvalid
		}
		// any other (custom) scheme is accepted
	}
	return nil
}

func validateHTTPSURI(raw string) error {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" {
		return ErrDCRURIFieldNotHTTPS
	}
	return nil
}

func containsString(ss []string, s string) bool {
	for _, x := range ss {
		if x == s {
			return true
		}
	}
	return false
}
