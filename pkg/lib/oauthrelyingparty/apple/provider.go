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

var appleOIDCConfig = oauthrelyingpartyutil.OIDCDiscoveryDocument{
	JWKSUri:               "https://appleid.apple.com/auth/keys",
	TokenEndpoint:         "https://appleid.apple.com/auth/token",
	AuthorizationEndpoint: "https://appleid.apple.com/auth/authorize",
}

type Apple struct{}

func (Apple) GetJSONSchema() map[string]interface{} {
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
		Property("key_id", validation.SchemaBuilder{}.Type(validation.TypeString)).
		Property("team_id", validation.SchemaBuilder{}.Type(validation.TypeString))
	builder.Required("type", "client_id", "key_id", "team_id")
	return builder
}

func (Apple) SetDefaults(cfg oauthrelyingparty.ProviderConfig) {
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
	// See this documentation on how to create a client secret
	// https://developer.apple.com/documentation/accountorganizationaldatasharing/creating-a-client-secret
	//
	// It was observed that Sign in with Apple has a weird behavior.
	// When the client_id (Services ID) is Team A, and the account signed in is a managed account under Team A,
	// client_secret IS NOT validated at all.
	//
	// For example, suppose the team is @mycompany.com.
	// johndoe@mycompany.com (a managed Apple ID account under the team @mycompany.com) can sign in
	// event client_secret is invalid.
	//
	// When client_secret is invalid, {"error": "invalid_client"} is returned.
	// In that case, you need to refer to this documentation to resolve the problem.
	// https://developer.apple.com/documentation/technotes/tn3107-resolving-sign-in-with-apple-response-errors#Possible-reasons-for-invalid-client-errors

	teamID := ProviderConfig(deps.ProviderConfig).TeamID()
	keyID := ProviderConfig(deps.ProviderConfig).KeyID()

	key, err := crypto.ParseAppleP8PrivateKey([]byte(deps.ClientSecret))
	if err != nil {
		return
	}

	now := deps.Clock.NowUTC()

	payload := jwt.New()
	_ = payload.Set(jwt.IssuerKey, teamID)
	_ = payload.Set(jwt.IssuedAtKey, now.Unix())
	_ = payload.Set(jwt.ExpirationKey, now.Add(duration.Short).Unix())

	// According to the documentation, aud is a string, not an array of string.
	payload.Options().Enable(jwt.FlattenAudience)
	_ = payload.Set(jwt.AudienceKey, "https://appleid.apple.com")

	_ = payload.Set(jwt.SubjectKey, deps.ProviderConfig.ClientID())

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

func (p Apple) GetAuthorizationURL(ctx context.Context, deps oauthrelyingparty.Dependencies, param oauthrelyingparty.GetAuthorizationURLOptions) (string, error) {
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

func (p Apple) GetUserProfile(ctx context.Context, deps oauthrelyingparty.Dependencies, param oauthrelyingparty.GetUserProfileOptions) (authInfo oauthrelyingparty.UserProfile, err error) {
	keySet, err := appleOIDCConfig.FetchJWKs(ctx, deps.HTTPClient)
	if err != nil {
		return
	}

	clientSecret, err := p.createClientSecret(deps)
	if err != nil {
		return
	}

	code, err := oauthrelyingpartyutil.GetCode(param.Query)
	if err != nil {
		return
	}

	var tokenResp oauthrelyingpartyutil.AccessTokenResp
	jwtToken, err := appleOIDCConfig.ExchangeCode(
		ctx,
		deps.HTTPClient,
		deps.Clock,
		code,
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

	claims, err := jwtToken.AsMap(oauthrelyingpartyutil.ContextForTheUnusedContextArgumentInJWXV2API)
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
