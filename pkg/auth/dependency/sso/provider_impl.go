package sso

import (
	"github.com/skygeario/skygear-server/pkg/core/config"
)

type providerImpl struct {
	OAuthConfig *config.OAuthConfiguration
}

var _ Provider = &providerImpl{}

func NewProvider(c *config.OAuthConfiguration) Provider {
	return &providerImpl{
		OAuthConfig: c,
	}
}

func (f *providerImpl) EncodeState(state State) (encodedState string, err error) {
	return EncodeState(f.OAuthConfig.StateJWTSecret, state)
}

func (f *providerImpl) DecodeState(encodedState string) (*State, error) {
	return DecodeState(f.OAuthConfig.StateJWTSecret, encodedState)
}

func (f *providerImpl) EncodeSkygearAuthorizationCode(code SkygearAuthorizationCode) (encoded string, err error) {
	return EncodeSkygearAuthorizationCode(f.OAuthConfig.StateJWTSecret, code)
}

func (f *providerImpl) DecodeSkygearAuthorizationCode(encoded string) (*SkygearAuthorizationCode, error) {
	return DecodeSkygearAuthorizationCode(f.OAuthConfig.StateJWTSecret, encoded)
}
