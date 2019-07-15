package sso

type AuthInfoProcessor interface {
	DecodeUserInfo(p map[string]interface{}) ProviderUserInfo
}

type defaultAuthInfoProcessor struct{}

func newDefaultAuthInfoProcessor() defaultAuthInfoProcessor {
	return defaultAuthInfoProcessor{}
}

func (d defaultAuthInfoProcessor) DecodeUserInfo(userProfile map[string]interface{}) ProviderUserInfo {
	id, _ := userProfile["id"].(string)
	email, _ := userProfile["email"].(string)

	return ProviderUserInfo{
		ID:    id,
		Email: email,
	}
}
