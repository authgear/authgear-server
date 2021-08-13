package webapp

import (
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
	VerifyIdentityEndpointURL() *url.URL
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

func (p *URLProvider) VerifyIdentityURL(code string, id string) *url.URL {
	return urlutil.WithQueryParamsAdded(
		p.Endpoints.VerifyIdentityEndpointURL(),
		map[string]string{"code": code, "id": id},
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
}

func (p *AuthenticateURLProvider) AuthenticateURL(options AuthenticateURLOptions) (httputil.Result, error) {
	endpoint := p.Endpoints.OAuthEntrypointURL().String()
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
	}
	return result, nil
}
