package handler

//go:generate go tool mockgen -source=handler_end_session.go -destination=handler_end_session_mock_test.go -package handler_test

import (
	"context"
	"net/http"
	"net/url"
	"slices"

	"github.com/lestrrat-go/jwx/v2/jwt"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oidc"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oidc/protocol"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/idpsession"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
	"github.com/authgear/authgear-server/pkg/util/urlutil"
)

var EndSessionHandlerLogger = slogutil.NewLogger("oidc-end-session")

type WebAppURLsProvider interface {
	LogoutURL(redirectURI *url.URL) *url.URL
	SettingsURL() *url.URL
}

type LogoutSessionManager interface {
	Logout(ctx context.Context, sessionBase session.SessionBase, w http.ResponseWriter) ([]session.ListableSession, error)
}

type CookieManager interface {
	GetCookie(r *http.Request, def *httputil.CookieDef) (*http.Cookie, error)
	ValueCookie(def *httputil.CookieDef, value string) *http.Cookie
	ClearCookie(def *httputil.CookieDef) *http.Cookie
}

// IDTokenVerifier verifies id_token_hint's signature so its claims can be
// trusted. It intentionally does not enforce exp, since id_token_hint must
// accept expired ID tokens.
type IDTokenVerifier interface {
	VerifyIDToken(idTokenHint string) (idToken jwt.Token, err error)
}

// IDTokenHintSessionProvider resolves id_token_hint's sid when it names an
// IDP session (session.TypeIdentityProvider).
type IDTokenHintSessionProvider interface {
	Get(ctx context.Context, id string) (*idpsession.IDPSession, error)
}

// IDTokenHintOfflineGrantService resolves id_token_hint's sid when it names
// an offline grant (session.TypeOfflineGrant) — the common case for a
// first-party client that requested offline_access, whose id_token is bound
// to its own refresh token session, not directly to the browser's IDP
// session cookie.
type IDTokenHintOfflineGrantService interface {
	GetOfflineGrant(ctx context.Context, id string) (*oauth.OfflineGrant, error)
}

type EndSessionHandler struct {
	Config           *config.OAuthConfig
	Endpoints        oidc.EndpointsProvider
	URLs             WebAppURLsProvider
	SessionManager   LogoutSessionManager
	SessionCookieDef session.CookieDef
	Cookies          CookieManager
	IDTokenVerifier  IDTokenVerifier
	Sessions         IDTokenHintSessionProvider
	OfflineGrants    IDTokenHintOfflineGrantService
}

func (h *EndSessionHandler) Handle(ctx context.Context, s session.ResolvedSession, req protocol.EndSessionRequest, r *http.Request, rw http.ResponseWriter) error {
	// Step 1: resume from a POST that was stashed on a previous pass through
	// this handler. Must run before anything else touches req or s.
	if sealed, hasStash := req[endSessionRefQueryParam]; hasStash {
		req = h.resumeFromStash(ctx, r, rw, sealed)
	} else if r.Method == http.MethodPost {
		// Step 2: a POST that hasn't been through the stash round trip yet.
		// The Lax session cookie may not be visible on this request even if
		// the end-user has a session (cross-site POST). Stash and force a
		// same-origin top-level GET so the Lax cookie becomes visible on the
		// next pass.
		key, sealed, err := sealEndSessionRequest(req)
		if err != nil {
			return err
		}
		httputil.UpdateCookie(rw, h.Cookies.ValueCookie(EndSessionRefKeyCookieDef, key))
		selfURL := urlutil.WithQueryParamsAdded(
			h.Endpoints.EndSessionEndpointURL(),
			map[string]string{endSessionRefQueryParam: sealed},
		)
		httputil.Redirect(ctx, rw, r, selfURL.String(), http.StatusFound)
		return nil
	}

	idTokenHint := req.IDTokenHint()

	if idTokenHint == "" {
		// Step 3: existing SameSiteStrict fast path (unrelated CSRF
		// safeguard, preserved as-is; covers same-site navigations, e.g. a
		// link from Authgear's own settings page). Only applies when the
		// caller didn't give an id_token_hint at all: once a hint is given,
		// its own resolution below is the sole authority on whether to log
		// out directly, regardless of this cookie.
		sameSiteStrict, err := h.Cookies.GetCookie(r, h.SessionCookieDef.SameSiteStrictDef)
		if s != nil && err == nil && sameSiteStrict.Value == "true" {
			// Logout directly.
			// TODO(SAML): Logout affected saml service providers
			_, err := h.SessionManager.Logout(ctx, s, rw)
			if err != nil {
				return err
			}
			// Set s to nil and fall through.
			s = nil
		}
	} else if s != nil {
		// Step 4: id_token_hint fast path (spec: sid matches the current
		// logged in IdP session, and first-party client => direct logout, no
		// confirmation). "Matches" means the same SSO group, not sid string
		// equality: a first-party client that requested offline_access (the
		// normal case) gets an id_token bound to its own offline grant, not
		// directly to the browser's IDP session cookie, but that offline
		// grant and the IDP session cookie both trace back to the same login
		// when the grant was issued with SSO enabled
		// (session.SessionBase.SSOGroupIDPSessionID), which is exactly what
		// IsSameSSOGroup checks. IssueOfflineGrantOptions.SSOEnabled
		// (pkg/lib/oauth/handler/handler_token.go) is set whenever the grant
		// has an IDPSessionID at all, not only when the client explicitly
		// requested x_sso_enabled=true, so this also covers the common case
		// of a client that never sends that Authgear-specific extension.
		if client, sidSession, ok := h.resolveIDTokenHintSession(ctx, idTokenHint); ok &&
			client.IsFirstParty() && sidSession.IsSameSSOGroup(s) {
			_, err := h.SessionManager.Logout(ctx, s, rw)
			if err != nil {
				return err
			}
			s = nil
		}
	}

	// Step 5: neither fast path fired; show the confirmation page. Strip
	// id_token_hint before forwarding: the /logout confirmation flow never
	// needs it (the direct-logout decision is already final by this point),
	// and forwarding it would re-expose it in the /logout?redirect_uri=<...>
	// URL, defeating the POST/stash mechanism above.
	if s != nil {
		endSessionURL := urlutil.WithQueryParamsAdded(
			h.Endpoints.EndSessionEndpointURL(),
			req.WithoutIDTokenHint(),
		)
		logoutURL := h.URLs.LogoutURL(endSessionURL)

		httputil.Redirect(ctx, rw, r, logoutURL.String(), http.StatusFound)
		return nil
	}

	redirectURI := req.PostLogoutRedirectURI()
	valid, client := h.validateRedirectURI(redirectURI)
	if !valid {
		// Invalid/empty redirect URI, redirect to home page/settings
		if client != nil && client.ClientURI != "" {
			redirectURI = client.ClientURI
		} else {
			redirectURI = h.URLs.SettingsURL().String()
		}
		http.Redirect(rw, r, redirectURI, http.StatusFound)
		return nil
	}

	if state := req.State(); state != "" {
		uri, err := url.Parse(redirectURI)
		if err != nil {
			return err
		}
		redirectURI = urlutil.WithQueryParamsAdded(uri, map[string]string{"state": state}).String()
	}

	redirectURIURL, err := url.Parse(redirectURI)
	if err != nil {
		panic(err)
	}

	writeResponseOptions := oauth.WriteResponseOptions{
		RedirectURI:  redirectURIURL,
		ResponseMode: "query",
		UseHTTP200:   client.UseHTTP200(),
		Response:     make(map[string]string),
	}
	oauth.WriteResponse(rw, r, writeResponseOptions)
	return nil
}

// resumeFromStash opens the request stashed by an earlier POST pass through
// this handler (see end_session_stash.go). If the stash cannot be opened —
// the cookie is missing, or doesn't match the sealed value in the URL — this
// is expected to happen in normal use (e.g. the end-user revisits a stale
// link from browser history after the short-lived stash cookie has already
// expired), so rather than surfacing an error it is logged and treated as an
// end_session request with no parameters at all: the rest of Handle already
// knows how to handle that (confirmation page if a session is present,
// otherwise straight through, same as any other parameterless call).
func (h *EndSessionHandler) resumeFromStash(ctx context.Context, r *http.Request, rw http.ResponseWriter, sealed string) protocol.EndSessionRequest {
	// A closure, not a bare `defer httputil.UpdateCookie(rw,
	// h.Cookies.ClearCookie(...))`: defer only postpones the outer call, not
	// evaluation of its arguments, so ClearCookie() would otherwise run
	// immediately, before GetCookie() below gets a chance to read the
	// still-live cookie.
	defer func() {
		httputil.UpdateCookie(rw, h.Cookies.ClearCookie(EndSessionRefKeyCookieDef))
	}()

	cookie, err := h.Cookies.GetCookie(r, EndSessionRefKeyCookieDef)
	if err != nil {
		h.logInvalidStash(ctx, err)
		return protocol.EndSessionRequest{}
	}

	opened, err := openEndSessionRequest(cookie.Value, sealed)
	if err != nil {
		h.logInvalidStash(ctx, err)
		return protocol.EndSessionRequest{}
	}

	return opened
}

func (h *EndSessionHandler) logInvalidStash(ctx context.Context, cause error) {
	logger := EndSessionHandlerLogger.GetLogger(ctx)
	logger.WithError(cause).Warn(ctx, "end_session: invalid or expired logout stash")
}

func (h *EndSessionHandler) validateRedirectURI(redirectURI string) (valid bool, client *config.OAuthClientConfig) {
	for _, client := range h.Config.Clients {
		if slices.Contains(client.PostLogoutRedirectURIs, redirectURI) {
			return true, &client
		}
	}
	return false, nil
}

// resolveIDTokenHintSession verifies idTokenHint's signature, extracts its
// aud (client_id) claim, and resolves its sid claim to the actual session or
// offline grant it names. ok is false if the token doesn't verify, is
// missing either claim, names a client that no longer exists, or names a
// session/offline grant that no longer exists — in all such cases the caller
// must treat it exactly like "no id_token_hint" (fall through to the
// confirmation page), not as an error: an unrecognized or malformed hint is
// not proof of anything, but it is also not a protocol violation by itself.
func (h *EndSessionHandler) resolveIDTokenHintSession(ctx context.Context, idTokenHint string) (client *config.OAuthClientConfig, sidSession session.ListableSession, ok bool) {
	idToken, err := h.IDTokenVerifier.VerifyIDToken(idTokenHint)
	if err != nil {
		return nil, nil, false
	}

	sidInterface, hasSID := idToken.Get(string(model.ClaimSID))
	sidStr, isString := sidInterface.(string)
	if !hasSID || !isString || sidStr == "" {
		return nil, nil, false
	}

	aud := idToken.Audience()
	if len(aud) != 1 {
		return nil, nil, false
	}

	client, ok = h.Config.GetClient(aud[0])
	if !ok {
		return nil, nil, false
	}

	typ, sessionID, ok := oauth.DecodeSID(sidStr)
	if !ok {
		return nil, nil, false
	}

	switch typ {
	case session.TypeIdentityProvider:
		sess, err := h.Sessions.Get(ctx, sessionID)
		if err != nil {
			return nil, nil, false
		}
		sidSession = sess
	case session.TypeOfflineGrant:
		grant, err := h.OfflineGrants.GetOfflineGrant(ctx, sessionID)
		if err != nil {
			return nil, nil, false
		}
		sidSession = grant
	default:
		return nil, nil, false
	}

	return client, sidSession, true
}
