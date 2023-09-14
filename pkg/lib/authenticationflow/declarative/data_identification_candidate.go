package declarative

import (
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type IdentificationCandidate struct {
	Identification config.AuthenticationFlowIdentification `json:"identification"`

	// ProviderType is specific to OAuth.
	ProviderType config.OAuthSSOProviderType `json:"provider_type,omitempty"`
	// Alias is specific to OAuth.
	Alias string `json:"alias,omitempty"`
	// WechatAppType is specific to OAuth.
	WechatAppType config.OAuthSSOWeChatAppType `json:"wechat_app_type,omitempty"`
}

func NewIdentificationCandidates(identifications []config.AuthenticationFlowIdentification, oauthCandidates []IdentificationCandidate) []IdentificationCandidate {
	output := []IdentificationCandidate{}
	for _, identification := range identifications {
		switch identification {
		case config.AuthenticationFlowIdentificationEmail:
			fallthrough
		case config.AuthenticationFlowIdentificationPhone:
			fallthrough
		case config.AuthenticationFlowIdentificationUsername:
			output = append(output, IdentificationCandidate{
				Identification: identification,
			})
		case config.AuthenticationFlowIdentificationOAuth:
			output = append(output, oauthCandidates...)
		}
	}
	return output
}

func NewIdentificationCandidatesOAuth(oauthConfig *config.OAuthSSOConfig, oauthFeatureConfig *config.OAuthSSOProvidersFeatureConfig) []IdentificationCandidate {
	output := []IdentificationCandidate{}
	for _, p := range oauthConfig.Providers {
		if !identity.IsOAuthSSOProviderTypeDisabled(p.Type, oauthFeatureConfig) {
			output = append(output, IdentificationCandidate{
				Identification: config.AuthenticationFlowIdentificationOAuth,
				ProviderType:   p.Type,
				Alias:          p.Alias,
				WechatAppType:  p.AppType,
			})
		}
	}
	return output
}
