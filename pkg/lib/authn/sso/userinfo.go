package sso

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

type UserInfoDecoder struct {
	LoginIDNormalizerFactory LoginIDNormalizerFactory
}

func (d *UserInfoDecoder) DecodeUserInfo(providerType config.OAuthSSOProviderType, userInfo map[string]interface{}) (providerUserInfo *ProviderUserInfo, err error) {
	switch providerType {
	case config.OAuthSSOProviderTypeGoogle:
		providerUserInfo = DecodeDefault(userInfo)
	case config.OAuthSSOProviderTypeFacebook:
		providerUserInfo = DecodeDefault(userInfo)
	case config.OAuthSSOProviderTypeLinkedIn:
		providerUserInfo = DecodeLinkedIn(userInfo)
	case config.OAuthSSOProviderTypeAzureADv2:
		providerUserInfo = DecodeAzureADv2(userInfo)
	case config.OAuthSSOProviderTypeApple:
		providerUserInfo = DecodeApple(userInfo)
	case config.OAuthSSOProviderTypeWechat:
		providerUserInfo = DecodeWechat(userInfo)
	default:
		panic(fmt.Sprintf("sso: unknown provider type: %v", providerType))
	}

	if providerUserInfo.Email != "" {
		var email string
		normalizer := d.LoginIDNormalizerFactory.NormalizerWithLoginIDType(config.LoginIDKeyTypeEmail)
		email, err = normalizer.Normalize(providerUserInfo.Email)
		if err != nil {
			return
		}
		providerUserInfo.Email = email
	}

	return
}

func DecodeDefault(userInfo map[string]interface{}) *ProviderUserInfo {
	id, _ := userInfo["id"].(string)
	email, _ := userInfo["email"].(string)

	return &ProviderUserInfo{
		ID:    id,
		Email: email,
	}
}

func DecodeAzureADv2(userInfo map[string]interface{}) *ProviderUserInfo {
	id, _ := userInfo["oid"].(string)
	email, _ := userInfo["email"].(string)

	return &ProviderUserInfo{
		ID:    id,
		Email: email,
	}
}

func DecodeApple(userInfo map[string]interface{}) *ProviderUserInfo {
	id, _ := userInfo["sub"].(string)
	email, _ := userInfo["email"].(string)

	return &ProviderUserInfo{
		ID:    id,
		Email: email,
	}
}

func DecodeLinkedIn(userInfo map[string]interface{}) *ProviderUserInfo {
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

func DecodeWechat(userInfo map[string]interface{}) *ProviderUserInfo {
	id, _ := userInfo["openid"].(string)

	return &ProviderUserInfo{
		ID: id,
	}
}
