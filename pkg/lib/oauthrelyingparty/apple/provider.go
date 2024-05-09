package apple

import (
	"github.com/authgear/authgear-server/pkg/api/oauthrelyingparty"
	liboauthrelyingparty "github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty"
	"github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/oauthrelyingpartyutil"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	oauthrelyingparty.RegisterProvider(Type, Apple{})
}

const Type = liboauthrelyingparty.TypeApple

type ProviderConfig oauthrelyingparty.ProviderConfig

func (c ProviderConfig) TeamID() string {
	team_id, _ := c["team_id"].(string)
	return team_id
}

func (c ProviderConfig) KeyID() string {
	key_id, _ := c["key_id"].(string)
	return key_id
}

var _ oauthrelyingparty.Provider = Apple{}
var _ liboauthrelyingparty.BuiltinProvider = Apple{}

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
		"key_id": { "type": "string" },
		"team_id": { "type": "string" }
	},
	"required": ["alias", "type", "client_id", "key_id", "team_id"]
}
`)

type Apple struct{}

func (Apple) ValidateProviderConfig(ctx *validation.Context, cfg oauthrelyingparty.ProviderConfig) {
	ctx.AddError(Schema.Validator().ValidateValue(cfg))
}

func (Apple) SetDefaults(cfg oauthrelyingparty.ProviderConfig) {
	cfg.SetDefaultsModifyDisabledFalse()
	cfg.SetDefaultsEmailClaimConfig(oauthrelyingpartyutil.Email_AssumeVerified_Required())
}

func (Apple) ProviderID(cfg oauthrelyingparty.ProviderConfig) oauthrelyingparty.ProviderID {
	team_id := ProviderConfig(cfg).TeamID()
	// Apple supports OIDC.
	// sub is pairwise and is scoped to team_id.
	// Therefore, ProviderID is Type + team_id.
	//
	// Rotating the OAuth application is OK.
	// But rotating the Apple Developer account is problematic.
	// Since Apple has private relay to hide the real email,
	// the user may not be associate their account.
	keys := map[string]interface{}{
		"team_id": team_id,
	}
	return oauthrelyingparty.NewProviderID(cfg.Type(), keys)
}

func (Apple) Scope(_ oauthrelyingparty.ProviderConfig) []string {
	// https://developer.apple.com/documentation/sign_in_with_apple/sign_in_with_apple_js/incorporating_sign_in_with_apple_into_other_platforms
	return []string{"name", "email"}
}
