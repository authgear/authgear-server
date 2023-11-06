package declarative

import (
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type oauthData struct {
	Alias                 string                       `json:"alias,omitempty"`
	OAuthProviderType     config.OAuthSSOProviderType  `json:"oauth_provider_type,omitempty"`
	OAuthAuthorizationURL string                       `json:"oauth_authorization_url,omitempty"`
	WechatAppType         config.OAuthSSOWeChatAppType `json:"wechat_app_type,omitempty"`
}

var _ authflow.Data = oauthData{}

func (oauthData) Data() {}
