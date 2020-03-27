package authn

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/loginid"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/mfa"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/session"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/authn"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/handler"
)

type ProviderFactory struct {
	OAuth                   *OAuthCoordinator
	Authn                   *AuthenticateProcess
	Signup                  *SignupProcess
	AuthnSession            *SessionProvider
	Session                 session.Provider
	SessionCookieConfig     session.CookieConfiguration
	BearerTokenCookieConfig mfa.BearerTokenCookieConfiguration
}

func (f *ProviderFactory) makeProvider(forAuthAPI bool) *Provider {
	return &Provider{
		ForAuthAPI:              forAuthAPI,
		OAuth:                   f.OAuth,
		Authn:                   f.Authn,
		Signup:                  f.Signup,
		AuthnSession:            f.AuthnSession,
		Session:                 f.Session,
		SessionCookieConfig:     f.SessionCookieConfig,
		BearerTokenCookieConfig: f.BearerTokenCookieConfig,
	}
}

func (f *ProviderFactory) ForAuthUI() *Provider  { return f.makeProvider(false) }
func (f *ProviderFactory) ForAuthAPI() *Provider { return f.makeProvider(true) }

type Provider struct {
	ForAuthAPI              bool
	OAuth                   *OAuthCoordinator
	Authn                   *AuthenticateProcess
	Signup                  *SignupProcess
	AuthnSession            *SessionProvider
	Session                 session.Provider
	SessionCookieConfig     session.CookieConfiguration
	BearerTokenCookieConfig mfa.BearerTokenCookieConfiguration
}

func (p *Provider) SignupWithLoginIDs(
	client config.OAuthClientConfiguration,
	loginIDs []loginid.LoginID,
	plainPassword string,
	metadata map[string]interface{},
	onUserDuplicate model.OnUserDuplicate,
) (Result, error) {
	pr, err := p.Signup.SignupWithLoginIDs(loginIDs, plainPassword, metadata, onUserDuplicate)
	if err != nil {
		return nil, err
	}

	s, err := p.AuthnSession.BeginSession(client, pr.PrincipalUserID(), pr, auth.SessionCreateReasonSignup)
	if err != nil {
		return nil, err
	}
	s.ForAuthAPI = p.ForAuthAPI

	return p.AuthnSession.StepSession(s)
}

func (p *Provider) ValidateSignupLoginID(loginID loginid.LoginID) error {
	return p.Signup.ValidateSignupLoginID(loginID)
}

func (p *Provider) LoginWithLoginID(
	client config.OAuthClientConfiguration,
	loginID loginid.LoginID,
	plainPassword string,
) (Result, error) {
	pr, err := p.Authn.AuthenticateWithLoginID(loginID, plainPassword)
	if err != nil {
		return nil, err
	}

	s, err := p.AuthnSession.BeginSession(client, pr.PrincipalUserID(), pr, auth.SessionCreateReasonLogin)
	if err != nil {
		return nil, err
	}
	s.ForAuthAPI = p.ForAuthAPI

	return p.AuthnSession.StepSession(s)
}

func (p *Provider) OAuthAuthenticate(
	authInfo sso.AuthInfo,
	codeChallenge string,
	loginState sso.LoginState,
) (*sso.SkygearAuthorizationCode, error) {
	return p.OAuth.Authenticate(authInfo, codeChallenge, loginState)
}

func (p *Provider) OAuthLink(
	authInfo sso.AuthInfo,
	codeChallenge string,
	linkState sso.LinkState,
) (*sso.SkygearAuthorizationCode, error) {
	return p.OAuth.Link(authInfo, codeChallenge, linkState)
}

func (p *Provider) OAuthExchangeCode(
	client config.OAuthClientConfiguration,
	s auth.AuthSession,
	code *sso.SkygearAuthorizationCode,
) (Result, error) {
	pr, err := p.OAuth.ExchangeCode(code)
	if err != nil {
		return nil, err
	}

	if code.Action == "link" {
		if s == nil {
			return nil, authz.ErrNotAuthenticated
		}
		return p.AuthnSession.MakeResult(client, s, "")
	}

	// code.Action == "login"
	reason := auth.SessionCreateReason(code.SessionCreateReason)
	as, err := p.AuthnSession.BeginSession(client, pr.PrincipalUserID(), pr, reason)
	if err != nil {
		return nil, err
	}
	as.ForAuthAPI = p.ForAuthAPI

	return p.AuthnSession.StepSession(as)
}

func (p *Provider) WriteCookie(rw http.ResponseWriter, result *CompletionResult) {
	if result.UseCookie() {
		if result.SessionToken != "" {
			p.SessionCookieConfig.WriteTo(rw, result.SessionToken)
		}
		if result.MFABearerToken != "" {
			p.BearerTokenCookieConfig.WriteTo(rw, result.MFABearerToken)
		}
	}
}

func (p *Provider) MakeAPIBody(rw http.ResponseWriter, result *CompletionResult) (resp model.AuthResponse) {
	resp.User = *result.User
	resp.Identity = result.Principal
	if result.Session != nil {
		resp.SessionID = result.Session.ID
	}
	if result.MFABearerToken != "" && !result.UseCookie() {
		resp.MFABearerToken = result.MFABearerToken
	}
	resp.AccessToken = result.AccessToken
	resp.RefreshToken = result.RefreshToken
	resp.ExpiresIn = result.ExpiresIn
	return
}

func (p *Provider) WriteAPIResult(rw http.ResponseWriter, result Result) {
	switch r := result.(type) {
	case *CompletionResult:
		p.WriteCookie(rw, r)
		resp := p.MakeAPIBody(rw, r)
		handler.WriteResponse(rw, handler.APIResponse{Result: resp})
	case *InProgressResult:
		handler.WriteResponse(rw, handler.APIResponse{Error: r.ToAPIError()})
	}
}

func (p *Provider) Resolve(
	client config.OAuthClientConfiguration,
	authnSessionToken string,
	stepPredicate func(SessionStep) bool,
) (*AuthnSession, error) {
	s, err := p.AuthnSession.ResolveSession(authnSessionToken)
	if err != nil {
		return nil, err
	}

	step, ok := s.NextStep()
	if !ok {
		return nil, ErrInvalidAuthenticationSession
	}

	if !stepPredicate(step) {
		return nil, authz.ErrNotAuthenticated
	}

	return s, nil
}

func (p *Provider) StepSession(
	client config.OAuthClientConfiguration,
	s authn.Attributer,
	mfaBearerToken string,
) (Result, error) {
	switch s := s.(type) {
	case *AuthnSession:
		return p.AuthnSession.StepSession(s)
	case *session.IDPSession:
		err := p.Session.Update(s)
		if err != nil {
			return nil, err
		}
		return p.AuthnSession.MakeResult(client, s, mfaBearerToken)
	default:
		panic("authn: unexpected session container type")
	}
}
