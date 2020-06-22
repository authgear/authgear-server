package sso

import (
	"fmt"

	"github.com/skygeario/skygear-server/pkg/core/config"
)

// UserInfoDecoder decodes user info.
type UserInfoDecoder interface {
	DecodeUserInfo(providerType config.OAuthProviderType, userInfo map[string]interface{}) (*ProviderUserInfo, error)
}

type UserInfoDecoderImpl struct {
	LoginIDNormalizerFactory LoginIDNormalizerFactory
}

func NewUserInfoDecoder(loginIDNormalizerFactory LoginIDNormalizerFactory) *UserInfoDecoderImpl {
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
	case config.OAuthProviderTypeLinkedIn:
		providerUserInfo = d.decodeLinkedIn(userInfo)
	case config.OAuthProviderTypeAzureADv2:
		providerUserInfo = d.decodeAzureADv2(userInfo)
	case config.OAuthProviderTypeApple:
		providerUserInfo = d.decodeApple(userInfo)
	default:
		panic(fmt.Sprintf("sso: unknown provider type: %v", providerType))
	}

	if providerUserInfo.Email != "" {
		var email string
		normalizer := d.LoginIDNormalizerFactory.NormalizerWithLoginIDType(config.LoginIDKeyType("email"))
		email, err = normalizer.Normalize(providerUserInfo.Email)
		if err != nil {
			return
		}
		providerUserInfo.Email = email
	}

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

func (d *UserInfoDecoderImpl) decodeLinkedIn(userInfo map[string]interface{}) *ProviderUserInfo {
	profile := userInfo["profile"].(map[string]interface{})
	id := profile["id"].(string)

	email := ""
	primaryContact := userInfo["primary_contact"].(map[string]interface{})
	elements := primaryContact["elements"].([]interface{})
	for _, e := range elements {
		element := e.(map[string]interface{})
		if primary, ok := element["primary"].(bool); !ok || !primary {
			continue
		}
		if typ, ok := element["type"].(string); !ok || typ != "EMAIL" {
			continue
		}
		handleTilde, ok := element["handle~"].(map[string]interface{})
		if !ok {
			continue
		}
		email, _ = handleTilde["emailAddress"].(string)
	}

	return &ProviderUserInfo{
		ID:    id,
		Email: email,
	}
}
