package sso

import (
	"context"
	"net/http"
	"strings"

	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jwt"

	"github.com/authgear/authgear-server/pkg/lib/config"
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
	Clock                    clock.Clock
	RedirectURL              RedirectURLProvider
	ProviderConfig           config.OAuthSSOProviderConfig
	Credentials              config.OAuthClientCredentialsItem
	LoginIDNormalizerFactory LoginIDNormalizerFactory
}

func (f *AppleImpl) createClientSecret() (clientSecret string, err error) {
	// https://developer.apple.com/documentation/signinwithapplerestapi/generate_and_validate_tokens
	key, err := crypto.ParseAppleP8PrivateKey([]byte(f.Credentials.ClientSecret))
	if err != nil {
		return
	}

	now := f.Clock.NowUTC()

	payload := jwt.New()
	_ = payload.Set(jwt.IssuerKey, f.ProviderConfig.TeamID)
	_ = payload.Set(jwt.IssuedAtKey, now.Unix())
	_ = payload.Set(jwt.ExpirationKey, now.Add(duration.Short).Unix())
	_ = payload.Set(jwt.AudienceKey, "https://appleid.apple.com")
	_ = payload.Set(jwt.SubjectKey, f.ProviderConfig.ClientID)

	jwkKey, err := jwk.New(key)
	if err != nil {
		return
	}
	_ = jwkKey.Set("kid", f.ProviderConfig.KeyID)

	token, err := jwtutil.Sign(payload, jwa.ES256, jwkKey)
	if err != nil {
		return
	}

	clientSecret = string(token)
	return
}

func (*AppleImpl) Type() config.OAuthSSOProviderType {
	return config.OAuthSSOProviderTypeApple
}

func (f *AppleImpl) Config() config.OAuthSSOProviderConfig {
	return f.ProviderConfig
}

func (f *AppleImpl) GetAuthURL(param GetAuthURLParam) (string, error) {
	return appleOIDCConfig.MakeOAuthURL(OIDCAuthParams{
		ProviderConfig: f.ProviderConfig,
		RedirectURI:    f.RedirectURL.SSOCallbackURL(f.ProviderConfig).String(),
		Nonce:          param.Nonce,
		State:          param.State,
		Prompt:         f.GetPrompt(param.Prompt),
	}), nil
}

func (f *AppleImpl) GetAuthInfo(r OAuthAuthorizationResponse, param GetAuthInfoParam) (authInfo AuthInfo, err error) {
	return f.OpenIDConnectGetAuthInfo(r, param)
}

func (f *AppleImpl) OpenIDConnectGetAuthInfo(r OAuthAuthorizationResponse, param GetAuthInfoParam) (authInfo AuthInfo, err error) {
	keySet, err := appleOIDCConfig.FetchJWKs(http.DefaultClient)
	if err != nil {
		return
	}

	clientSecret, err := f.createClientSecret()
	if err != nil {
		return
	}

	var tokenResp AccessTokenResp
	jwtToken, err := appleOIDCConfig.ExchangeCode(
		http.DefaultClient,
		f.Clock,
		r.Code,
		keySet,
		f.ProviderConfig.ClientID,
		clientSecret,
		f.RedirectURL.SSOCallbackURL(f.ProviderConfig).String(),
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

	email, _ := claims["email"].(string)
	if email != "" {
		normalizer := f.LoginIDNormalizerFactory.NormalizerWithLoginIDType(config.LoginIDKeyTypeEmail)
		email, err = normalizer.Normalize(email)
		if err != nil {
			return
		}
	}

	authInfo.ProviderConfig = f.ProviderConfig
	authInfo.ProviderRawProfile = claims
	authInfo.ProviderAccessTokenResp = tokenResp
	authInfo.ProviderUserInfo = ProviderUserInfo{
		ID:    sub,
		Email: email,
	}

	return
}

func (f *AppleImpl) GetPrompt(prompt []string) []string {
	// apple doesn't support prompt parameter
	// see "Send the Required Query Parameters" section for supporting parameters
	// ref: https://developer.apple.com/documentation/sign_in_with_apple/sign_in_with_apple_js/incorporating_sign_in_with_apple_into_other_platforms
	return []string{}
}

var (
	_ OAuthProvider         = &AppleImpl{}
	_ OpenIDConnectProvider = &AppleImpl{}
)
