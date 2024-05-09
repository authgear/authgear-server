package wechat

import (
	"strconv"

	"github.com/authgear/authgear-server/pkg/api/oauthrelyingparty"
	liboauthrelyingparty "github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty"
	"github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/oauthrelyingpartyutil"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	oauthrelyingparty.RegisterProvider(Type, Wechat{})
}

type AppType string

const (
	AppTypeWeb    AppType = "web"
	AppTypeMobile AppType = "mobile"
)

type ProviderConfig oauthrelyingparty.ProviderConfig

func (c ProviderConfig) AppType() AppType {
	app_type, _ := c["app_type"].(string)
	return AppType(app_type)
}

func (c ProviderConfig) AccountID() string {
	account_id, _ := c["account_id"].(string)
	return account_id
}

func (c ProviderConfig) IsSandboxAccount() bool {
	is_sandbox_account, _ := c["is_sandbox_account"].(bool)
	return is_sandbox_account
}

func (c ProviderConfig) WechatRedirectURIs() []string {
	var out []string
	wechat_redirect_uris, _ := c["wechat_redirect_uris"].([]interface{})
	for _, iface := range wechat_redirect_uris {
		if s, ok := iface.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

const Type = liboauthrelyingparty.TypeWechat

var _ oauthrelyingparty.Provider = Wechat{}
var _ liboauthrelyingparty.BuiltinProvider = Wechat{}

var Schema = validation.NewSimpleSchema(`
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"alias": { "type": "string" },
		"type": { "type": "string" },
		"modify_disabled": { "type": "boolean" },
		"client_id": { "type": "string", "minLength": 1 },
		"claims": {
			"type": "object",
			"additionalProperties": false,
			"properties": {
				"email": {
					"type": "object",
					"additionalProperties": false,
					"properties": {
						"assume_verified": { "type": "boolean" },
						"required": { "type": "boolean" }
					}
				}
			}
		},
		"app_type": { "type": "string", "enum": ["mobile", "web"] },
		"account_id": { "type": "string", "format": "wechat_account_id" },
		"is_sandbox_account": { "type": "boolean" },
		"wechat_redirect_uris": { "type": "array", "items": { "type": "string", "format": "uri" } }
	},
	"required": ["alias", "type", "client_id", "app_type", "account_id"]
}
`)

type Wechat struct{}

func (Wechat) ValidateProviderConfig(ctx *validation.Context, cfg oauthrelyingparty.ProviderConfig) {
	ctx.AddError(Schema.Validator().ValidateValue(cfg))
}

func (Wechat) SetDefaults(cfg oauthrelyingparty.ProviderConfig) {
	cfg.SetDefaultsModifyDisabledFalse()
	cfg.SetDefaultsEmailClaimConfig(oauthrelyingpartyutil.Email_AssumeVerified_NOT_Required())
}

func (Wechat) ProviderID(cfg oauthrelyingparty.ProviderConfig) oauthrelyingparty.ProviderID {
	// WeChat does NOT support OIDC.
	// In the same Weixin Open Platform account, the user UnionID is unique.
	// The id is scoped to Open Platform account.
	// https://developers.weixin.qq.com/miniprogram/en/dev/framework/open-ability/union-id.html

	wechatCfg := ProviderConfig(cfg)
	account_id := wechatCfg.AccountID()
	is_sandbox_account := wechatCfg.IsSandboxAccount()
	keys := map[string]interface{}{
		"account_id":         account_id,
		"is_sandbox_account": strconv.FormatBool(is_sandbox_account),
	}

	return oauthrelyingparty.NewProviderID(cfg.Type(), keys)
}

func (Wechat) Scope(_ oauthrelyingparty.ProviderConfig) []string {
	// https://developers.weixin.qq.com/doc/offiaccount/OA_Web_Apps/Wechat_webpage_authorization.html
	return []string{"snsapi_userinfo"}
}
