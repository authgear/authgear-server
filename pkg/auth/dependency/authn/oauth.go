package authn

import (
	"errors"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
)

// OAuthCoordinator controls OAuth SSO flow
type OAuthCoordinator struct {
	Authn  *AuthenticateProcess
	Signup *SignupProcess
}

func (c *OAuthCoordinator) AuthenticateCode(authInfo sso.AuthInfo, codeChallenge string, loginState sso.LoginState) (code *sso.SkygearAuthorizationCode, err error) {
	p, sessionCreateReason, err := c.Authenticate(authInfo, loginState)
	if err != nil {
		return nil, err
	}

	return &sso.SkygearAuthorizationCode{
		Action:              "login",
		CodeChallenge:       codeChallenge,
		UserID:              p.PrincipalUserID(),
		PrincipalID:         p.PrincipalID(),
		SessionCreateReason: string(sessionCreateReason),
	}, nil
}

func (c *OAuthCoordinator) LinkCode(authInfo sso.AuthInfo, codeChallenge string, linkState sso.LinkState) (code *sso.SkygearAuthorizationCode, err error) {
	p, err := c.Link(authInfo, linkState)
	if err != nil {
		return nil, err
	}
	return &sso.SkygearAuthorizationCode{
		Action:        "link",
		CodeChallenge: codeChallenge,
		UserID:        p.PrincipalUserID(),
		PrincipalID:   p.PrincipalID(),
	}, nil
}

func (c *OAuthCoordinator) ExchangeCode(code *sso.SkygearAuthorizationCode) (principal.Principal, error) {
	return c.Authn.AuthenticateAsPrincipal(code.PrincipalID)
}

func (c *OAuthCoordinator) Authenticate(authInfo sso.AuthInfo, loginState sso.LoginState) (principal.Principal, auth.SessionCreateReason, error) {
	// Authenticate using SSO info
	p, err := c.Authn.AuthenticateWithOAuth(authInfo)
	if err == nil {
		return p, auth.SessionCreateReasonLogin, nil
	}
	if !errors.Is(err, principal.ErrNotFound) {
		return nil, "", err
	}

	// Authentication failed, try signing up a new user
	p, err = c.Signup.SignupWithOAuth(authInfo, loginState.OnUserDuplicate)
	if err == nil {
		return p, auth.SessionCreateReasonSignup, nil
	}
	var mergeErr *oAuthRequireMergeError
	if !errors.As(err, &mergeErr) {
		return nil, "", err
	}

	// Signup failed and require user merge, try linking existing user
	p, err = c.Signup.LinkWithOAuth(authInfo, mergeErr.UserID)
	if err == nil {
		return p, auth.SessionCreateReasonLogin, nil
	}
	return nil, "", err
}

func (c *OAuthCoordinator) Link(authInfo sso.AuthInfo, linkState sso.LinkState) (principal.Principal, error) {
	return c.Signup.LinkWithOAuth(authInfo, linkState.UserID)
}
