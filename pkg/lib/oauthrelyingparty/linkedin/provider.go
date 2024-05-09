package linkedin

import (
	"github.com/authgear/authgear-server/pkg/api/oauthrelyingparty"
	liboauthrelyingparty "github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty"
	"github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/oauthrelyingpartyutil"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	oauthrelyingparty.RegisterProvider(Type, Linkedin{})
}

const Type = liboauthrelyingparty.TypeLinkedin

var _ oauthrelyingparty.Provider = Linkedin{}
var _ liboauthrelyingparty.BuiltinProvider = Linkedin{}

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

type Linkedin struct{}

func (Linkedin) ValidateProviderConfig(ctx *validation.Context, cfg oauthrelyingparty.ProviderConfig) {
	ctx.AddError(Schema.Validator().ValidateValue(cfg))
}

func (Linkedin) SetDefaults(cfg oauthrelyingparty.ProviderConfig) {
	cfg.SetDefaultsModifyDisabledFalse()
	cfg.SetDefaultsEmailClaimConfig(oauthrelyingpartyutil.Email_AssumeVerified_Required())
}

func (Linkedin) ProviderID(cfg oauthrelyingparty.ProviderConfig) oauthrelyingparty.ProviderID {
	// Linkedin does NOT support OIDC.
	// Linkedin user ID is scoped to client_id.
	// Therefore, ProviderID is Type + client_id.
	//
	// Rotating the OAuth application is problematic.
	keys := map[string]interface{}{
		"client_id": cfg.ClientID(),
	}
	return oauthrelyingparty.NewProviderID(cfg.Type(), keys)
}

func (Linkedin) Scope(_ oauthrelyingparty.ProviderConfig) []string {
	// https://docs.microsoft.com/en-us/linkedin/shared/references/v2/profile/lite-profile
	// https://docs.microsoft.com/en-us/linkedin/shared/integrations/people/primary-contact-api?context=linkedin/compliance/context
	return []string{"r_liteprofile", "r_emailaddress"}
}
