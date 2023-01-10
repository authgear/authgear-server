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
	OAuthEntrypointURL() *url.URL
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
	Client         *config.OAuthClientConfig
	// RedirectURL will be used only when the WebSession doesn't have the redirect URI
	// When the WebSession has a redirect URI, it usually starts from the authorization endpoint
	// User will be redirected back to the authorization endpoint after authentication
	// Authorization endpoint will use the redirect URI in the OAuthSession
	RedirectURL   string
	CustomUIQuery string
}

func (p *AuthenticateURLProvider) AuthenticateURL(options AuthenticateURLOptions) (httputil.Result, error) {
	var endpoint *url.URL
	if options.Client != nil && options.Client.CustomUIURI != "" {
		var err error
		endpoint, err = p.customUIURL(options.Client.CustomUIURI, options.CustomUIQuery)
		if err != nil {
			return nil, ErrInvalidCustomURI.Errorf("invalid custom ui uri: %w", err)
		}
	} else {
		endpoint = p.Endpoints.OAuthEntrypointURL()
	}

	// Assign client id and redirect url to the endpoint
	q := endpoint.Query()
	if options.Client != nil {
		q.Set("client_id", options.Client.ClientID)
	}
	if options.RedirectURL != "" {
		q.Set("redirect_uri", options.RedirectURL)
	}
	endpoint.RawQuery = q.Encode()

	now := p.Clock.NowUTC()
	sessionOpts := options.SessionOptions
	sessionOpts.UpdatedAt = now
	session := NewSession(sessionOpts)
	result, err := p.Pages.CreateSession(session, endpoint.String())
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

func (p *AuthenticateURLProvider) customUIURL(customUIURI string, customUIQuery string) (*url.URL, error) {
	customUIURL, err := url.Parse(customUIURI)
	if err != nil {
		return nil, err
	}

	q := customUIURL.Query()

	// Assign query from the SDK to the url
	queryFromSDK, err := url.ParseQuery(customUIQuery)
	if err != nil {
		return nil, err
	}
	for key, values := range queryFromSDK {
		for idx, val := range values {
			if idx == 0 {
				q.Set(key, val)
			} else {
				q.Add(key, val)
			}
		}
	}
	customUIURL.RawQuery = q.Encode()

	return customUIURL, nil
}
