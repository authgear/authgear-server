package facebook

import (
	"github.com/authgear/authgear-server/pkg/api/oauthrelyingparty"
	liboauthrelyingparty "github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty"
	"github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/oauthrelyingpartyutil"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	oauthrelyingparty.RegisterProvider(Type, Facebook{})
}

const Type = liboauthrelyingparty.TypeFacebook

var _ oauthrelyingparty.Provider = Facebook{}
var _ liboauthrelyingparty.BuiltinProvider = Facebook{}

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

type Facebook struct{}

func (Facebook) ValidateProviderConfig(ctx *validation.Context, cfg oauthrelyingparty.ProviderConfig) {
	ctx.AddError(Schema.Validator().ValidateValue(cfg))
}

func (Facebook) SetDefaults(cfg oauthrelyingparty.ProviderConfig) {
	cfg.SetDefaultsModifyDisabledFalse()
	cfg.SetDefaultsEmailClaimConfig(oauthrelyingpartyutil.Email_AssumeVerified_Required())
}

func (Facebook) ProviderID(cfg oauthrelyingparty.ProviderConfig) oauthrelyingparty.ProviderID {
	// Facebook does NOT support OIDC.
	// Facebook user ID is scoped to client_id.
	// Therefore, ProviderID is Type + client_id.
	//
	// Rotating the OAuth application is problematic.
	// But if email remains unchanged, the user can associate their account.
	keys := map[string]interface{}{
		"client_id": cfg.ClientID(),
	}
	return oauthrelyingparty.NewProviderID(cfg.Type(), keys)
}

func (Facebook) Scope(_ oauthrelyingparty.ProviderConfig) []string {
	// https://developers.facebook.com/docs/permissions/reference
	return []string{"email", "public_profile"}
}
