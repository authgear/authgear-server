package azureadv2

import (
	"github.com/authgear/authgear-server/pkg/api/oauthrelyingparty"
	liboauthrelyingparty "github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty"
	"github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/oauthrelyingpartyutil"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	oauthrelyingparty.RegisterProvider(Type, AzureADv2{})
}

const Type = liboauthrelyingparty.TypeAzureADv2

type ProviderConfig oauthrelyingparty.ProviderConfig

func (c ProviderConfig) Tenant() string {
	tenant, _ := c["tenant"].(string)
	return tenant
}

var _ oauthrelyingparty.Provider = AzureADv2{}
var _ liboauthrelyingparty.BuiltinProvider = AzureADv2{}

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
		"tenant": { "type": "string" }
	},
	"required": ["alias", "type", "client_id", "tenant"]
}
`)

type AzureADv2 struct{}

func (AzureADv2) ValidateProviderConfig(ctx *validation.Context, cfg oauthrelyingparty.ProviderConfig) {
	ctx.AddError(Schema.Validator().ValidateValue(cfg))
}

func (AzureADv2) SetDefaults(cfg oauthrelyingparty.ProviderConfig) {
	cfg.SetDefaultsModifyDisabledFalse()
	cfg.SetDefaultsEmailClaimConfig(oauthrelyingpartyutil.Email_AssumeVerified_Required())
}

func (AzureADv2) ProviderID(cfg oauthrelyingparty.ProviderConfig) oauthrelyingparty.ProviderID {
	// Azure AD v2 supports OIDC.
	// sub is pairwise and is scoped to client_id.
	// However, oid is powerful alternative to sub.
	// oid is also pairwise and is scoped to tenant.
	// We use oid as ProviderSubjectID so ProviderID is Type + tenant.
	//
	// Rotating the OAuth application is OK.
	// But rotating the tenant is problematic.
	// But if email remains unchanged, the user can associate their account.
	tenant := ProviderConfig(cfg).Tenant()
	keys := map[string]interface{}{
		"tenant": tenant,
	}
	return oauthrelyingparty.NewProviderID(cfg.Type(), keys)
}

func (AzureADv2) Scope(_ oauthrelyingparty.ProviderConfig) []string {
	// https://docs.microsoft.com/en-us/azure/active-directory/develop/v2-permissions-and-consent#openid-connect-scopes
	return []string{"openid", "profile", "email"}
}
