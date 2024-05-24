package azureadb2c

import (
	"context"
	"fmt"

	"github.com/authgear/oauthrelyingparty/pkg/api/oauthrelyingparty"

	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
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

type AzureADB2C struct{}

func (AzureADB2C) GetJSONSchema() map[string]interface{} {
	builder := validation.SchemaBuilder{}
	builder.Type(validation.TypeObject)
	builder.Properties().
		Property("type", validation.SchemaBuilder{}.Type(validation.TypeString)).
		Property("client_id", validation.SchemaBuilder{}.Type(validation.TypeString).MinLength(1)).
		Property("claims", validation.SchemaBuilder{}.Type(validation.TypeObject).
			AdditionalPropertiesFalse().
			Properties().
			Property("email", validation.SchemaBuilder{}.Type(validation.TypeObject).
				AdditionalPropertiesFalse().Properties().
				Property("assume_verified", validation.SchemaBuilder{}.Type(validation.TypeBoolean)).
				Property("required", validation.SchemaBuilder{}.Type(validation.TypeBoolean)),
			),
		).
		Property("tenant", validation.SchemaBuilder{}.Type(validation.TypeString)).
		Property("policy", validation.SchemaBuilder{}.Type(validation.TypeString))
	builder.Required("type", "client_id", "tenant", "policy")
	return builder
}

func (AzureADB2C) SetDefaults(cfg oauthrelyingparty.ProviderConfig) {
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

func (AzureADB2C) scope() []string {
	// Instead of specifying scope to request a specific claim,
	// the developer must customize the policy to allow which claims are returned to the relying party.
	// If the developer is using User Flow policy, then those claims are called Application Claims.
	return []string{"openid"}
}

func (AzureADB2C) getOpenIDConfiguration(deps oauthrelyingparty.Dependencies) (*oauthrelyingpartyutil.OIDCDiscoveryDocument, error) {
	azureadb2cConfig := ProviderConfig(deps.ProviderConfig)
	tenant := azureadb2cConfig.Tenant()
	policy := azureadb2cConfig.Policy()

	endpoint := fmt.Sprintf(
		"https://%s.b2clogin.com/%s.onmicrosoft.com/%s/v2.0/.well-known/openid-configuration",
		tenant,
		tenant,
		policy,
	)

	return oauthrelyingpartyutil.FetchOIDCDiscoveryDocument(deps.HTTPClient, endpoint)
}

func (p AzureADB2C) GetAuthorizationURL(deps oauthrelyingparty.Dependencies, param oauthrelyingparty.GetAuthorizationURLOptions) (string, error) {
	c, err := p.getOpenIDConfiguration(deps)
	if err != nil {
		return "", err
	}
	return c.MakeOAuthURL(oauthrelyingpartyutil.AuthorizationURLParams{
		ClientID:     deps.ProviderConfig.ClientID(),
		RedirectURI:  param.RedirectURI,
		Scope:        p.scope(),
		ResponseType: oauthrelyingparty.ResponseTypeCode,
		ResponseMode: param.ResponseMode,
		State:        param.State,
		Prompt:       p.getPrompt(param.Prompt),
		Nonce:        param.Nonce,
	}), nil
}

func (p AzureADB2C) GetUserProfile(deps oauthrelyingparty.Dependencies, param oauthrelyingparty.GetUserProfileOptions) (authInfo oauthrelyingparty.UserProfile, err error) {
	c, err := p.getOpenIDConfiguration(deps)
	if err != nil {
		return
	}
	// OPTIMIZE(sso): Cache JWKs
	keySet, err := c.FetchJWKs(deps.HTTPClient)
	if err != nil {
		return
	}

	code, err := oauthrelyingpartyutil.GetCode(param.Query)
	if err != nil {
		return
	}

	var tokenResp oauthrelyingpartyutil.AccessTokenResp
	jwtToken, err := c.ExchangeCode(
		deps.HTTPClient,
		deps.Clock,
		code,
		keySet,
		deps.ProviderConfig.ClientID(),
		deps.ClientSecret,
		param.RedirectURI,
		param.Nonce,
		&tokenResp,
	)
	if err != nil {
		return
	}

	claims, err := jwtToken.AsMap(context.TODO())
	if err != nil {
		return
	}

	iss, ok := claims["iss"].(string)
	if !ok {
		err = oauthrelyingpartyutil.OAuthProtocolError.New("iss not found in ID token")
		return
	}
	if iss != c.Issuer {
		err = oauthrelyingpartyutil.OAuthProtocolError.New(
			fmt.Sprintf("iss: %v != %v", iss, c.Issuer),
		)
		return
	}

	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		err = oauthrelyingpartyutil.OAuthProtocolError.New("sub not found in ID Token")
		return
	}

	authInfo.ProviderRawProfile = claims
	authInfo.ProviderUserID = sub

	stdAttrs, err := p.extract(deps, claims)
	if err != nil {
		return
	}
	authInfo.StandardAttributes = stdAttrs

	return
}

func (AzureADB2C) extract(deps oauthrelyingparty.Dependencies, claims map[string]interface{}) (stdattrs.T, error) {
	// Here is the list of possible builtin claims of user flows
	// https://learn.microsoft.com/en-us/azure/active-directory-b2c/user-flow-overview#user-flows
	// city: free text
	// country: free text
	// jobTitle: free text
	// legalAgeGroupClassification: a enum with undocumented variants
	// postalCode: free text
	// state: free text
	// streetAddress: free text
	// newUser: true means the user signed up newly
	// oid: sub is identical to it by default.
	// emails: if non-empty, the first value corresponds to standard claim
	// name: correspond to standard claim
	// given_name: correspond to standard claim
	// family_name: correspond to standard claim

	// For custom policy we further recognize the following claims.
	// https://learn.microsoft.com/en-us/azure/active-directory-b2c/user-profile-attributes
	// signInNames.emailAddress: string

	extractString := func(input map[string]interface{}, output stdattrs.T, key string) {
		if value, ok := input[key].(string); ok && value != "" {
			output[key] = value
		}
	}

	out := stdattrs.T{}

	extractString(claims, out, stdattrs.Name)
	extractString(claims, out, stdattrs.GivenName)
	extractString(claims, out, stdattrs.FamilyName)

	var email string
	if email == "" {
		if ifaceSlice, ok := claims["emails"].([]interface{}); ok {
			for _, iface := range ifaceSlice {
				if str, ok := iface.(string); ok && str != "" {
					email = str
				}
			}
		}
	}
	if email == "" {
		if str, ok := claims["signInNames.emailAddress"].(string); ok {
			if str != "" {
				email = str
			}
		}
	}
	out[stdattrs.Email] = email

	emailRequired := deps.ProviderConfig.EmailClaimConfig().Required()
	return stdattrs.Extract(out, stdattrs.ExtractOptions{
		EmailRequired: emailRequired,
	})
}

func (AzureADB2C) getPrompt(prompt []string) []string {
	// The only supported value is login.
	// See https://docs.microsoft.com/en-us/azure/active-directory-b2c/openid-connect
	for _, p := range prompt {
		if p == "login" {
			return []string{"login"}
		}
	}
	return []string{}
}
