package authn

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/loginid"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/session"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

type Provider struct {
	OAuth   *OAuthCoordinator
	Authn   *AuthenticateProcess
	Signup  *SignupProcess
	Session *SessionProviderImpl
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

	s, err := p.Session.BeginSession(client, pr.PrincipalUserID(), pr, session.CreateReasonSignup)
	if err != nil {
		return nil, err
	}

	return p.Session.StepSession(s)
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

	s, err := p.Session.BeginSession(client, pr.PrincipalUserID(), pr, session.CreateReasonLogin)
	if err != nil {
		return nil, err
	}

	return p.Session.StepSession(s)
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
	code *sso.SkygearAuthorizationCode,
) (Result, error) {
	pr, err := p.OAuth.ExchangeCode(code)
	if err != nil {
		return nil, err
	}

	reason := session.CreateReason(code.SessionCreateReason)
	s, err := p.Session.BeginSession(client, pr.PrincipalUserID(), pr, reason)
	if err != nil {
		return nil, err
	}

	return p.Session.StepSession(s)
}
