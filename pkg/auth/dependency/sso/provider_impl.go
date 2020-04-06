package sso

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"

	"github.com/skygeario/skygear-server/pkg/core/config"
)

type providerImpl struct {
	AppID       string
	Context     context.Context
	OAuthConfig *config.OAuthConfiguration
}

var _ Provider = &providerImpl{}

func NewProvider(ctx context.Context, appID string, c *config.OAuthConfiguration) Provider {
	return &providerImpl{
		AppID:       appID,
		Context:     ctx,
		OAuthConfig: c,
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

func (f *providerImpl) IsValidCallbackURL(client config.OAuthClientConfiguration, u string) bool {
	var redirectURIs []string
	if client != nil {
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
