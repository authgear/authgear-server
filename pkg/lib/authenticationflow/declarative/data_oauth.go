package declarative

import (
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/wechat"
)

type OAuthProviderStatus string

const (
	OAuthProviderStatusActive               OAuthProviderStatus = "active"
	OAuthProviderStatusMissingCredentials   OAuthProviderStatus = "missing_credentials"
	OAuthProviderStatusUsingDemoCredentials OAuthProviderStatus = "using_demo_credentials"
)

type OAuthData struct {
	TypedData
	Alias                 string              `json:"alias,omitempty"`
	OAuthProviderType     string              `json:"oauth_provider_type,omitempty"`
	OAuthAuthorizationURL string              `json:"oauth_authorization_url,omitempty"`
	WechatAppType         wechat.AppType      `json:"wechat_app_type,omitempty"`
	ProviderStatus        OAuthProviderStatus `json:"provider_status,omitempty"`
}

var _ authflow.Data = OAuthData{}

func (OAuthData) Data() {}

func NewOAuthData(d OAuthData) OAuthData {
	d.Type = DataTypeOAuthData
	return d
}
