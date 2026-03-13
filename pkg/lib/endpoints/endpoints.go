package endpoints

import (
	"fmt"
	"net/url"
	"path"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/urlutil"
)

type OAuthEndpoints struct {
	HTTPHost               httputil.HTTPHost
	HTTPProto              httputil.HTTPProto
	SharedAuthgearEndpoint config.SharedAuthgearEndpoint
}

type EndpointsUIImplementationService interface {
	GetUIImplementation() config.UIImplementation
}

type Endpoints struct {
	*OAuthEndpoints
	UIImplementationService EndpointsUIImplementationService
}

func (e *OAuthEndpoints) Origin() *url.URL {
	return &url.URL{
		Host:   string(e.HTTPHost),
		Scheme: string(e.HTTPProto),
	}
}

func (e *OAuthEndpoints) urlOf(relPath string) *url.URL {
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

func (e *OAuthEndpoints) ssoCallbackEndpointURL() *url.URL { return e.urlOf("sso/oauth2/callback") }
func (e *OAuthEndpoints) SSOCallbackURL(alias string) *url.URL {
	u := e.ssoCallbackEndpointURL()
	u.Path = path.Join(u.Path, url.PathEscape(alias))
	return u
}
func (e *OAuthEndpoints) SharedSSOCallbackURL() *url.URL {
	u, err := url.Parse(string(e.SharedAuthgearEndpoint))
	if err != nil {
		panic(fmt.Errorf("SHARED_AUTHGEAR_ENDPOINT is not a valid uri: %w", err))
	}
	u.Path = path.Join(u.Path, "/noproject/sso/oauth2/callback")
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
	default:
		panic(fmt.Errorf("unexpected ui implementation %s", uiImpl))
	}
}
func (e *Endpoints) ErrorEndpointURL() *url.URL {
	uiImpl := e.UIImplementationService.GetUIImplementation()
	switch uiImpl {
	case config.UIImplementationAuthflowV2:
		return e.urlOf("/v2/errors/error")
	default:
		panic(fmt.Errorf("unexpected ui implementation %s", uiImpl))
	}
}
func (e *Endpoints) SelectAccountEndpointURL() *url.URL {
	uiImpl := e.UIImplementationService.GetUIImplementation()
	switch uiImpl {
	case config.UIImplementationAuthflowV2:
		return e.urlOf("/authflow/v2/select_account")
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

func (e *Endpoints) WeChatAuthorizeEndpointURL() *url.URL { return e.urlOf("sso/wechat/auth") }
func (e *Endpoints) WeChatCallbackEndpointURL() *url.URL {
	return e.urlOf("sso/wechat/callback")
}

func (e *Endpoints) LoginLinkVerificationEndpointURL() *url.URL {
	uiImpl := e.UIImplementationService.GetUIImplementation()
	switch uiImpl {
	case config.UIImplementationAuthflowV2:
		return e.urlOf("/authflow/v2/verify_login_link")
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

func (e *Endpoints) SettingsAddLoginIDEmail(loginIDKey string) *url.URL {
	u := e.urlOf("settings/identity/add_email")
	q := u.Query()
	q.Set("q_login_id_key", loginIDKey)
	u.RawQuery = q.Encode()
	return u
}

func (e *Endpoints) SettingsAddLoginIDPhone(loginIDKey string) *url.URL {
	u := e.urlOf("settings/identity/add_phone")
	q := u.Query()
	q.Set("q_login_id_key", loginIDKey)
	u.RawQuery = q.Encode()
	return u
}

func (e *Endpoints) SettingsAddLoginIDUsername(loginIDKey string) *url.URL {
	u := e.urlOf("settings/identity/add_username")
	q := u.Query()
	q.Set("q_login_id_key", loginIDKey)
	u.RawQuery = q.Encode()
	return u
}

func (e *Endpoints) SettingsEditLoginIDEmail(loginIDKey string) *url.URL {
	u := e.urlOf("/settings/identity/change_email")
	q := u.Query()
	q.Set("q_login_id_key", loginIDKey)
	u.RawQuery = q.Encode()
	return u
}

func (e *Endpoints) SettingsEditLoginIDPhone(loginIDKey string) *url.URL {
	u := e.urlOf("settings/identity/change_phone")
	q := u.Query()
	q.Set("q_login_id_key", loginIDKey)
	u.RawQuery = q.Encode()
	return u
}

func (e *Endpoints) SettingsEditLoginIDUsername(loginIDKey string) *url.URL {
	u := e.urlOf("settings/identity/change_username")
	q := u.Query()
	q.Set("q_login_id_key", loginIDKey)
	u.RawQuery = q.Encode()
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
func (e *Endpoints) SAMLLogoutURL(serviceProviderId string) *url.URL {
	return e.urlOf(fmt.Sprintf("saml2/logout/%s", serviceProviderId))
}
