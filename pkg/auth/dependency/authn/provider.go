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
	Session *SessionProvider
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
	client config.OAuthClientConfiguration,
	session *session.Session,
	authInfo sso.AuthInfo,
	linkState sso.LinkState,
) (Result, error) {
	code, err := p.OAuth.Link(authInfo, "", linkState)
	if err != nil {
		return nil, err
	}

	_, err = p.OAuth.ExchangeCode(code)
	if err != nil {
		return nil, err
	}

	return p.Session.MakeResult(client, session, "")
}

func (p *Provider) OAuthExchangeCode(
	client config.OAuthClientConfiguration,
	s *session.Session,
	code *sso.SkygearAuthorizationCode,
) (Result, error) {
	pr, err := p.OAuth.ExchangeCode(code)
	if err != nil {
		return nil, err
	}

	if code.Action == "link" {
		return p.Session.MakeResult(client, s, "")
	}

	// code.Action == "login"
	reason := session.CreateReason(code.SessionCreateReason)
	as, err := p.Session.BeginSession(client, pr.PrincipalUserID(), pr, reason)
	if err != nil {
		return nil, err
	}

	return p.Session.StepSession(as)
}
