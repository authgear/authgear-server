package authn

import (
	"github.com/sirupsen/logrus"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/loginid"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/core/errors"
	coreTime "github.com/skygeario/skygear-server/pkg/core/time"
)

// AuthenticateProcess handles user authentication: validate credentials and return a principal
type AuthenticateProcess struct {
	Logger           *logrus.Entry
	TimeProvider     coreTime.Provider
	PasswordProvider password.Provider
	OAuthProvider    oauth.Provider
	IdentityProvider principal.IdentityProvider
}

func (p *AuthenticateProcess) AuthenticateWithLoginID(loginID loginid.LoginID, plainPassword string) (prin principal.Principal, err error) {
	var passwordPrincipal password.Principal
	realm := password.DefaultRealm
	err = p.PasswordProvider.GetPrincipalByLoginIDWithRealm(loginID.Key, loginID.Value, realm, &passwordPrincipal)
	if err != nil {
		if errors.Is(err, principal.ErrNotFound) {
			err = password.ErrInvalidCredentials
		}
		if errors.Is(err, principal.ErrMultipleResultsFound) {
			p.Logger.WithError(err).Warn("multiple results found for password principal query")
			err = password.ErrInvalidCredentials
		}
		return
	}

	err = passwordPrincipal.VerifyPassword(plainPassword)
	if err != nil {
		return
	}

	if err := p.PasswordProvider.MigratePassword(&passwordPrincipal, plainPassword); err != nil {
		p.Logger.WithError(err).Error("failed to migrate password")
	}

	prin = &passwordPrincipal
	return
}

func (p *AuthenticateProcess) AuthenticateWithOAuth(oauthAuthInfo sso.AuthInfo) (principal.Principal, error) {
	oauthPrincipal, err := p.findExistingOAuthPrincipal(oauthAuthInfo)
	if err != nil && !errors.Is(err, principal.ErrNotFound) {
		return nil, err
	}

	now := p.TimeProvider.NowUTC()

	// Case: OAuth principal was found
	// => Simple update case
	// We do not need to consider other principals
	if err == nil {
		oauthPrincipal.AccessTokenResp = oauthAuthInfo.ProviderAccessTokenResp
		oauthPrincipal.UserProfile = oauthAuthInfo.ProviderRawProfile
		oauthPrincipal.ClaimsValue = oauthAuthInfo.ProviderUserInfo.ClaimsValue()
		oauthPrincipal.UpdatedAt = &now
		if err = p.OAuthProvider.UpdatePrincipal(oauthPrincipal); err != nil {
			return nil, err
		}
		// Always return here because we are done with this case.
		return oauthPrincipal, nil
	}

	// Case: OAuth principal was not found
	// => Cannot authenticate, may need to merge existing user.
	return nil, principal.ErrNotFound
}

func (p *AuthenticateProcess) findExistingOAuthPrincipal(oauthAuthInfo sso.AuthInfo) (*oauth.Principal, error) {
	// Find oauth principal from by (provider_id, provider_user_id)
	return p.OAuthProvider.GetPrincipalByProvider(oauth.GetByProviderOptions{
		ProviderType:   string(oauthAuthInfo.ProviderConfig.Type),
		ProviderKeys:   oauth.ProviderKeysFromProviderConfig(oauthAuthInfo.ProviderConfig),
		ProviderUserID: oauthAuthInfo.ProviderUserInfo.ID,
	})
}

func (p *AuthenticateProcess) AuthenticateAsPrincipal(principalID string) (principal.Principal, error) {
	prin, err := p.IdentityProvider.GetPrincipalByID(principalID)
	if err != nil {
		return nil, err
	}

	return prin, nil
}
