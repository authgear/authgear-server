package sso

import (
	"github.com/skygeario/skygear-server/pkg/auth/model"
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

func (f *providerImpl) IsAllowedOnUserDuplicate(a model.OnUserDuplicate) bool {
	return model.IsAllowedOnUserDuplicate(
		f.OAuthConfig.OnUserDuplicateAllowMerge,
		f.OAuthConfig.OnUserDuplicateAllowCreate,
		a,
	)
}

func (f *providerImpl) IsValidCallbackURL(u string) bool {
	err := ValidateCallbackURL(f.OAuthConfig.AllowedCallbackURLs, u)
	return err == nil
}

func (f *providerImpl) IsExternalAccessTokenFlowEnabled() bool {
	return f.OAuthConfig.ExternalAccessTokenFlowEnabled
}
