package handler_test

import (
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oauthsession"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type mockURLsProvider struct{}

func (mockURLsProvider) AuthorizeURL(r protocol.AuthorizationRequest) *url.URL {
	u, _ := url.Parse("https://auth/authorize")
	return u
}

func (mockURLsProvider) ConsentURL(r protocol.AuthorizationRequest) *url.URL {
	u, _ := url.Parse("https://auth/consent")
	return u
}

func (mockURLsProvider) AuthenticateURL(opts webapp.AuthenticateURLOptions) (httputil.Result, error) {
	return &httputil.ResultRedirect{URL: "https://auth/authenticate"}, nil
}

type mockCodeGrantStore struct {
	grants []oauth.CodeGrant
}

func (m *mockCodeGrantStore) GetCodeGrant(codeHash string) (*oauth.CodeGrant, error) {
	for _, g := range m.grants {
		if g.CodeHash == codeHash {
			return &g, nil
		}
	}
	return nil, oauth.ErrGrantNotFound
}

func (m *mockCodeGrantStore) CreateCodeGrant(grant *oauth.CodeGrant) error {
	m.grants = append(m.grants, *grant)
	return nil
}

func (m *mockCodeGrantStore) DeleteCodeGrant(grant *oauth.CodeGrant) error {
	n := 0
	for _, g := range m.grants {
		if g.CodeHash != grant.CodeHash {
			m.grants[n] = g
			n++
		}
	}
	m.grants = m.grants[:n]
	return nil
}

type mockAuthenticationInfoService struct {
	Entry *authenticationinfo.Entry
}

func (m *mockAuthenticationInfoService) Get(entryID string) (*authenticationinfo.Entry, error) {
	if m.Entry == nil {
		return nil, authenticationinfo.ErrNotFound
	}

	return m.Entry, nil
}

func (m *mockAuthenticationInfoService) Delete(entryID string) error {
	m.Entry = nil
	return nil
}

type mockOAuthSessionService struct {
	Entry *oauthsession.Entry
}

func (m *mockOAuthSessionService) Save(entry *oauthsession.Entry) (err error) {
	m.Entry = entry
	return nil
}

func (m *mockOAuthSessionService) Get(entryID string) (*oauthsession.Entry, error) {
	if m.Entry == nil {
		return nil, oauthsession.ErrNotFound
	}

	return m.Entry, nil
}

func (m *mockOAuthSessionService) Delete(entryID string) error {
	m.Entry = nil
	return nil
}

type mockCookieManager struct{}

func (m *mockCookieManager) GetCookie(r *http.Request, def *httputil.CookieDef) (*http.Cookie, error) {
	return &http.Cookie{}, nil
}

func (m *mockCookieManager) ClearCookie(def *httputil.CookieDef) *http.Cookie {
	return &http.Cookie{}
}

func (m *mockCookieManager) ValueCookie(def *httputil.CookieDef, value string) *http.Cookie {
	return &http.Cookie{}
}
