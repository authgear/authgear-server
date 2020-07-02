package oauth

import (
	"net/http"
	"strings"

	"github.com/authgear/authgear-server/pkg/auth/dependency/auth"
	"github.com/authgear/authgear-server/pkg/auth/dependency/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/auth/dependency/session"
)

const FullAccessScope = "https://authgear.com/scopes/full-access"

func SessionScopes(s auth.AuthSession) []string {
	switch s := s.(type) {
	case *session.IDPSession:
		return []string{FullAccessScope}
	case *OfflineGrant:
		return s.Scopes
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
			session := auth.GetSession(r.Context())
			status, errResp := checkAuthz(session, requiredScopes, scope)
			if errResp != nil {
				h := errResp.ToWWWAuthenticateHeader()
				rw.Header().Add("WWW-Authenticate", h)
				rw.WriteHeader(status)
				return
			}
			next.ServeHTTP(rw, r)
		})
	}
}

func checkAuthz(session auth.AuthSession, requiredScopes map[string]struct{}, scope string) (int, protocol.ErrorResponse) {
	if session == nil {
		return http.StatusUnauthorized, protocol.NewErrorResponse("invalid_token", "invalid access token")
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
