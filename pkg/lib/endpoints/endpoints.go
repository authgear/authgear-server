package endpoints

import (
	"net/url"
	"path"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/urlutil"
)

type Endpoints struct {
	HTTPHost  httputil.HTTPHost
	HTTPProto httputil.HTTPProto
}

func (e *Endpoints) Origin() *url.URL {
	return &url.URL{
		Host:   string(e.HTTPHost),
		Scheme: string(e.HTTPProto),
	}
}

func (e *Endpoints) BaseURL() *url.URL {
	return e.Origin()
}

func (e *Endpoints) urlOf(relPath string) *url.URL {
	u := e.BaseURL()
	u.Path = path.Join(u.Path, relPath)
	return u
}

func (e *Endpoints) AuthorizeEndpointURL() *url.URL  { return e.urlOf("oauth2/authorize") }
func (e *Endpoints) ConsentEndpointURL() *url.URL    { return e.urlOf("oauth2/consent") }
func (e *Endpoints) TokenEndpointURL() *url.URL      { return e.urlOf("oauth2/token") }
func (e *Endpoints) RevokeEndpointURL() *url.URL     { return e.urlOf("oauth2/revoke") }
func (e *Endpoints) JWKSEndpointURL() *url.URL       { return e.urlOf("oauth2/jwks") }
func (e *Endpoints) UserInfoEndpointURL() *url.URL   { return e.urlOf("oauth2/userinfo") }
func (e *Endpoints) EndSessionEndpointURL() *url.URL { return e.urlOf("oauth2/end_session") }
func (e *Endpoints) OAuthEntrypointURL() *url.URL {
	return e.urlOf("_internals/oauth_entrypoint")
}
func (e *Endpoints) LoginEndpointURL() *url.URL       { return e.urlOf("./login") }
func (e *Endpoints) SignupEndpointURL() *url.URL      { return e.urlOf("./signup") }
func (e *Endpoints) PromoteUserEndpointURL() *url.URL { return e.urlOf("flows/promote_user") }
func (e *Endpoints) LogoutEndpointURL() *url.URL      { return e.urlOf("./logout") }
func (e *Endpoints) SettingsEndpointURL() *url.URL    { return e.urlOf("./settings") }
func (e *Endpoints) ResetPasswordEndpointURL() *url.URL {
	return e.urlOf("flows/reset_password")
}
func (e *Endpoints) SSOCallbackEndpointURL() *url.URL { return e.urlOf("sso/oauth2/callback") }

func (e *Endpoints) WeChatAuthorizeEndpointURL() *url.URL { return e.urlOf("sso/wechat/auth") }
func (e *Endpoints) WeChatCallbackEndpointURL() *url.URL {
	return e.urlOf("sso/wechat/callback")
}

func (e *Endpoints) LoginLinkVerificationEndpointURL() *url.URL {
	return e.urlOf("flows/verify_login_link")
}

func (e *Endpoints) LogoutURL(redirectURI *url.URL) *url.URL {
	return urlutil.WithQueryParamsAdded(
		e.LogoutEndpointURL(),
		map[string]string{"redirect_uri": redirectURI.String()},
	)
}

func (e *Endpoints) SettingsURL() *url.URL {
	return e.SettingsEndpointURL()
}

func (e *Endpoints) ResetPasswordURL(code string) *url.URL {
	return urlutil.WithQueryParamsAdded(
		e.ResetPasswordEndpointURL(),
		map[string]string{"code": code},
	)
}

func (e *Endpoints) SSOCallbackURL(c config.OAuthSSOProviderConfig) *url.URL {
	u := e.SSOCallbackEndpointURL()
	u.Path = path.Join(u.Path, url.PathEscape(c.Alias))
	return u
}

func (e *Endpoints) WeChatAuthorizeURL(c config.OAuthSSOProviderConfig) *url.URL {
	u := e.WeChatAuthorizeEndpointURL()
	u.Path = path.Join(u.Path, url.PathEscape(c.Alias))
	return u
}
