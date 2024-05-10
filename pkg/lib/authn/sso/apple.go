package sso

import (
	"context"
	"strings"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"

	"github.com/authgear/authgear-server/pkg/api/oauthrelyingparty"
	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
	"github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/apple"
	"github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/oauthrelyingpartyutil"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/crypto"
	"github.com/authgear/authgear-server/pkg/util/duration"
	"github.com/authgear/authgear-server/pkg/util/jwtutil"
)

var appleOIDCConfig = OIDCDiscoveryDocument{
	JWKSUri:               "https://appleid.apple.com/auth/keys",
	TokenEndpoint:         "https://appleid.apple.com/auth/token",
	AuthorizationEndpoint: "https://appleid.apple.com/auth/authorize",
}

type AppleImpl struct {
	Clock                        clock.Clock
	ProviderConfig               oauthrelyingparty.ProviderConfig
	ClientSecret                 string
	StandardAttributesNormalizer StandardAttributesNormalizer
	HTTPClient                   OAuthHTTPClient
}

func (f *AppleImpl) createClientSecret() (clientSecret string, err error) {
	teamID := apple.ProviderConfig(f.ProviderConfig).TeamID()
	keyID := apple.ProviderConfig(f.ProviderConfig).KeyID()

	// https://developer.apple.com/documentation/signinwithapplerestapi/generate_and_validate_tokens
	key, err := crypto.ParseAppleP8PrivateKey([]byte(f.ClientSecret))
	if err != nil {
		return
	}

	now := f.Clock.NowUTC()

	payload := jwt.New()
	_ = payload.Set(jwt.IssuerKey, teamID)
	_ = payload.Set(jwt.IssuedAtKey, now.Unix())
	_ = payload.Set(jwt.ExpirationKey, now.Add(duration.Short).Unix())
	_ = payload.Set(jwt.AudienceKey, "https://appleid.apple.com")
	_ = payload.Set(jwt.SubjectKey, f.ProviderConfig.ClientID)

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

func (f *AppleImpl) Config() oauthrelyingparty.ProviderConfig {
	return f.ProviderConfig
}

func (f *AppleImpl) GetAuthorizationURL(param oauthrelyingparty.GetAuthorizationURLOptions) (string, error) {
	return appleOIDCConfig.MakeOAuthURL(oauthrelyingpartyutil.AuthorizationURLParams{
		ClientID:     f.ProviderConfig.ClientID(),
		RedirectURI:  param.RedirectURI,
		Scope:        f.ProviderConfig.Scope(),
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

func (f *AppleImpl) GetAuthInfo(param GetAuthInfoParam) (authInfo AuthInfo, err error) {
	keySet, err := appleOIDCConfig.FetchJWKs(f.HTTPClient)
	if err != nil {
		return
	}

	clientSecret, err := f.createClientSecret()
	if err != nil {
		return
	}

	var tokenResp oauthrelyingpartyutil.AccessTokenResp
	jwtToken, err := appleOIDCConfig.ExchangeCode(
		f.HTTPClient,
		f.Clock,
		param.Code,
		keySet,
		f.ProviderConfig.ClientID(),
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
		err = OAuthProtocolError.New("iss not found in ID token")
		return
	}
	if !strings.Contains(iss, "https://appleid.apple.com") {
		err = OAuthProtocolError.New("iss does not equal to `https://appleid.apple.com`")
		return
	}

	// Ensure sub exists
	sub, ok := claims["sub"].(string)
	if !ok {
		err = OAuthProtocolError.New("sub not found in ID Token")
		return
	}

	// By observation, if the first time of authentication does NOT include the `name` scope,
	// Even the Services ID is unauthorized on https://appleid.apple.com,
	// and the `name` scope is included,
	// The ID Token still does not include the `name` claim.

	authInfo.ProviderRawProfile = claims
	authInfo.ProviderUserID = sub

	emailRequired := f.ProviderConfig.EmailClaimConfig().Required()
	stdAttrs, err := stdattrs.Extract(claims, stdattrs.ExtractOptions{
		EmailRequired: emailRequired,
	})
	if err != nil {
		return
	}
	authInfo.StandardAttributes = stdAttrs.WithNameCopiedToGivenName()

	err = f.StandardAttributesNormalizer.Normalize(authInfo.StandardAttributes)
	if err != nil {
		return
	}

	return
}

var (
	_ OAuthProvider = &AppleImpl{}
)
