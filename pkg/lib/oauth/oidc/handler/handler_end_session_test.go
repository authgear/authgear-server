package handler_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/lestrrat-go/jwx/v2/jwt"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oidc/handler"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oidc/protocol"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/idpsession"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

// endSessionRefQueryParam mirrors the unexported constant of the same name in
// end_session_stash.go; this external test package cannot reference it
// directly.
const endSessionRefQueryParam = "x_end_session_ref"

type fakeEndpointsProvider struct{}

func (fakeEndpointsProvider) Origin() *url.URL {
	u, _ := url.Parse("https://app.example.com")
	return u
}

func (fakeEndpointsProvider) JWKSEndpointURL() *url.URL {
	u, _ := url.Parse("https://app.example.com/oauth2/jwks")
	return u
}

func (fakeEndpointsProvider) UserInfoEndpointURL() *url.URL {
	u, _ := url.Parse("https://app.example.com/oauth2/userinfo")
	return u
}

func (fakeEndpointsProvider) EndSessionEndpointURL() *url.URL {
	u, _ := url.Parse("https://app.example.com/oauth2/end_session")
	return u
}

// fakeCookieManager is a stateful in-memory stand-in for httputil.CookieManager.
// Unlike a gomock, it actually remembers what was set via ValueCookie so a
// value set on one Handle call can be read back via GetCookie on a later
// Handle call, which is what the POST -> stash -> resumed GET tests need.
type fakeCookieManager struct {
	values map[string]string
}

func newFakeCookieManager() *fakeCookieManager {
	return &fakeCookieManager{values: map[string]string{}}
}

func (m *fakeCookieManager) Set(nameSuffix string, value string) {
	m.values[nameSuffix] = value
}

func (m *fakeCookieManager) GetCookie(r *http.Request, def *httputil.CookieDef) (*http.Cookie, error) {
	v, ok := m.values[def.NameSuffix]
	if !ok {
		return nil, http.ErrNoCookie
	}
	return &http.Cookie{Name: def.NameSuffix, Value: v}, nil
}

func (m *fakeCookieManager) ValueCookie(def *httputil.CookieDef, value string) *http.Cookie {
	m.values[def.NameSuffix] = value
	return &http.Cookie{Name: def.NameSuffix, Value: value}
}

func (m *fakeCookieManager) ClearCookie(def *httputil.CookieDef) *http.Cookie {
	delete(m.values, def.NameSuffix)
	return &http.Cookie{Name: def.NameSuffix, Value: "", MaxAge: -1}
}

func newIDToken(sid string, clientID string) jwt.Token {
	token := jwt.New()
	if sid != "" {
		_ = token.Set(string(model.ClaimSID), sid)
	}
	if clientID != "" {
		_ = token.Set(jwt.AudienceKey, []string{clientID})
	}
	return token
}

func TestEndSessionHandlerHandle(t *testing.T) {
	firstPartyClient := config.OAuthClientConfig{
		ClientID:               "first-party-client",
		ApplicationType:        config.OAuthClientApplicationTypeSPA,
		PostLogoutRedirectURIs: []string{"https://rp.example.com/after-logout"},
	}
	thirdPartyClient := config.OAuthClientConfig{
		ClientID:               "third-party-client",
		ApplicationType:        config.OAuthClientApplicationTypeThirdPartyApp,
		PostLogoutRedirectURIs: []string{"https://thirdparty.example.com/after-logout"},
	}
	oauthConfig := &config.OAuthConfig{
		Clients: []config.OAuthClientConfig{firstPartyClient, thirdPartyClient},
	}
	sessionCookieDef := session.NewSessionCookieDef(&config.SessionConfig{})

	// sess is a real *idpsession.IDPSession, not a hand-rolled test double:
	// SSOGroupIDPSessionID() must return its own SessionID (as the real type
	// does) for the IsSameSSOGroup checks below to mean anything.
	sess := &idpsession.IDPSession{ID: "session-id"}
	sessOfflineGrantSID := oauth.EncodeSIDByRawValues(session.TypeOfflineGrant, "grant-same-login")

	newHandler := func(
		cookies *fakeCookieManager,
		sessionManager handler.LogoutSessionManager,
		urls handler.WebAppURLsProvider,
		idTokenVerifier handler.IDTokenVerifier,
		sessions handler.IDTokenHintSessionProvider,
		offlineGrants handler.IDTokenHintOfflineGrantService,
	) *handler.EndSessionHandler {
		return &handler.EndSessionHandler{
			Config:           oauthConfig,
			Endpoints:        fakeEndpointsProvider{},
			URLs:             urls,
			SessionManager:   sessionManager,
			SessionCookieDef: sessionCookieDef,
			Cookies:          cookies,
			IDTokenVerifier:  idTokenVerifier,
			Sessions:         sessions,
			OfflineGrants:    offlineGrants,
		}
	}

	Convey("EndSessionHandler.Handle", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		cookies := newFakeCookieManager()
		sessionManager := NewMockLogoutSessionManager(ctrl)
		urls := NewMockWebAppURLsProvider(ctrl)
		idTokenVerifier := NewMockIDTokenVerifier(ctrl)
		sessions := NewMockIDTokenHintSessionProvider(ctrl)
		offlineGrants := NewMockIDTokenHintOfflineGrantService(ctrl)

		h := newHandler(cookies, sessionManager, urls, idTokenVerifier, sessions, offlineGrants)

		// expectSameLoginOfflineGrant sets up the normal, spec-intended case:
		// a first-party client requested offline_access, so its id_token's
		// sid is bound to that offline grant, not directly to the IDP
		// session cookie — but the grant was issued SSOEnabled and from the
		// same login (same IDPSessionID as sess), so it is in the same SSO
		// group as sess.
		expectSameLoginOfflineGrant := func() {
			offlineGrants.EXPECT().GetOfflineGrant(gomock.Any(), "grant-same-login").Return(&oauth.OfflineGrant{
				ID:           "grant-same-login",
				IDPSessionID: sess.ID,
				SSOEnabled:   true,
			}, nil)
		}

		Convey("GET, valid id_token_hint bound to an offline grant in the same SSO group, first-party client: silent logout", func() {
			idTokenVerifier.EXPECT().VerifyIDToken("valid-hint").Return(newIDToken(sessOfflineGrantSID, firstPartyClient.ClientID), nil)
			expectSameLoginOfflineGrant()
			sessionManager.EXPECT().Logout(gomock.Any(), sess, gomock.Any()).Return(nil, nil)

			req := protocol.EndSessionRequest{
				"id_token_hint":            "valid-hint",
				"post_logout_redirect_uri": "https://rp.example.com/after-logout",
			}
			r := httptest.NewRequest(http.MethodGet, "https://app.example.com/oauth2/end_session", nil)
			rw := httptest.NewRecorder()

			err := h.Handle(context.Background(), sess, req, r, rw)
			So(err, ShouldBeNil)
			So(rw.Code, ShouldEqual, http.StatusSeeOther)

			loc, parseErr := url.Parse(rw.Header().Get("Location"))
			So(parseErr, ShouldBeNil)
			So(loc.Scheme+"://"+loc.Host+loc.Path, ShouldEqual, "https://rp.example.com/after-logout")
		})

		Convey("GET, valid id_token_hint whose sid is the IDP session directly, first-party client: silent logout", func() {
			// Less common than the offline-grant case above, but the same
			// decision must hold when id_token_hint's sid names the IDP
			// session directly (session.TypeIdentityProvider branch of
			// resolveIDTokenHintSession) rather than an offline grant.
			idTokenVerifier.EXPECT().VerifyIDToken("valid-hint").Return(newIDToken(oauth.EncodeSID(sess), firstPartyClient.ClientID), nil)
			sessions.EXPECT().Get(gomock.Any(), sess.ID).Return(&idpsession.IDPSession{ID: sess.ID}, nil)
			sessionManager.EXPECT().Logout(gomock.Any(), sess, gomock.Any()).Return(nil, nil)

			req := protocol.EndSessionRequest{
				"id_token_hint":            "valid-hint",
				"post_logout_redirect_uri": "https://rp.example.com/after-logout",
			}
			r := httptest.NewRequest(http.MethodGet, "https://app.example.com/oauth2/end_session", nil)
			rw := httptest.NewRecorder()

			err := h.Handle(context.Background(), sess, req, r, rw)
			So(err, ShouldBeNil)
			So(rw.Code, ShouldEqual, http.StatusSeeOther)
		})

		Convey("GET, valid id_token_hint bound to an offline grant in the same SSO group, third-party client: confirmation page", func() {
			idTokenVerifier.EXPECT().VerifyIDToken("valid-hint").Return(newIDToken(sessOfflineGrantSID, thirdPartyClient.ClientID), nil)
			expectSameLoginOfflineGrant()

			var capturedEndSessionURL *url.URL
			urls.EXPECT().LogoutURL(gomock.Any()).DoAndReturn(func(u *url.URL) *url.URL {
				capturedEndSessionURL = u
				result, _ := url.Parse("https://app.example.com/logout?redirect_uri=" + url.QueryEscape(u.String()))
				return result
			})

			req := protocol.EndSessionRequest{
				"id_token_hint":            "valid-hint",
				"post_logout_redirect_uri": "https://thirdparty.example.com/after-logout",
			}
			r := httptest.NewRequest(http.MethodGet, "https://app.example.com/oauth2/end_session", nil)
			rw := httptest.NewRecorder()

			err := h.Handle(context.Background(), sess, req, r, rw)
			So(err, ShouldBeNil)
			So(rw.Code, ShouldEqual, http.StatusFound)
			So(capturedEndSessionURL, ShouldNotBeNil)
			So(capturedEndSessionURL.Query().Get("id_token_hint"), ShouldEqual, "")
		})

		Convey("GET, valid id_token_hint bound to an offline grant from a different login: confirmation page", func() {
			// Same shape as the matching case, but the grant's IDPSessionID
			// names a different login: not the same SSO group as sess.
			idTokenVerifier.EXPECT().VerifyIDToken("valid-hint").Return(newIDToken(sessOfflineGrantSID, firstPartyClient.ClientID), nil)
			offlineGrants.EXPECT().GetOfflineGrant(gomock.Any(), "grant-same-login").Return(&oauth.OfflineGrant{
				ID:           "grant-same-login",
				IDPSessionID: "some-other-session",
				SSOEnabled:   true,
			}, nil)

			var capturedEndSessionURL *url.URL
			urls.EXPECT().LogoutURL(gomock.Any()).DoAndReturn(func(u *url.URL) *url.URL {
				capturedEndSessionURL = u
				result, _ := url.Parse("https://app.example.com/logout?redirect_uri=" + url.QueryEscape(u.String()))
				return result
			})

			req := protocol.EndSessionRequest{
				"id_token_hint":            "valid-hint",
				"post_logout_redirect_uri": "https://rp.example.com/after-logout",
			}
			r := httptest.NewRequest(http.MethodGet, "https://app.example.com/oauth2/end_session", nil)
			rw := httptest.NewRecorder()

			err := h.Handle(context.Background(), sess, req, r, rw)
			So(err, ShouldBeNil)
			So(rw.Code, ShouldEqual, http.StatusFound)
			So(capturedEndSessionURL.Query().Get("id_token_hint"), ShouldEqual, "")
		})

		Convey("GET, valid id_token_hint bound to an offline grant from the same login but not SSO enabled: silent logout", func() {
			// IsSameSSOGroup matches on IDPSessionID equality alone,
			// regardless of this grant's own SSOEnabled: a client that never
			// requested x_sso_enabled (the common case for any RP unaware of
			// Authgear's extension) still authenticated through this exact
			// IDP session, and that alone is what makes it part of this
			// session's group — creating the grant from that session is
			// what "SSO" means here, not a separate opt-in the client must
			// also request.
			idTokenVerifier.EXPECT().VerifyIDToken("valid-hint").Return(newIDToken(sessOfflineGrantSID, firstPartyClient.ClientID), nil)
			offlineGrants.EXPECT().GetOfflineGrant(gomock.Any(), "grant-same-login").Return(&oauth.OfflineGrant{
				ID:           "grant-same-login",
				IDPSessionID: sess.ID,
				SSOEnabled:   false,
			}, nil)
			sessionManager.EXPECT().Logout(gomock.Any(), sess, gomock.Any()).Return(nil, nil)

			req := protocol.EndSessionRequest{
				"id_token_hint":            "valid-hint",
				"post_logout_redirect_uri": "https://rp.example.com/after-logout",
			}
			r := httptest.NewRequest(http.MethodGet, "https://app.example.com/oauth2/end_session", nil)
			rw := httptest.NewRecorder()

			err := h.Handle(context.Background(), sess, req, r, rw)
			So(err, ShouldBeNil)
			So(rw.Code, ShouldEqual, http.StatusSeeOther)
		})

		Convey("GET, valid id_token_hint bound to an offline grant from a different login and not SSO enabled: confirmation page", func() {
			// IDPSessionID names a different session entirely: this is a
			// genuinely unrelated grant, not just one that opted out of SSO
			// sharing.
			idTokenVerifier.EXPECT().VerifyIDToken("valid-hint").Return(newIDToken(sessOfflineGrantSID, firstPartyClient.ClientID), nil)
			offlineGrants.EXPECT().GetOfflineGrant(gomock.Any(), "grant-same-login").Return(&oauth.OfflineGrant{
				ID:           "grant-same-login",
				IDPSessionID: "some-other-session",
				SSOEnabled:   false,
			}, nil)

			urls.EXPECT().LogoutURL(gomock.Any()).DoAndReturn(func(u *url.URL) *url.URL {
				result, _ := url.Parse("https://app.example.com/logout?redirect_uri=" + url.QueryEscape(u.String()))
				return result
			})

			req := protocol.EndSessionRequest{
				"id_token_hint":            "valid-hint",
				"post_logout_redirect_uri": "https://rp.example.com/after-logout",
			}
			r := httptest.NewRequest(http.MethodGet, "https://app.example.com/oauth2/end_session", nil)
			rw := httptest.NewRecorder()

			err := h.Handle(context.Background(), sess, req, r, rw)
			So(err, ShouldBeNil)
			So(rw.Code, ShouldEqual, http.StatusFound)
		})

		Convey("GET, no id_token_hint, SameSiteStrict cookie true, session present: silent logout", func() {
			cookies.Set(sessionCookieDef.SameSiteStrictDef.NameSuffix, "true")
			sessionManager.EXPECT().Logout(gomock.Any(), sess, gomock.Any()).Return(nil, nil)

			req := protocol.EndSessionRequest{
				"post_logout_redirect_uri": "https://rp.example.com/after-logout",
			}
			r := httptest.NewRequest(http.MethodGet, "https://app.example.com/oauth2/end_session", nil)
			rw := httptest.NewRecorder()

			err := h.Handle(context.Background(), sess, req, r, rw)
			So(err, ShouldBeNil)
			So(rw.Code, ShouldEqual, http.StatusSeeOther)
		})

		Convey("GET, id_token_hint given but unresolvable, SameSiteStrict cookie true: confirmation page, SameSiteStrict fast path not used", func() {
			// Once id_token_hint is present at all, its own resolution is
			// the sole authority on whether to log out directly: the
			// SameSiteStrict cookie must not be consulted, even though it
			// would have logged out unconditionally had id_token_hint been
			// absent (see the "no id_token_hint" case above).
			cookies.Set(sessionCookieDef.SameSiteStrictDef.NameSuffix, "true")
			idTokenVerifier.EXPECT().VerifyIDToken("garbage").Return(nil, errors.New("bad signature"))

			urls.EXPECT().LogoutURL(gomock.Any()).DoAndReturn(func(u *url.URL) *url.URL {
				result, _ := url.Parse("https://app.example.com/logout?redirect_uri=" + url.QueryEscape(u.String()))
				return result
			})

			req := protocol.EndSessionRequest{
				"id_token_hint":            "garbage",
				"post_logout_redirect_uri": "https://rp.example.com/after-logout",
			}
			r := httptest.NewRequest(http.MethodGet, "https://app.example.com/oauth2/end_session", nil)
			rw := httptest.NewRecorder()

			err := h.Handle(context.Background(), sess, req, r, rw)
			So(err, ShouldBeNil)
			So(rw.Code, ShouldEqual, http.StatusFound)
		})

		Convey("GET, malformed id_token_hint: treated as no hint, confirmation page, not an error", func() {
			idTokenVerifier.EXPECT().VerifyIDToken("garbage").Return(nil, errors.New("bad signature"))

			urls.EXPECT().LogoutURL(gomock.Any()).DoAndReturn(func(u *url.URL) *url.URL {
				result, _ := url.Parse("https://app.example.com/logout?redirect_uri=" + url.QueryEscape(u.String()))
				return result
			})

			req := protocol.EndSessionRequest{
				"id_token_hint":            "garbage",
				"post_logout_redirect_uri": "https://rp.example.com/after-logout",
			}
			r := httptest.NewRequest(http.MethodGet, "https://app.example.com/oauth2/end_session", nil)
			rw := httptest.NewRecorder()

			err := h.Handle(context.Background(), sess, req, r, rw)
			So(err, ShouldBeNil)
			So(rw.Code, ShouldEqual, http.StatusFound)
		})

		Convey("GET, no session at all: straight to post_logout_redirect_uri, no confirmation, no error", func() {
			req := protocol.EndSessionRequest{
				"post_logout_redirect_uri": "https://rp.example.com/after-logout",
			}
			r := httptest.NewRequest(http.MethodGet, "https://app.example.com/oauth2/end_session", nil)
			rw := httptest.NewRecorder()

			err := h.Handle(context.Background(), nil, req, r, rw)
			So(err, ShouldBeNil)
			So(rw.Code, ShouldEqual, http.StatusSeeOther)

			loc, parseErr := url.Parse(rw.Header().Get("Location"))
			So(parseErr, ShouldBeNil)
			So(loc.Scheme+"://"+loc.Host+loc.Path, ShouldEqual, "https://rp.example.com/after-logout")
		})

		Convey("POST, valid id_token_hint bound to an offline grant in the same SSO group: stash round trip then silent logout", func() {
			idTokenVerifier.EXPECT().VerifyIDToken("valid-hint").Return(newIDToken(sessOfflineGrantSID, firstPartyClient.ClientID), nil)
			expectSameLoginOfflineGrant()
			sessionManager.EXPECT().Logout(gomock.Any(), sess, gomock.Any()).Return(nil, nil)

			postReq := protocol.EndSessionRequest{
				"id_token_hint":            "valid-hint",
				"post_logout_redirect_uri": "https://rp.example.com/after-logout",
			}
			rPost := httptest.NewRequest(http.MethodPost, "https://app.example.com/oauth2/end_session", nil)
			rwPost := httptest.NewRecorder()

			err := h.Handle(context.Background(), nil, postReq, rPost, rwPost)
			So(err, ShouldBeNil)
			So(rwPost.Code, ShouldEqual, http.StatusFound)

			location := rwPost.Header().Get("Location")
			So(strings.Contains(location, "id_token_hint"), ShouldBeFalse)

			locURL, parseErr := url.Parse(location)
			So(parseErr, ShouldBeNil)
			ref := locURL.Query().Get(endSessionRefQueryParam)
			So(ref, ShouldNotBeEmpty)

			resumedReq := protocol.EndSessionRequest{endSessionRefQueryParam: ref}
			rResumed := httptest.NewRequest(http.MethodGet, location, nil)
			rwResumed := httptest.NewRecorder()

			err = h.Handle(context.Background(), sess, resumedReq, rResumed, rwResumed)
			So(err, ShouldBeNil)
			So(rwResumed.Code, ShouldEqual, http.StatusSeeOther)

			loc, parseErr := url.Parse(rwResumed.Header().Get("Location"))
			So(parseErr, ShouldBeNil)
			So(loc.Scheme+"://"+loc.Host+loc.Path, ShouldEqual, "https://rp.example.com/after-logout")

			// The stash cookie must not survive being consumed: a resumed
			// request always clears it (see resumeFromStash's defer), whether
			// or not the stash opened successfully, so a stale
			// x_end_session_ref revisited later (e.g. from browser history)
			// can never be replayed against a live cookie.
			assertStashCookieCleared(cookies, rwResumed)
		})

		Convey("POST with no id_token_hint at all, session present: stash round trip still runs, ends at confirmation page", func() {
			var capturedEndSessionURL *url.URL
			urls.EXPECT().LogoutURL(gomock.Any()).DoAndReturn(func(u *url.URL) *url.URL {
				capturedEndSessionURL = u
				result, _ := url.Parse("https://app.example.com/logout?redirect_uri=" + url.QueryEscape(u.String()))
				return result
			})

			postReq := protocol.EndSessionRequest{
				"post_logout_redirect_uri": "https://rp.example.com/after-logout",
			}
			rPost := httptest.NewRequest(http.MethodPost, "https://app.example.com/oauth2/end_session", nil)
			rwPost := httptest.NewRecorder()

			err := h.Handle(context.Background(), nil, postReq, rPost, rwPost)
			So(err, ShouldBeNil)
			So(rwPost.Code, ShouldEqual, http.StatusFound)

			location := rwPost.Header().Get("Location")
			locURL, parseErr := url.Parse(location)
			So(parseErr, ShouldBeNil)
			ref := locURL.Query().Get(endSessionRefQueryParam)
			So(ref, ShouldNotBeEmpty)

			resumedReq := protocol.EndSessionRequest{endSessionRefQueryParam: ref}
			rResumed := httptest.NewRequest(http.MethodGet, location, nil)
			rwResumed := httptest.NewRecorder()

			err = h.Handle(context.Background(), sess, resumedReq, rResumed, rwResumed)
			So(err, ShouldBeNil)
			So(rwResumed.Code, ShouldEqual, http.StatusFound)
			So(capturedEndSessionURL, ShouldNotBeNil)
		})

		Convey("Resumed GET with x_end_session_ref set but the stash cookie missing: falls back to a no-parameter request, confirmation page shown", func() {
			postReq := protocol.EndSessionRequest{
				"id_token_hint":            "valid-hint",
				"post_logout_redirect_uri": "https://rp.example.com/after-logout",
			}
			rPost := httptest.NewRequest(http.MethodPost, "https://app.example.com/oauth2/end_session", nil)
			rwPost := httptest.NewRecorder()

			err := h.Handle(context.Background(), nil, postReq, rPost, rwPost)
			So(err, ShouldBeNil)

			location := rwPost.Header().Get("Location")
			locURL, parseErr := url.Parse(location)
			So(parseErr, ShouldBeNil)
			ref := locURL.Query().Get(endSessionRefQueryParam)
			So(ref, ShouldNotBeEmpty)

			// A different handler instance with a fresh (empty) cookie
			// manager simulates the stash cookie never having arrived
			// (expired, blocked, or a different browser/tab).
			cookiesMissing := newFakeCookieManager()
			hMissingCookie := newHandler(cookiesMissing, sessionManager, urls, idTokenVerifier, sessions, offlineGrants)

			var capturedEndSessionURL *url.URL
			urls.EXPECT().LogoutURL(gomock.Any()).DoAndReturn(func(u *url.URL) *url.URL {
				capturedEndSessionURL = u
				result, _ := url.Parse("https://app.example.com/logout?redirect_uri=" + url.QueryEscape(u.String()))
				return result
			})

			resumedReq := protocol.EndSessionRequest{endSessionRefQueryParam: ref}
			rResumed := httptest.NewRequest(http.MethodGet, location, nil)
			rwResumed := httptest.NewRecorder()

			err = hMissingCookie.Handle(context.Background(), sess, resumedReq, rResumed, rwResumed)
			So(err, ShouldBeNil)
			So(rwResumed.Code, ShouldEqual, http.StatusFound)
			So(capturedEndSessionURL, ShouldNotBeNil)
			// The unopenable stash's id_token_hint/post_logout_redirect_uri
			// must not leak through as if they were genuinely the caller's.
			So(capturedEndSessionURL.Query().Get("id_token_hint"), ShouldEqual, "")
			So(capturedEndSessionURL.Query().Get("post_logout_redirect_uri"), ShouldEqual, "")
			// resumeFromStash's defer clears the stash cookie unconditionally,
			// even on this graceful-fallback path.
			assertStashCookieCleared(cookiesMissing, rwResumed)
		})

		Convey("Resumed GET with x_end_session_ref set and a stash cookie present but not matching: falls back to a no-parameter request, confirmation page shown", func() {
			postReq1 := protocol.EndSessionRequest{
				"id_token_hint":            "valid-hint",
				"post_logout_redirect_uri": "https://rp.example.com/after-logout",
			}
			rPost1 := httptest.NewRequest(http.MethodPost, "https://app.example.com/oauth2/end_session", nil)
			rwPost1 := httptest.NewRecorder()
			err := h.Handle(context.Background(), nil, postReq1, rPost1, rwPost1)
			So(err, ShouldBeNil)

			// A second, independent POST produces a different sealed value
			// under a different key.
			cookies2 := newFakeCookieManager()
			h2 := newHandler(cookies2, sessionManager, urls, idTokenVerifier, sessions, offlineGrants)
			postReq2 := protocol.EndSessionRequest{
				"id_token_hint":            "valid-hint",
				"post_logout_redirect_uri": "https://rp.example.com/after-logout",
			}
			rPost2 := httptest.NewRequest(http.MethodPost, "https://app.example.com/oauth2/end_session", nil)
			rwPost2 := httptest.NewRecorder()
			err = h2.Handle(context.Background(), nil, postReq2, rPost2, rwPost2)
			So(err, ShouldBeNil)

			locURL2, parseErr := url.Parse(rwPost2.Header().Get("Location"))
			So(parseErr, ShouldBeNil)
			mismatchedRef := locURL2.Query().Get(endSessionRefQueryParam)
			So(mismatchedRef, ShouldNotBeEmpty)

			var capturedEndSessionURL *url.URL
			urls.EXPECT().LogoutURL(gomock.Any()).DoAndReturn(func(u *url.URL) *url.URL {
				capturedEndSessionURL = u
				result, _ := url.Parse("https://app.example.com/logout?redirect_uri=" + url.QueryEscape(u.String()))
				return result
			})

			// Use the first handler's cookie manager (holding the first
			// POST's key) but the second POST's sealed ref: the pairing
			// must not authenticate.
			resumedReq := protocol.EndSessionRequest{endSessionRefQueryParam: mismatchedRef}
			rResumed := httptest.NewRequest(http.MethodGet, "https://app.example.com/oauth2/end_session", nil)
			rwResumed := httptest.NewRecorder()

			err = h.Handle(context.Background(), sess, resumedReq, rResumed, rwResumed)
			So(err, ShouldBeNil)
			So(rwResumed.Code, ShouldEqual, http.StatusFound)
			So(capturedEndSessionURL, ShouldNotBeNil)
			So(capturedEndSessionURL.Query().Get("id_token_hint"), ShouldEqual, "")
			So(capturedEndSessionURL.Query().Get("post_logout_redirect_uri"), ShouldEqual, "")
			// h's own stash cookie (from postReq1) must still be cleared even
			// though the ref presented here (mismatchedRef) belonged to h2:
			// resumeFromStash clears whatever stash cookie the request it's
			// given actually carries, regardless of why opening it failed.
			assertStashCookieCleared(cookies, rwResumed)
		})
	})
}

// assertStashCookieCleared checks that a resumed request's response actually
// clears EndSessionRefKeyCookieDef: both server-side (removed from the fake
// cookie manager's backing store, so it can't be read again) and via the
// Set-Cookie header a real browser would need to see (negative MaxAge).
// Without this, a stale x_end_session_ref revisited later (e.g. from browser
// history) could still pair with a cookie that was never actually deleted.
func assertStashCookieCleared(cookies *fakeCookieManager, rw *httptest.ResponseRecorder) {
	_, stillPresent := cookies.values[handler.EndSessionRefKeyCookieDef.NameSuffix]
	So(stillPresent, ShouldBeFalse)

	var cleared *http.Cookie
	for _, c := range rw.Result().Cookies() {
		if c.Name == handler.EndSessionRefKeyCookieDef.NameSuffix {
			cleared = c
			break
		}
	}
	So(cleared, ShouldNotBeNil)
	So(cleared.MaxAge, ShouldBeLessThan, 0)
}
