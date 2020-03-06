package sso

import (
	"fmt"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/loginid"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

// UserInfoDecoder decodes user info.
type UserInfoDecoder interface {
	DecodeUserInfo(providerType config.OAuthProviderType, userInfo map[string]interface{}) (*ProviderUserInfo, error)
}

type UserInfoDecoderImpl struct {
	LoginIDNormalizerFactory loginid.LoginIDNormalizerFactory
}

func NewUserInfoDecoder(loginIDNormalizerFactory loginid.LoginIDNormalizerFactory) *UserInfoDecoderImpl {
	return &UserInfoDecoderImpl{
		LoginIDNormalizerFactory: loginIDNormalizerFactory,
	}
}

func (d *UserInfoDecoderImpl) DecodeUserInfo(providerType config.OAuthProviderType, userInfo map[string]interface{}) (providerUserInfo *ProviderUserInfo, err error) {
	switch providerType {
	case config.OAuthProviderTypeGoogle:
		providerUserInfo = d.decodeDefault(userInfo)
	case config.OAuthProviderTypeFacebook:
		providerUserInfo = d.decodeDefault(userInfo)
	case config.OAuthProviderTypeInstagram:
		providerUserInfo = d.decodeInstagram(userInfo)
	case config.OAuthProviderTypeLinkedIn:
		providerUserInfo = d.decodeDefault(userInfo)
	case config.OAuthProviderTypeAzureADv2:
		providerUserInfo = d.decodeAzureADv2(userInfo)
	case config.OAuthProviderTypeApple:
		providerUserInfo = d.decodeApple(userInfo)
	default:
		panic(fmt.Sprintf("sso: unknown provider type: %v", providerType))
	}

	normalizer := d.LoginIDNormalizerFactory.NormalizerWithLoginIDType(config.LoginIDKeyType("email"))
	email, err := normalizer.Normalize(providerUserInfo.Email)
	if err != nil {
		return
	}
	providerUserInfo.Email = email

	return
}

func (d *UserInfoDecoderImpl) decodeDefault(userInfo map[string]interface{}) *ProviderUserInfo {
	id, _ := userInfo["id"].(string)
	email, _ := userInfo["email"].(string)

	return &ProviderUserInfo{
		ID:    id,
		Email: email,
	}
}

func (d *UserInfoDecoderImpl) decodeInstagram(userInfo map[string]interface{}) *ProviderUserInfo {
	// Check GET /users/self response
	// https://www.instagram.com/developer/endpoints/users/
	info := &ProviderUserInfo{}
	data, ok := userInfo["data"].(map[string]interface{})
	if !ok {
		return info
	}

	info.ID, _ = data["id"].(string)
	info.Email, _ = data["email"].(string)
	return info
}

func (d *UserInfoDecoderImpl) decodeAzureADv2(userInfo map[string]interface{}) *ProviderUserInfo {
	id, _ := userInfo["oid"].(string)
	email, _ := userInfo["email"].(string)

	return &ProviderUserInfo{
		ID:    id,
		Email: email,
	}
}

func (d *UserInfoDecoderImpl) decodeApple(userInfo map[string]interface{}) *ProviderUserInfo {
	id, _ := userInfo["sub"].(string)
	email, _ := userInfo["email"].(string)

	return &ProviderUserInfo{
		ID:    id,
		Email: email,
	}
}
