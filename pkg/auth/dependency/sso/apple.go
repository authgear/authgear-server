package sso

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"

	"github.com/skygeario/skygear-server/pkg/auth/config"
	"github.com/skygeario/skygear-server/pkg/clock"
	"github.com/skygeario/skygear-server/pkg/core/crypto"
	"github.com/skygeario/skygear-server/pkg/core/errors"
)

var appleOIDCConfig = OIDCDiscoveryDocument{
	JWKSUri:               "https://appleid.apple.com/auth/keys",
	TokenEndpoint:         "https://appleid.apple.com/auth/token",
	AuthorizationEndpoint: "https://appleid.apple.com/auth/authorize",
}

type AppleImpl struct {
	URLPrefix                *url.URL
	RedirectURLFunc          RedirectURLFunc
	ProviderConfig           config.OAuthSSOProviderConfig
	Clock                    clock.Clock
	LoginIDNormalizerFactory LoginIDNormalizerFactory
}

func (f *AppleImpl) createClientSecret() (clientSecret string, err error) {
	// https://developer.apple.com/documentation/signinwithapplerestapi/generate_and_validate_tokens
	// FIXME: retrieve client secret
	key, err := crypto.ParseAppleP8PrivateKey([]byte(""))
	if err != nil {
		return
	}

	now := f.Clock.NowUTC()
	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.StandardClaims{
		Issuer:    f.ProviderConfig.TeamID,
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(5 * time.Minute).Unix(),
		Audience:  "https://appleid.apple.com",
		Subject:   f.ProviderConfig.ClientID,
	})
	token.Header["kid"] = f.ProviderConfig.KeyID

	clientSecret, err = token.SignedString(key)
	if err != nil {
		return
	}

	return
}

func (f *AppleImpl) Type() config.OAuthSSOProviderType {
	return config.OAuthSSOProviderTypeApple
}

func (f *AppleImpl) GetAuthURL(state State, encodedState string) (string, error) {
	return appleOIDCConfig.MakeOAuthURL(OIDCAuthParams{
		ProviderConfig: f.ProviderConfig,
		RedirectURI:    f.RedirectURLFunc(f.URLPrefix, f.ProviderConfig),
		Nonce:          state.HashedNonce,
		EncodedState:   encodedState,
	}), nil
}

func (f *AppleImpl) GetAuthInfo(r OAuthAuthorizationResponse, state State) (authInfo AuthInfo, err error) {
	return f.OpenIDConnectGetAuthInfo(r, state)
}

func (f *AppleImpl) OpenIDConnectGetAuthInfo(r OAuthAuthorizationResponse, state State) (authInfo AuthInfo, err error) {
	keySet, err := appleOIDCConfig.FetchJWKs(http.DefaultClient)
	if err != nil {
		err = NewSSOFailed(NetworkFailed, "failed to get OIDC JWKs")
		return
	}

	clientSecret, err := f.createClientSecret()
	if err != nil {
		err = errors.Newf("failed to create client secret: %w", err)
		return
	}

	var tokenResp AccessTokenResp
	claims, err := appleOIDCConfig.ExchangeCode(
		http.DefaultClient,
		r.Code,
		keySet,
		f.URLPrefix,
		f.ProviderConfig.ClientID,
		clientSecret,
		f.RedirectURLFunc(f.URLPrefix, f.ProviderConfig),
		state.HashedNonce,
		f.Clock.NowUTC,
		&tokenResp,
	)
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
		err = NewSSOFailed(SSOUnauthorized, "invalid iss")
		return
	}
	if !strings.Contains(iss, "https://appleid.apple.com") {
		err = NewSSOFailed(SSOUnauthorized, "invalid iss")
		return
	}

	// Ensure sub exists
	sub, ok := claims["sub"].(string)
	if !ok {
		err = NewSSOFailed(SSOUnauthorized, "no sub")
		return
	}

	email, _ := claims["email"].(string)
	if email != "" {
		normalizer := f.LoginIDNormalizerFactory.NormalizerWithLoginIDType(config.LoginIDKeyType("email"))
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

var (
	_ OAuthProvider         = &AppleImpl{}
	_ OpenIDConnectProvider = &AppleImpl{}
)
