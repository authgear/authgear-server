package apple

import (
	"context"
	"strings"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"

	"github.com/authgear/oauthrelyingparty/pkg/api/oauthrelyingparty"

	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
	liboauthrelyingparty "github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty"
	"github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/oauthrelyingpartyutil"
	"github.com/authgear/authgear-server/pkg/util/crypto"
	"github.com/authgear/authgear-server/pkg/util/duration"
	"github.com/authgear/authgear-server/pkg/util/jwtutil"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	oauthrelyingparty.RegisterProvider(Type, Apple{})
}

const Type = liboauthrelyingparty.TypeApple

type ProviderConfig oauthrelyingparty.ProviderConfig

func (c ProviderConfig) TeamID() string {
	team_id, _ := c["team_id"].(string)
	return team_id
}

func (c ProviderConfig) KeyID() string {
	key_id, _ := c["key_id"].(string)
	return key_id
}

var _ oauthrelyingparty.Provider = Apple{}
var _ liboauthrelyingparty.BuiltinProvider = Apple{}

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
		"key_id": { "type": "string" },
		"team_id": { "type": "string" }
	},
	"required": ["alias", "type", "client_id", "key_id", "team_id"]
}
`)

var appleOIDCConfig = oauthrelyingpartyutil.OIDCDiscoveryDocument{
	JWKSUri:               "https://appleid.apple.com/auth/keys",
	TokenEndpoint:         "https://appleid.apple.com/auth/token",
	AuthorizationEndpoint: "https://appleid.apple.com/auth/authorize",
}

type Apple struct{}

func (Apple) ValidateProviderConfig(ctx *validation.Context, cfg oauthrelyingparty.ProviderConfig) {
	ctx.AddError(Schema.Validator().ValidateValue(cfg))
}

func (Apple) SetDefaults(cfg oauthrelyingparty.ProviderConfig) {
	cfg.SetDefaultsModifyDisabledFalse()
	cfg.SetDefaultsEmailClaimConfig(oauthrelyingpartyutil.Email_AssumeVerified_Required())
}

func (Apple) ProviderID(cfg oauthrelyingparty.ProviderConfig) oauthrelyingparty.ProviderID {
	team_id := ProviderConfig(cfg).TeamID()
	// Apple supports OIDC.
	// sub is pairwise and is scoped to team_id.
	// Therefore, ProviderID is Type + team_id.
	//
	// Rotating the OAuth application is OK.
	// But rotating the Apple Developer account is problematic.
	// Since Apple has private relay to hide the real email,
	// the user may not be associate their account.
	keys := map[string]interface{}{
		"team_id": team_id,
	}
	return oauthrelyingparty.NewProviderID(cfg.Type(), keys)
}

func (Apple) scope() []string {
	// https://developer.apple.com/documentation/sign_in_with_apple/sign_in_with_apple_js/incorporating_sign_in_with_apple_into_other_platforms
	return []string{"name", "email"}
}

func (Apple) createClientSecret(deps oauthrelyingparty.Dependencies) (clientSecret string, err error) {
	teamID := ProviderConfig(deps.ProviderConfig).TeamID()
	keyID := ProviderConfig(deps.ProviderConfig).KeyID()

	// https://developer.apple.com/documentation/signinwithapplerestapi/generate_and_validate_tokens
	key, err := crypto.ParseAppleP8PrivateKey([]byte(deps.ClientSecret))
	if err != nil {
		return
	}

	now := deps.Clock.NowUTC()

	payload := jwt.New()
	_ = payload.Set(jwt.IssuerKey, teamID)
	_ = payload.Set(jwt.IssuedAtKey, now.Unix())
	_ = payload.Set(jwt.ExpirationKey, now.Add(duration.Short).Unix())
	_ = payload.Set(jwt.AudienceKey, "https://appleid.apple.com")
	_ = payload.Set(jwt.SubjectKey, deps.ProviderConfig.ClientID)

	jwkKey, err := jwk.FromRaw(key)
	if err != nil {
		return
	}
	_ = jwkKey.Set("kid", keyID)

	token, err := jwtutil.Sign(payload, jwa.ES256, jwkKey)
	if err != nil {
		return
	}

	clientSecret = string(token)
	return
}

func (p Apple) GetAuthorizationURL(deps oauthrelyingparty.Dependencies, param oauthrelyingparty.GetAuthorizationURLOptions) (string, error) {
	return appleOIDCConfig.MakeOAuthURL(oauthrelyingpartyutil.AuthorizationURLParams{
		ClientID:     deps.ProviderConfig.ClientID(),
		RedirectURI:  param.RedirectURI,
		Scope:        p.scope(),
		ResponseType: oauthrelyingparty.ResponseTypeCode,
		ResponseMode: param.ResponseMode,
		State:        param.State,
		// Prompt is unset.
		// Apple doesn't support prompt parameter
		// See "Send the Required Query Parameters" section for supporting parameters
		// https://developer.apple.com/documentation/sign_in_with_apple/sign_in_with_apple_js/incorporating_sign_in_with_apple_into_other_platforms
		Nonce: param.Nonce,
	}), nil
}

func (p Apple) GetUserProfile(deps oauthrelyingparty.Dependencies, param oauthrelyingparty.GetUserProfileOptions) (authInfo oauthrelyingparty.UserProfile, err error) {
	keySet, err := appleOIDCConfig.FetchJWKs(deps.HTTPClient)
	if err != nil {
		return
	}

	clientSecret, err := p.createClientSecret(deps)
	if err != nil {
		return
	}

	var tokenResp oauthrelyingpartyutil.AccessTokenResp
	jwtToken, err := appleOIDCConfig.ExchangeCode(
		deps.HTTPClient,
		deps.Clock,
		param.Code,
		keySet,
		deps.ProviderConfig.ClientID(),
		clientSecret,
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
	// https://developer.apple.com/documentation/signinwithapplerestapi/verifying_a_user
	// The exact spec is
	// Verify that the iss field contains https://appleid.apple.com
	// Therefore, we use strings.Contains here.
	iss, ok := claims["iss"].(string)
	if !ok {
		err = oauthrelyingpartyutil.OAuthProtocolError.New("iss not found in ID token")
		return
	}
	if !strings.Contains(iss, "https://appleid.apple.com") {
		err = oauthrelyingpartyutil.OAuthProtocolError.New("iss does not equal to `https://appleid.apple.com`")
		return
	}

	// Ensure sub exists
	sub, ok := claims["sub"].(string)
	if !ok {
		err = oauthrelyingpartyutil.OAuthProtocolError.New("sub not found in ID Token")
		return
	}

	// By observation, if the first time of authentication does NOT include the `name` scope,
	// Even the Services ID is unauthorized on https://appleid.apple.com,
	// and the `name` scope is included,
	// The ID Token still does not include the `name` claim.

	authInfo.ProviderRawProfile = claims
	authInfo.ProviderUserID = sub

	emailRequired := deps.ProviderConfig.EmailClaimConfig().Required()
	stdAttrs, err := stdattrs.Extract(claims, stdattrs.ExtractOptions{
		EmailRequired: emailRequired,
	})
	if err != nil {
		return
	}
	authInfo.StandardAttributes = stdAttrs.WithNameCopiedToGivenName()

	return
}
