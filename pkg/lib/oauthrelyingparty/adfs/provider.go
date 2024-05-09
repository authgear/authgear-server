package adfs

import (
	"github.com/authgear/authgear-server/pkg/api/oauthrelyingparty"
	liboauthrelyingparty "github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty"
	"github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/oauthrelyingpartyutil"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	oauthrelyingparty.RegisterProvider(Type, ADFS{})
}

const Type = liboauthrelyingparty.TypeADFS

type ProviderConfig oauthrelyingparty.ProviderConfig

func (c ProviderConfig) DiscoveryDocumentEndpoint() string {
	discovery_document_endpoint, _ := c["discovery_document_endpoint"].(string)
	return discovery_document_endpoint
}

var _ oauthrelyingparty.Provider = ADFS{}
var _ liboauthrelyingparty.BuiltinProvider = ADFS{}

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
		"discovery_document_endpoint": { "type": "string", "format": "uri" }
	},
	"required": ["alias", "type", "client_id", "discovery_document_endpoint"]
}
`)

type ADFS struct{}

func (ADFS) ValidateProviderConfig(ctx *validation.Context, cfg oauthrelyingparty.ProviderConfig) {
	ctx.AddError(Schema.Validator().ValidateValue(cfg))
}

func (ADFS) SetDefaults(cfg oauthrelyingparty.ProviderConfig) {
	cfg.SetDefaultsModifyDisabledFalse()
	cfg.SetDefaultsEmailClaimConfig(oauthrelyingpartyutil.Email_AssumeVerified_Required())
}

func (ADFS) ProviderID(cfg oauthrelyingparty.ProviderConfig) oauthrelyingparty.ProviderID {
	// In the original implementation, provider ID is just type.
	return oauthrelyingparty.NewProviderID(cfg.Type(), nil)
}

func (ADFS) Scope(_ oauthrelyingparty.ProviderConfig) []string {
	// The supported scopes are observed from a AD FS server.
	return []string{"openid", "profile", "email"}
}
