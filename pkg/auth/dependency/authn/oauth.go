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

func (c *OAuthCoordinator) Authenticate(authInfo sso.AuthInfo, codeChallenge string, loginState sso.LoginState) (code *sso.SkygearAuthorizationCode, err error) {
	// Authenticate using SSO info
	p, err := c.Authn.AuthenticateWithOAuth(authInfo)
	if err == nil {
		return &sso.SkygearAuthorizationCode{
			Action:              "login",
			CodeChallenge:       codeChallenge,
			UserID:              p.PrincipalUserID(),
			PrincipalID:         p.PrincipalID(),
			SessionCreateReason: string(auth.SessionCreateReasonLogin),
		}, nil
	}
	if err != nil && !errors.Is(err, principal.ErrNotFound) {
		return nil, err
	}

	// Authentication failed, try signing up a new user
	p, err = c.Signup.SignupWithOAuth(authInfo, loginState.OnUserDuplicate)
	if err == nil {
		return &sso.SkygearAuthorizationCode{
			Action:              "login",
			CodeChallenge:       codeChallenge,
			UserID:              p.PrincipalUserID(),
			PrincipalID:         p.PrincipalID(),
			SessionCreateReason: string(auth.SessionCreateReasonSignup),
		}, nil
	}
	var mergeErr *oAuthRequireMergeError
	if err != nil && !errors.As(err, &mergeErr) {
		return nil, err
	}

	// Signup failed and require user merge, try linking existing user
	p, err = c.Signup.LinkWithOAuth(authInfo, mergeErr.UserID)
	if err != nil {
		return nil, err
	}

	return &sso.SkygearAuthorizationCode{
		Action:              "login",
		CodeChallenge:       codeChallenge,
		UserID:              p.PrincipalUserID(),
		PrincipalID:         p.PrincipalID(),
		SessionCreateReason: string(auth.SessionCreateReasonLogin),
	}, nil
}

func (c *OAuthCoordinator) Link(authInfo sso.AuthInfo, codeChallenge string, linkState sso.LinkState) (code *sso.SkygearAuthorizationCode, err error) {
	p, err := c.Signup.LinkWithOAuth(authInfo, linkState.UserID)
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
