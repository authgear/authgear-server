package endpoints

import (
	"fmt"
	"net/url"
	"path"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/urlutil"
)

type EndpointsUIImplementationService interface {
	GetUIImplementation() config.UIImplementation
}

type Endpoints struct {
	HTTPHost                httputil.HTTPHost
	HTTPProto               httputil.HTTPProto
	UIImplementationService EndpointsUIImplementationService
}

func (e *Endpoints) Origin() *url.URL {
	return &url.URL{
		Host:   string(e.HTTPHost),
		Scheme: string(e.HTTPProto),
	}
}

func (e *Endpoints) urlOf(relPath string) *url.URL {
	// If we do not set Path = "/", then in urlOf,
	// Path will have no leading /.
	// It is problematic when Path is used in comparison.
	//
	// u, _ := url.Parse("https://example.com/path")
	// // u.Path is "/path"
	// uu := endpoints.urlOf("path")
	// // uu.Path is "path"
	// So direct comparison will yield a surprising result.
	// More confusing is that u.String() == uu.String()
	// Because String() will add leading / to make the URL legal.
	u := e.Origin()
	u.Path = path.Join("/", relPath)
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
	uiImpl := e.UIImplementationService.GetUIImplementation()
	switch uiImpl {
	case config.UIImplementationAuthflowV2:
		return e.urlOf("authflow/v2/reset_password")
	case config.UIImplementationAuthflow:
		return e.urlOf("authflow/reset_password")
	case config.UIImplementationInteraction:
		return e.urlOf("flows/reset_password")
	default:
		panic(fmt.Errorf("unexpected ui implementation %s", uiImpl))
	}
}
func (e *Endpoints) ErrorEndpointURL() *url.URL {
	uiImpl := e.UIImplementationService.GetUIImplementation()
	switch uiImpl {
	case config.UIImplementationAuthflowV2:
		return e.urlOf("/v2/errors/error")
	case config.UIImplementationInteraction:
		fallthrough
	case config.UIImplementationAuthflow:
		return e.urlOf("/errors/error")
	default:
		panic(fmt.Errorf("unexpected ui implementation %s", uiImpl))
	}
}
func (e *Endpoints) SelectAccountEndpointURL() *url.URL {
	uiImpl := e.UIImplementationService.GetUIImplementation()
	switch uiImpl {
	case config.UIImplementationAuthflowV2:
		return e.urlOf("/authflow/v2/select_account")
	case config.UIImplementationInteraction:
		fallthrough
	case config.UIImplementationAuthflow:
		return e.urlOf("/flows/select_account")
	default:
		panic(fmt.Errorf("unexpected ui implementation %s", uiImpl))
	}
}
func (e *Endpoints) VerifyBotProtectionEndpointURL() *url.URL {
	uiImpl := e.UIImplementationService.GetUIImplementation()
	switch uiImpl {
	case config.UIImplementationAuthflowV2:
		return e.urlOf("/authflow/v2/verify_bot_protection")
	default:
		panic(fmt.Errorf("unexpected ui implementation %s", uiImpl))
	}
}
func (e *Endpoints) SSOCallbackEndpointURL() *url.URL { return e.urlOf("sso/oauth2/callback") }

func (e *Endpoints) WeChatAuthorizeEndpointURL() *url.URL { return e.urlOf("sso/wechat/auth") }
func (e *Endpoints) WeChatCallbackEndpointURL() *url.URL {
	return e.urlOf("sso/wechat/callback")
}

func (e *Endpoints) LoginLinkVerificationEndpointURL() *url.URL {
	uiImpl := e.UIImplementationService.GetUIImplementation()
	switch uiImpl {
	case config.UIImplementationAuthflowV2:
		return e.urlOf("/authflow/v2/verify_login_link")
	case config.UIImplementationInteraction:
		fallthrough
	case config.UIImplementationAuthflow:
		return e.urlOf("flows/verify_login_link")
	default:
		panic(fmt.Errorf("unexpected ui implementation %s", uiImpl))
	}
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
func (e *Endpoints) SettingsChangePasswordURL() *url.URL {
	return e.urlOf("settings/change_password")
}

func (e *Endpoints) SettingsDeleteAccountURL() *url.URL {
	return e.urlOf("settings/delete_account")
}

func (e *Endpoints) SSOCallbackURL(alias string) *url.URL {
	u := e.SSOCallbackEndpointURL()
	u.Path = path.Join(u.Path, url.PathEscape(alias))
	return u
}

func (e *Endpoints) WeChatAuthorizeURL(alias string) *url.URL {
	u := e.WeChatAuthorizeEndpointURL()
	u.Path = path.Join(u.Path, url.PathEscape(alias))
	return u
}

func (e *Endpoints) TesterURL() *url.URL { return e.urlOf("tester") }

func (e *Endpoints) SAMLLoginURL(serviceProviderId string) *url.URL {
	return e.urlOf(fmt.Sprintf("saml2/login/%s", serviceProviderId))
}
func (e *Endpoints) SAMLLoginFinishURL() *url.URL {
	return e.urlOf(fmt.Sprintf("saml2/login_finish"))
}
