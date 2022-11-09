package webapp

import (
	"net/http"
	"net/url"
	"path"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/urlutil"
)

type EndpointsProvider interface {
	BaseURL() *url.URL
	OAuthEntrypointURL(clientID string, redirectURI string) *url.URL
	LoginEndpointURL() *url.URL
	SignupEndpointURL() *url.URL
	LogoutEndpointURL() *url.URL
	SettingsEndpointURL() *url.URL
	ResetPasswordEndpointURL() *url.URL
	SSOCallbackEndpointURL() *url.URL
	WeChatAuthorizeEndpointURL() *url.URL
	WeChatCallbackEndpointURL() *url.URL
}

type URLProvider struct {
	Endpoints EndpointsProvider
}

func (p *URLProvider) LogoutURL(redirectURI *url.URL) *url.URL {
	return urlutil.WithQueryParamsAdded(
		p.Endpoints.LogoutEndpointURL(),
		map[string]string{"redirect_uri": redirectURI.String()},
	)
}

func (p *URLProvider) SettingsURL() *url.URL {
	return p.Endpoints.SettingsEndpointURL()
}

func (p *URLProvider) ResetPasswordURL(code string) *url.URL {
	return urlutil.WithQueryParamsAdded(
		p.Endpoints.ResetPasswordEndpointURL(),
		map[string]string{"code": code},
	)
}

func (p *URLProvider) SSOCallbackURL(c config.OAuthSSOProviderConfig) *url.URL {
	u := p.Endpoints.SSOCallbackEndpointURL()
	u.Path = path.Join(u.Path, url.PathEscape(c.Alias))
	return u
}

type AuthenticateURLPageService interface {
	CreateSession(session *Session, redirectURI string) (*Result, error)
}

type AuthenticateURLProvider struct {
	Endpoints EndpointsProvider
	Pages     AuthenticateURLPageService
	Clock     clock.Clock
}

type AuthenticateURLOptions struct {
	SessionOptions SessionOptions
	UILocales      string
	ColorScheme    string
	Cookies        []*http.Cookie
	ClientID       string
	// RedirectURL will be used only when the WebSession doesn't have the redirect URI
	// When the WebSession has a redirect URI, it usually starts from the authorization endpoint
	// User will be redirected back to the authorization endpoint after authentication
	// Authorization endpoint will use the redirect URI in the OAuthSession
	RedirectURL string
}

func (p *AuthenticateURLProvider) AuthenticateURL(options AuthenticateURLOptions) (httputil.Result, error) {
	endpoint := p.Endpoints.OAuthEntrypointURL(options.ClientID, options.RedirectURL).String()
	now := p.Clock.NowUTC()
	sessionOpts := options.SessionOptions
	sessionOpts.UpdatedAt = now
	session := NewSession(sessionOpts)
	result, err := p.Pages.CreateSession(session, endpoint)
	if err != nil {
		return nil, err
	}

	if result != nil {
		result.UILocales = options.UILocales
		result.ColorScheme = options.ColorScheme
	}

	if options.Cookies != nil {
		result.Cookies = append(result.Cookies, options.Cookies...)
	}
	return result, nil
}
