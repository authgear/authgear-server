package oauth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/lib/session"
)

const FullAccessScope = "https://authgear.com/scopes/full-access"
const FullUserInfoScope = "https://authgear.com/scopes/full-userinfo"
const PreAuthenticatedURLScope = "https://authgear.com/scopes/pre-authenticated-url"
const OfflineAccess = "offline_access"
const DeviceSSOScope = "device_sso"

const (
	// The scope openid must be present.
	// https://openid.net/specs/openid-connect-core-1_0.html#AuthRequest
	ScopeOpenID = "openid"
	// Scope "profile" is defined in
	// https://openid.net/specs/openid-connect-core-1_0.html#ScopeClaims
	ScopeProfile = "profile"
	// Scope "email" is defined in
	// https://openid.net/specs/openid-connect-core-1_0.html#ScopeClaims
	ScopeEmail = "email"
	// Scope "address" is defined in
	// https://openid.net/specs/openid-connect-core-1_0.html#ScopeClaims
	ScopeAddress = "address"
	// Scope "phone" is defined in
	// https://openid.net/specs/openid-connect-core-1_0.html#ScopeClaims
	ScopePhone = "phone"
)

var AllowedScopes = []string{
	// OAuth 2.0 scopes
	OfflineAccess,
	DeviceSSOScope,

	// OIDC scopes.
	ScopeOpenID,
	ScopeProfile,
	ScopeEmail,
	ScopeAddress,
	ScopePhone,

	// Authgear specific scopes.
	FullAccessScope,
	FullUserInfoScope,
	PreAuthenticatedURLScope,
}

// IsResourceScope reports whether scope is resource-specific rather than
// project-level. AllowedScopes is the closed set of project-level scopes,
// and ValidateScopesByClientConfig admits a resource scope by appending the
// Resource's own scopes to it -- so "not in AllowedScopes" is the same
// discriminator, used from the other side.
func IsResourceScope(scope string) bool {
	return !slices.Contains(AllowedScopes, scope)
}

var scopeClaims = map[string]map[string]struct{}{
	ScopeProfile: {
		// https://openid.net/specs/openid-connect-core-1_0.html#ScopeClaims
		stdattrs.Name:              {},
		stdattrs.FamilyName:        {},
		stdattrs.GivenName:         {},
		stdattrs.MiddleName:        {},
		stdattrs.Nickname:          {},
		stdattrs.PreferredUsername: {},
		stdattrs.Profile:           {},
		stdattrs.Picture:           {},
		stdattrs.Website:           {},
		stdattrs.Gender:            {},
		stdattrs.Birthdate:         {},
		stdattrs.Zoneinfo:          {},
		stdattrs.Locale:            {},
		stdattrs.UpdatedAt:         {},
	},
	ScopeEmail: {
		// https://openid.net/specs/openid-connect-core-1_0.html#ScopeClaims
		stdattrs.Email:         {},
		stdattrs.EmailVerified: {},
	},
	ScopeAddress: {
		// https://openid.net/specs/openid-connect-core-1_0.html#ScopeClaims
		stdattrs.Address: {},
	},
	ScopePhone: {
		// https://openid.net/specs/openid-connect-core-1_0.html#ScopeClaims
		stdattrs.PhoneNumber:         {},
		stdattrs.PhoneNumberVerified: {},
	},
}

func SessionScopes(s session.ResolvedSession) []string {
	if s == nil {
		return []string{}
	}
	switch s.SessionType() {
	case session.TypeIdentityProvider:
		return []string{FullAccessScope, PreAuthenticatedURLScope}
	case session.TypeOfflineGrant:
		ss := s.(*OfflineGrantSession)
		return ss.Scopes
	default:
		panic("oauth: unexpected session type")
	}
}

// RequireScope allow request to pass if session contains one of the required scopes.
// If there is no required scopes, only validity of session is checked.
func RequireScope(scopes ...string) func(http.Handler) http.Handler {
	requiredScopes := map[string]struct{}{}
	for _, s := range scopes {
		requiredScopes[s] = struct{}{}
	}
	scope := strings.Join(scopes, " ")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			session := session.GetSession(r.Context())
			status, errResp := checkAuthz(session, requiredScopes, scope)
			if errResp != nil {
				h := errResp.ToWWWAuthenticateHeader()
				rw.Header().Add("WWW-Authenticate", h)
				rw.WriteHeader(status)
				encoder := json.NewEncoder(rw)
				err := encoder.Encode(errResp)
				if err != nil {
					http.Error(rw, err.Error(), 500)
				}
				return
			}
			next.ServeHTTP(rw, r)
		})
	}
}

func checkAuthz(session session.ResolvedSession, requiredScopes map[string]struct{}, scope string) (int, protocol.ErrorResponse) {
	if session == nil {
		return http.StatusUnauthorized, protocol.NewErrorResponse("invalid_grant", "invalid session")
	}

	// Check scopes only if there are required scopes.
	if len(requiredScopes) > 0 {
		sessionScopes := SessionScopes(session)
		pass := false
		for _, s := range sessionScopes {
			if _, ok := requiredScopes[s]; ok {
				pass = true
				break
			}
		}

		if !pass {
			resp := protocol.NewErrorResponse("insufficient_scope", "required scope not granted")
			resp["scope"] = scope
			return http.StatusForbidden, resp
		}
	}

	return http.StatusOK, nil
}

func ContainsAllScopes(scopes []string, shouldContainsScopes []string) bool {
	scopesSet := map[string]struct{}{}
	for _, scope := range scopes {
		scopesSet[scope] = struct{}{}
	}
	for _, scope := range shouldContainsScopes {
		if _, exist := scopesSet[scope]; !exist {
			return false
		}
	}
	return true
}

func IsScopeAllowed(scope string, allowedScopes []string) bool {
	return slices.Contains(allowedScopes, scope)
}

func ScopeAllowsClaim(scope string, claimName string) bool {
	// Empty claim is never allowed.
	if claimName == "" {
		return false
	}

	switch scope {
	case FullAccessScope:
		// full access scope allows everything.
		return true
	case FullUserInfoScope:
		// full user info scope allows everything.
		return true
	case ScopeProfile:
		_, ok := scopeClaims[ScopeProfile][claimName]
		return ok
	case ScopeEmail:
		_, ok := scopeClaims[ScopeEmail][claimName]
		return ok
	case ScopePhone:
		_, ok := scopeClaims[ScopePhone][claimName]
		return ok
	case ScopeAddress:
		_, ok := scopeClaims[ScopeAddress][claimName]
		return ok
	default:
		// Other scope does not allow any claim.
		return false
	}
}

// ValidateScopes returns invalid_scope for the first scope not in knownScopes.
func ValidateScopes(scopes []string, knownScopes []string) error {
	for _, s := range scopes {
		if !IsScopeAllowed(s, knownScopes) {
			return protocol.NewError("invalid_scope", fmt.Sprintf("unknown scope: %s", s))
		}
	}
	return nil
}

// ValidateScopesByClientConfig rejects unknown scopes and malformed scope
// combinations. knownResourceScopes is every scope of the requested
// resource, or nil when no resource= was requested. Whether the client is
// granted each scope is decided by GrantedScopes.
func ValidateScopesByClientConfig(scopes []string, knownResourceScopes []string) error {
	knownScopes := AllowedScopes
	if len(knownResourceScopes) > 0 {
		knownScopes = append(slices.Clone(AllowedScopes), knownResourceScopes...)
	}
	if err := ValidateScopes(scopes, knownScopes); err != nil {
		return err
	}
	if slices.Contains(scopes, PreAuthenticatedURLScope) && !slices.Contains(scopes, DeviceSSOScope) {
		return protocol.NewError("invalid_scope", "device_sso must be requested when using pre-authenticated url")
	}
	if !slices.Contains(scopes, ScopeOpenID) {
		return protocol.NewError("invalid_scope", "must request 'openid' scope")
	}
	return nil
}

// GrantedScopes returns scopes without those the client is not granted.
// grantedResourceScopes is the scopes of the requested resource opened to
// the client. See docs/specs/oidc.md § Scope Validation.
func GrantedScopes(client *config.OAuthClientConfig, scopes []string, grantedResourceScopes []string) []string {
	var granted []string
	for _, s := range scopes {
		if IsResourceScope(s) {
			if slices.Contains(grantedResourceScopes, s) {
				granted = append(granted, s)
			}
		} else if isBuiltinScopeGranted(client, s) {
			granted = append(granted, s)
		}
	}
	return granted
}

func isBuiltinScopeGranted(client *config.OAuthClientConfig, scope string) bool {
	if client.ApplicationType == config.OAuthClientApplicationTypeM2M {
		return false
	}
	switch scope {
	case OfflineAccess:
		return slices.Contains(GetAllowedGrantTypes(client), RefreshTokenGrantType)
	case FullAccessScope:
		return client.HasFullAccessScope()
	case FullUserInfoScope:
		return client.IsFirstParty()
	case DeviceSSOScope, PreAuthenticatedURLScope:
		return client.PreAuthenticatedURLEnabled
	default:
		return true
	}
}
