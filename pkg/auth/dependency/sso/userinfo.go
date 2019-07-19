package sso

import (
	"fmt"

	"github.com/skygeario/skygear-server/pkg/core/config"
)

// UserInfoDecoder decodes user info.
type UserInfoDecoder interface {
	DecodeUserInfo(p map[string]interface{}) ProviderUserInfo
}

type DefaultUserInfoDecoder struct{}

func NewDefaultUserInfoDecoder() DefaultUserInfoDecoder {
	return DefaultUserInfoDecoder{}
}

func (d DefaultUserInfoDecoder) DecodeUserInfo(userProfile map[string]interface{}) ProviderUserInfo {
	id, _ := userProfile["id"].(string)
	email, _ := userProfile["email"].(string)

	return ProviderUserInfo{
		ID:    id,
		Email: email,
	}
}

type InstagramUserInfoDecoder struct{}

func NewInstagramUserInfoDecoder() InstagramUserInfoDecoder {
	return InstagramUserInfoDecoder{}
}

func (d InstagramUserInfoDecoder) DecodeUserInfo(userProfile map[string]interface{}) (info ProviderUserInfo) {
	// Check GET /users/self response
	// https://www.instagram.com/developer/endpoints/users/
	data, ok := userProfile["data"].(map[string]interface{})
	if !ok {
		return
	}

	info.ID, _ = data["id"].(string)
	info.Email, _ = data["email"].(string)
	return
}

type AzureADv2UserInfoDecoder struct{}

func NewAzureADv2UserInfoDecoder() AzureADv2UserInfoDecoder {
	return AzureADv2UserInfoDecoder{}
}

func (d AzureADv2UserInfoDecoder) DecodeUserInfo(userProfile map[string]interface{}) ProviderUserInfo {

	id, _ := userProfile["oid"].(string)
	email, _ := userProfile["email"].(string)

	return ProviderUserInfo{
		ID:    id,
		Email: email,
	}
}

func GetUserInfoDecoder(providerType config.OAuthProviderType) UserInfoDecoder {
	switch providerType {
	case config.OAuthProviderTypeGoogle:
		return NewDefaultUserInfoDecoder()
	case config.OAuthProviderTypeFacebook:
		return NewDefaultUserInfoDecoder()
	case config.OAuthProviderTypeInstagram:
		return NewInstagramUserInfoDecoder()
	case config.OAuthProviderTypeLinkedIn:
		return NewDefaultUserInfoDecoder()
	case config.OAuthProviderTypeAzureADv2:
		return NewAzureADv2UserInfoDecoder()
	}
	panic(fmt.Errorf("unknown provider type: %v", providerType))
}
