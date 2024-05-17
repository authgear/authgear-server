package azureadv2

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

func (AzureADv2) scope() []string {
	// https://docs.microsoft.com/en-us/azure/active-directory/develop/v2-permissions-and-consent#openid-connect-scopes
	return []string{"openid", "profile", "email"}
}

func (AzureADv2) getOpenIDConfiguration(deps oauthrelyingparty.Dependencies) (*oauthrelyingpartyutil.OIDCDiscoveryDocument, error) {
	// OPTIMIZE(sso): Cache OpenID configuration

	tenant := ProviderConfig(deps.ProviderConfig).Tenant()

	var endpoint string
	// Azure special tenant
	//
	// If the azure tenant is `organizations` or `common`,
	// the developer should make use of `before_user_create` and `before_identity_create` hook
	// to disallow any undesire identity.
	// The `raw_profile` of the identity is the ID Token claims.
	// Refer to https://docs.microsoft.com/en-us/azure/active-directory/develop/id-tokens
	// to see what claims the token could contain.
	//
	// For `organizations`, the user can be any user of any organizational AD.
	// Therefore the developer should have a whitelist of AD tenant IDs.
	// In the incoming hook, check if `tid` matches one of the entry of the whitelist.
	//
	// For `common`, in addition to the users from `organizations`, any Microsoft personal account
	// could be the user.
	// In case of personal account, the `tid` is "9188040d-6c67-4c5b-b112-36a304b66dad".
	// Therefore the developer should first check if `tid` indicates personal account.
	// If yes, apply their logic to disallow the user creation.
	// One very common example is to look at the claim `email`.
	// Use a email address parser to parse the email address.
	// Obtain the domain and check if the domain is whitelisted.
	// For example, if the developer only wants user from hotmail.com to create user,
	// ensure `tid` is "9188040d-6c67-4c5b-b112-36a304b66dad" and ensure `email`
	// is of domain `@hotmail.com`.

	// As of 2019-09-23, two special values are observed.
	// To discover these values, create a new client
	// and try different options.
	switch tenant {
	// Special value for any organizational AD
	case "organizations":
		endpoint = "https://login.microsoftonline.com/organizations/v2.0/.well-known/openid-configuration"
	// Special value for any organizational AD and personal accounts (Xbox etc)
	case "common":
		endpoint = "https://login.microsoftonline.com/common/v2.0/.well-known/openid-configuration"
	default:
		endpoint = fmt.Sprintf("https://login.microsoftonline.com/%s/v2.0/.well-known/openid-configuration", tenant)
	}

	return oauthrelyingpartyutil.FetchOIDCDiscoveryDocument(deps.HTTPClient, endpoint)
}

func (p AzureADv2) GetAuthorizationURL(deps oauthrelyingparty.Dependencies, param oauthrelyingparty.GetAuthorizationURLOptions) (string, error) {
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

func (p AzureADv2) GetUserProfile(deps oauthrelyingparty.Dependencies, param oauthrelyingparty.GetUserProfileOptions) (authInfo oauthrelyingparty.UserProfile, err error) {
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

	oid, ok := claims["oid"].(string)
	if !ok {
		err = oauthrelyingpartyutil.OAuthProtocolError.New("oid not found in ID Token")
		return
	}
	// For "Microsoft Account", email usually exists.
	// For "AD guest user", email usually exists because to invite an user, the inviter must provide email.
	// For "AD user", email never exists even one is provided in "Authentication Methods".

	authInfo.ProviderRawProfile = claims
	authInfo.ProviderUserID = oid
	emailRequired := deps.ProviderConfig.EmailClaimConfig().Required()
	stdAttrs, err := stdattrs.Extract(claims, stdattrs.ExtractOptions{
		EmailRequired: emailRequired,
	})
	if err != nil {
		return
	}
	authInfo.StandardAttributes = stdAttrs

	return
}

func (AzureADv2) getPrompt(prompt []string) []string {
	// Azureadv2 only supports single value for prompt.
	// The first supporting value in the list will be used.
	// The usage of `none` is for checking existing authentication and/or consent
	// which doesn't fit auth ui case.
	// https://docs.microsoft.com/en-us/azure/active-directory/develop/v2-oauth2-auth-code-flow
	for _, p := range prompt {
		if p == "login" {
			return []string{"login"}
		} else if p == "consent" {
			return []string{"consent"}
		} else if p == "select_account" {
			return []string{"select_account"}
		}
	}
	return []string{}
}
