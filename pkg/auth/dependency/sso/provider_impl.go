package sso

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"

	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/apiclientconfig"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

type providerImpl struct {
	AppID          string
	ClientProvider apiclientconfig.Provider
	OAuthConfig    *config.OAuthConfiguration
}

var _ Provider = &providerImpl{}

func NewProvider(appID string, clientProvider apiclientconfig.Provider, c *config.OAuthConfiguration) Provider {
	return &providerImpl{
		AppID:          appID,
		ClientProvider: clientProvider,
		OAuthConfig:    c,
	}
}

func (f *providerImpl) EncodeState(state State) (encodedState string, err error) {
	return EncodeState(f.OAuthConfig.StateJWTSecret, f.AppID, state)
}

func (f *providerImpl) DecodeState(encodedState string) (*State, error) {
	return DecodeState(f.OAuthConfig.StateJWTSecret, f.AppID, encodedState)
}

func (f *providerImpl) EncodeSkygearAuthorizationCode(code SkygearAuthorizationCode) (encoded string, err error) {
	return EncodeSkygearAuthorizationCode(f.OAuthConfig.StateJWTSecret, f.AppID, code)
}

func (f *providerImpl) DecodeSkygearAuthorizationCode(encoded string) (*SkygearAuthorizationCode, error) {
	return DecodeSkygearAuthorizationCode(f.OAuthConfig.StateJWTSecret, f.AppID, encoded)
}

func (f *providerImpl) IsAllowedOnUserDuplicate(a model.OnUserDuplicate) bool {
	return model.IsAllowedOnUserDuplicate(
		f.OAuthConfig.OnUserDuplicateAllowMerge,
		f.OAuthConfig.OnUserDuplicateAllowCreate,
		a,
	)
}

func (f *providerImpl) IsValidCallbackURL(u string) bool {
	var redirectURIs []string
	_, client, ok := f.ClientProvider.Get()
	if ok {
		redirectURIs = client.RedirectURIs()
	}
	err := ValidateCallbackURL(redirectURIs, u)
	return err == nil
}

func (f *providerImpl) IsExternalAccessTokenFlowEnabled() bool {
	return f.OAuthConfig.ExternalAccessTokenFlowEnabled
}

func (f *providerImpl) VerifyPKCE(code *SkygearAuthorizationCode, codeVerifier string) error {
	sha256Arr := sha256.Sum256([]byte(codeVerifier))
	sha256Slice := sha256Arr[:]
	codeChallenge := base64.RawURLEncoding.EncodeToString(sha256Slice)
	if subtle.ConstantTimeCompare([]byte(code.CodeChallenge), []byte(codeChallenge)) != 1 {
		return NewSSOFailed(InvalidCodeVerifier, "invalid code verifier")
	}
	return nil
}
