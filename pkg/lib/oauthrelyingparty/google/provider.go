package google

import (
	"github.com/authgear/authgear-server/pkg/api/oauthrelyingparty"
	liboauthrelyingparty "github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty"
	"github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/oauthrelyingpartyutil"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	oauthrelyingparty.RegisterProvider(Type, Google{})
}

const Type = liboauthrelyingparty.TypeGoogle

var _ oauthrelyingparty.Provider = Google{}
var _ liboauthrelyingparty.BuiltinProvider = Google{}

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
		}
	},
	"required": ["alias", "type", "client_id"]
}
`)

type Google struct{}

func (Google) ValidateProviderConfig(ctx *validation.Context, cfg oauthrelyingparty.ProviderConfig) {
	ctx.AddError(Schema.Validator().ValidateValue(cfg))
}

func (Google) SetDefaults(cfg oauthrelyingparty.ProviderConfig) {
	cfg.SetDefaultsModifyDisabledFalse()
	cfg.SetDefaultsEmailClaimConfig(oauthrelyingpartyutil.Email_AssumeVerified_Required())
}

func (Google) ProviderID(cfg oauthrelyingparty.ProviderConfig) oauthrelyingparty.ProviderID {
	// Google supports OIDC.
	// sub is public, not scoped to anything so changing client_id does not affect sub.
	// Therefore, ProviderID is simply the type.
	//
	// Rotating the OAuth application is OK.
	return oauthrelyingparty.NewProviderID(cfg.Type(), nil)
}

func (Google) Scope(_ oauthrelyingparty.ProviderConfig) []string {
	// https://developers.google.com/identity/protocols/oauth2/openid-connect
	return []string{"openid", "profile", "email"}
}
