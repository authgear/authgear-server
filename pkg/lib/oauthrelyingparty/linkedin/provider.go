package linkedin

import (
	"context"

	"github.com/authgear/oauthrelyingparty/pkg/api/oauthrelyingparty"

	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
	liboauthrelyingparty "github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty"
	"github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/oauthrelyingpartyutil"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	oauthrelyingparty.RegisterProvider(Type, Linkedin{})
}

const Type = liboauthrelyingparty.TypeLinkedin

var _ oauthrelyingparty.Provider = Linkedin{}

const (
	linkedinOIDCDiscoveryDocumentURL string = "https://www.linkedin.com/oauth/.well-known/openid-configuration"
)

type Linkedin struct{}

func (Linkedin) GetJSONSchema() map[string]interface{} {
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
		)
	builder.Required("type", "client_id")
	return builder
}

func (Linkedin) SetDefaults(cfg oauthrelyingparty.ProviderConfig) {
	cfg.SetDefaultsEmailClaimConfig(oauthrelyingpartyutil.Email_AssumeVerified_Required())
}

func (Linkedin) ProviderID(cfg oauthrelyingparty.ProviderConfig) oauthrelyingparty.ProviderID {
	// Linkedin supports OIDC.
	// sub is pairwise and is scoped to client_id.
	// Therefore, ProviderID is Type + client_id.
	//
	// Rotating the OAuth application is problematic.
	keys := map[string]interface{}{
		"client_id": cfg.ClientID(),
	}
	return oauthrelyingparty.NewProviderID(cfg.Type(), keys)
}

func (Linkedin) scope() []string {
	// https://learn.microsoft.com/en-us/linkedin/consumer/integrations/self-serve/sign-in-with-linkedin-v2#authenticating-members
	return []string{"openid", "profile", "email"}
}

func (p Linkedin) GetAuthorizationURL(deps oauthrelyingparty.Dependencies, param oauthrelyingparty.GetAuthorizationURLOptions) (string, error) {
	d, err := oauthrelyingpartyutil.FetchOIDCDiscoveryDocument(deps.HTTPClient, linkedinOIDCDiscoveryDocumentURL)
	if err != nil {
		return "", err
	}
	return d.MakeOAuthURL(oauthrelyingpartyutil.AuthorizationURLParams{
		ClientID:     deps.ProviderConfig.ClientID(),
		RedirectURI:  param.RedirectURI,
		Scope:        p.scope(),
		ResponseType: oauthrelyingparty.ResponseTypeCode,
		// ResponseMode is unset.
		State: param.State,
		// Prompt is unset.
		// Linkedin doesn't support prompt parameter
		// https://docs.microsoft.com/en-us/linkedin/shared/authentication/authorization-code-flow?tabs=HTTPS#step-2-request-an-authorization-code

		// Nonce is unset
	}), nil
}

func (Linkedin) GetUserProfile(deps oauthrelyingparty.Dependencies, param oauthrelyingparty.GetUserProfileOptions) (authInfo oauthrelyingparty.UserProfile, err error) {
	code, err := oauthrelyingpartyutil.GetCode(param.Query)
	if err != nil {
		return
	}

	d, err := oauthrelyingpartyutil.FetchOIDCDiscoveryDocument(deps.HTTPClient, linkedinOIDCDiscoveryDocumentURL)
	if err != nil {
		return
	}

	keySet, err := d.FetchJWKs(deps.HTTPClient)
	if err != nil {
		return
	}

	var tokenResp oauthrelyingpartyutil.AccessTokenResp
	jwtToken, err := d.ExchangeCode(
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
	if iss != d.Issuer {
		err = oauthrelyingpartyutil.OAuthProtocolError.New("iss is not from LinkedIn")
		return
	}

	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		err = oauthrelyingpartyutil.OAuthProtocolError.New("sub not found in ID Token")
		return
	}

	rawProfile, err := d.FetchUserInfo(deps.HTTPClient, tokenResp)
	if err != nil {
		return
	}

	authInfo.ProviderRawProfile = rawProfile
	authInfo.ProviderUserID = sub

	emailRequired := deps.ProviderConfig.EmailClaimConfig().Required()
	stdAttrs, err := stdattrs.Extract(rawProfile, stdattrs.ExtractOptions{
		EmailRequired: emailRequired,
	})
	if err != nil {
		return
	}
	authInfo.StandardAttributes = stdAttrs

	return
}
