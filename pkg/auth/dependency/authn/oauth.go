package authn

import (
	"errors"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
)

type AuthorizationCodeStore interface {
	Get(codeHash string) (*sso.SkygearAuthorizationCode, error)
	Set(code *sso.SkygearAuthorizationCode) error
	Delete(codeHash string) error
}

// OAuthCoordinator controls OAuth SSO flow
type OAuthCoordinator struct {
	Authn                  *AuthenticateProcess
	Signup                 *SignupProcess
	AuthorizationCodeStore AuthorizationCodeStore
}

func (c *OAuthCoordinator) AuthenticateCode(authInfo sso.AuthInfo, codeChallenge string, loginState sso.LoginState) (*sso.SkygearAuthorizationCode, string, error) {
	p, sessionCreateReason, err := c.Authenticate(authInfo, loginState)
	if err != nil {
		return nil, "", err
	}

	codeStr := sso.GenerateCode()
	codeHash := sso.HashCode(codeStr)
	code := &sso.SkygearAuthorizationCode{
		CodeHash:            codeHash,
		Action:              "login",
		CodeChallenge:       codeChallenge,
		UserID:              p.PrincipalUserID(),
		PrincipalID:         p.PrincipalID(),
		SessionCreateReason: string(sessionCreateReason),
	}

	err = c.AuthorizationCodeStore.Set(code)
	if err != nil {
		return nil, "", err
	}
	return code, codeStr, nil
}

func (c *OAuthCoordinator) LinkCode(authInfo sso.AuthInfo, codeChallenge string, linkState sso.LinkState) (*sso.SkygearAuthorizationCode, string, error) {
	p, err := c.Link(authInfo, linkState)
	if err != nil {
		return nil, "", err
	}
	codeStr := sso.GenerateCode()
	codeHash := sso.HashCode(codeStr)
	code := &sso.SkygearAuthorizationCode{
		CodeHash:      codeHash,
		Action:        "link",
		CodeChallenge: codeChallenge,
		UserID:        p.PrincipalUserID(),
		PrincipalID:   p.PrincipalID(),
	}
	err = c.AuthorizationCodeStore.Set(code)
	if err != nil {
		return nil, "", err
	}
	return code, codeStr, nil
}

func (c *OAuthCoordinator) ConsumeCode(codeHash string) (*sso.SkygearAuthorizationCode, error) {
	code, err := c.AuthorizationCodeStore.Get(codeHash)
	if err != nil {
		return nil, err
	}
	err = c.AuthorizationCodeStore.Delete(codeHash)
	if err != nil {
		return nil, err
	}
	return code, nil
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
