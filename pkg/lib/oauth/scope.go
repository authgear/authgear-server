package oauth

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/lib/session"
)

const FullAccessScope = "https://authgear.com/scopes/full-access"
const FullUserInfoScope = "https://authgear.com/scopes/full-userinfo"
const AppInitiatedSSOToWebScope = "https://authgear.com/scopes/app-initiated-sso-to-web"
const OfflineAccess = "offline_access"
const DeviceSSOScope = "device_sso"

func SessionScopes(s session.ResolvedSession) []string {
	if s == nil {
		return []string{}
	}
	switch s.SessionType() {
	case session.TypeIdentityProvider:
		return []string{FullAccessScope, AppInitiatedSSOToWebScope}
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
