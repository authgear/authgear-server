package declarative

import (
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/wechat"
)

type OAuthData struct {
	TypedData
	Alias                 string         `json:"alias,omitempty"`
	OAuthProviderType     string         `json:"oauth_provider_type,omitempty"`
	OAuthAuthorizationURL string         `json:"oauth_authorization_url,omitempty"`
	WechatAppType         wechat.AppType `json:"wechat_app_type,omitempty"`
}

var _ authflow.Data = OAuthData{}

func (OAuthData) Data() {}

func NewOAuthData(d OAuthData) OAuthData {
	d.Type = DataTypeOAuthData
	return d
}
