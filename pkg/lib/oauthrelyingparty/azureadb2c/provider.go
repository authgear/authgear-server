package azureadb2c

import (
	"github.com/authgear/authgear-server/pkg/api/oauthrelyingparty"
	liboauthrelyingparty "github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty"
	"github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/oauthrelyingpartyutil"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	oauthrelyingparty.RegisterProvider(Type, AzureADB2C{})
}

const Type = liboauthrelyingparty.TypeAzureADB2C

type ProviderConfig oauthrelyingparty.ProviderConfig

func (c ProviderConfig) Tenant() string {
	tenant, _ := c["tenant"].(string)
	return tenant
}

func (c ProviderConfig) Policy() string {
	policy, _ := c["policy"].(string)
	return policy
}

var _ oauthrelyingparty.Provider = AzureADB2C{}
var _ liboauthrelyingparty.BuiltinProvider = AzureADB2C{}

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
		"tenant": { "type": "string" },
		"policy": { "type": "string" }
	},
	"required": ["alias", "type", "client_id", "tenant", "policy"]
}
`)

type AzureADB2C struct{}

func (AzureADB2C) ValidateProviderConfig(ctx *validation.Context, cfg oauthrelyingparty.ProviderConfig) {
	ctx.AddError(Schema.Validator().ValidateValue(cfg))
}

func (AzureADB2C) SetDefaults(cfg oauthrelyingparty.ProviderConfig) {
	cfg.SetDefaultsModifyDisabledFalse()
	cfg.SetDefaultsEmailClaimConfig(oauthrelyingpartyutil.Email_AssumeVerified_Required())
}

func (AzureADB2C) ProviderID(cfg oauthrelyingparty.ProviderConfig) oauthrelyingparty.ProviderID {
	// By default sub is the Object ID of the user in the directory.
	// A tenant is a directory.
	// sub is scoped to the tenant only.
	// Therefore, ProviderID is Type + tenant.
	//
	// See https://docs.microsoft.com/en-us/azure/active-directory-b2c/tokens-overview#claims
	tenant := ProviderConfig(cfg).Tenant()
	keys := map[string]interface{}{
		"tenant": tenant,
	}
	return oauthrelyingparty.NewProviderID(cfg.Type(), keys)
}

func (AzureADB2C) Scope(_ oauthrelyingparty.ProviderConfig) []string {
	// Instead of specifying scope to request a specific claim,
	// the developer must customize the policy to allow which claims are returned to the relying party.
	// If the developer is using User Flow policy, then those claims are called Application Claims.
	return []string{"openid"}
}
