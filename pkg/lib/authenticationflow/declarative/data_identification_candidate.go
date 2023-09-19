package declarative

import (
	"github.com/authgear/authgear-server/pkg/api/model"
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

	// WebAuthnRequestOptions is specific to Passkey.
	RequestOptions *model.WebAuthnRequestOptions `json:"request_options,omitempty"`
}

func NewIdentificationCandidateLoginID(i config.AuthenticationFlowIdentification) IdentificationCandidate {
	return IdentificationCandidate{
		Identification: i,
	}
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

func NewIdentificationCandidatePasskey(requestOptions *model.WebAuthnRequestOptions) IdentificationCandidate {
	return IdentificationCandidate{
		Identification: config.AuthenticationFlowIdentificationPasskey,
		RequestOptions: requestOptions,
	}
}
