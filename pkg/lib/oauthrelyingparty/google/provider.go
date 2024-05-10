package google

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/oauthrelyingparty"
	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
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

const (
	googleOIDCDiscoveryDocumentURL string = "https://accounts.google.com/.well-known/openid-configuration"
)

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

func (Google) scope() []string {
	// https://developers.google.com/identity/protocols/oauth2/openid-connect
	return []string{"openid", "profile", "email"}
}

func (p Google) GetAuthorizationURL(deps oauthrelyingparty.Dependencies, param oauthrelyingparty.GetAuthorizationURLOptions) (string, error) {
	d, err := oauthrelyingpartyutil.FetchOIDCDiscoveryDocument(deps.HTTPClient, googleOIDCDiscoveryDocumentURL)
	if err != nil {
		return "", err
	}
	return d.MakeOAuthURL(oauthrelyingpartyutil.AuthorizationURLParams{
		ClientID:     deps.ProviderConfig.ClientID(),
		RedirectURI:  param.RedirectURI,
		Scope:        p.scope(),
		ResponseType: oauthrelyingparty.ResponseTypeCode,
		ResponseMode: param.ResponseMode,
		State:        param.State,
		Nonce:        param.Nonce,
		Prompt:       p.getPrompt(param.Prompt),
	}), nil
}

func (Google) GetUserProfile(deps oauthrelyingparty.Dependencies, param oauthrelyingparty.GetUserProfileOptions) (authInfo oauthrelyingparty.UserProfile, err error) {
	d, err := oauthrelyingpartyutil.FetchOIDCDiscoveryDocument(deps.HTTPClient, googleOIDCDiscoveryDocumentURL)
	if err != nil {
		return
	}
	// OPTIMIZE(sso): Cache JWKs
	keySet, err := d.FetchJWKs(deps.HTTPClient)
	if err != nil {
		return
	}

	var tokenResp oauthrelyingpartyutil.AccessTokenResp
	jwtToken, err := d.ExchangeCode(
		deps.HTTPClient,
		deps.Clock,
		param.Code,
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

	// Verify the issuer
	// https://developers.google.com/identity/protocols/OpenIDConnect#validatinganidtoken
	iss, ok := claims["iss"].(string)
	if !ok {
		err = oauthrelyingpartyutil.OAuthProtocolError.New("iss not found in ID token")
		return
	}
	if iss != "https://accounts.google.com" && iss != "accounts.google.com" {
		err = oauthrelyingpartyutil.OAuthProtocolError.New("iss is not from Google")
		return
	}

	// Ensure sub exists
	sub, ok := claims["sub"].(string)
	if !ok {
		err = oauthrelyingpartyutil.OAuthProtocolError.New("sub not found in ID token")
		return
	}

	authInfo.ProviderRawProfile = claims
	authInfo.ProviderUserID = sub
	// Google supports
	// given_name, family_name, email, picture, profile, locale
	// https://developers.google.com/identity/protocols/oauth2/openid-connect#obtainuserinfo
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

func (Google) getPrompt(prompt []string) []string {
	// Google supports `none`, `consent` and `select_account` for prompt.
	// The usage of `none` is for checking existing authentication and/or consent
	// which doesn't fit auth ui case.
	// https://developers.google.com/identity/protocols/oauth2/openid-connect#authenticationuriparameters
	newPrompt := []string{}
	for _, p := range prompt {
		if p == "consent" ||
			p == "select_account" {
			newPrompt = append(newPrompt, p)
		}
	}
	if len(newPrompt) == 0 {
		// default
		return []string{"select_account"}
	}
	return newPrompt
}
