package sso

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/oauthrelyingparty"
	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
	"github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/oauthrelyingpartyutil"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

const (
	googleOIDCDiscoveryDocumentURL string = "https://accounts.google.com/.well-known/openid-configuration"
)

type GoogleImpl struct {
	Clock                        clock.Clock
	ProviderConfig               oauthrelyingparty.ProviderConfig
	ClientSecret                 string
	StandardAttributesNormalizer StandardAttributesNormalizer
	HTTPClient                   OAuthHTTPClient
}

func (f *GoogleImpl) GetAuthURL(param GetAuthURLParam) (string, error) {
	d, err := FetchOIDCDiscoveryDocument(f.HTTPClient, googleOIDCDiscoveryDocumentURL)
	if err != nil {
		return "", err
	}
	return d.MakeOAuthURL(oauthrelyingpartyutil.AuthorizationURLParams{
		ClientID:     f.ProviderConfig.ClientID(),
		RedirectURI:  param.RedirectURI,
		Scope:        f.ProviderConfig.Scope(),
		ResponseType: oauthrelyingparty.ResponseTypeCode,
		ResponseMode: param.ResponseMode,
		State:        param.State,
		Nonce:        param.Nonce,
		Prompt:       f.GetPrompt(param.Prompt),
	}), nil
}

func (f *GoogleImpl) Config() oauthrelyingparty.ProviderConfig {
	return f.ProviderConfig
}

func (f *GoogleImpl) GetAuthInfo(r OAuthAuthorizationResponse, param GetAuthInfoParam) (authInfo AuthInfo, err error) {
	d, err := FetchOIDCDiscoveryDocument(f.HTTPClient, googleOIDCDiscoveryDocumentURL)
	if err != nil {
		return
	}
	// OPTIMIZE(sso): Cache JWKs
	keySet, err := d.FetchJWKs(f.HTTPClient)
	if err != nil {
		return
	}

	var tokenResp oauthrelyingpartyutil.AccessTokenResp
	jwtToken, err := d.ExchangeCode(
		f.HTTPClient,
		f.Clock,
		r.Code,
		keySet,
		f.ProviderConfig.ClientID(),
		f.ClientSecret,
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
		err = OAuthProtocolError.New("iss not found in ID token")
		return
	}
	if iss != "https://accounts.google.com" && iss != "accounts.google.com" {
		err = OAuthProtocolError.New("iss is not from Google")
		return
	}

	// Ensure sub exists
	sub, ok := claims["sub"].(string)
	if !ok {
		err = OAuthProtocolError.New("sub not found in ID token")
		return
	}

	authInfo.ProviderRawProfile = claims
	authInfo.ProviderUserID = sub
	// Google supports
	// given_name, family_name, email, picture, profile, locale
	// https://developers.google.com/identity/protocols/oauth2/openid-connect#obtainuserinfo
	emailRequired := f.ProviderConfig.EmailClaimConfig().Required()
	stdAttrs, err := stdattrs.Extract(claims, stdattrs.ExtractOptions{
		EmailRequired: emailRequired,
	})
	if err != nil {
		return
	}
	authInfo.StandardAttributes = stdAttrs

	err = f.StandardAttributesNormalizer.Normalize(authInfo.StandardAttributes)
	if err != nil {
		return
	}

	return
}

func (f *GoogleImpl) GetPrompt(prompt []string) []string {
	// google support `none`, `concent` and `select_account` for prompt
	// the usage of `none` is for checking existing authentication and/or consent
	// which doesn't fit auth ui case
	// ref: https://developers.google.com/identity/protocols/oauth2/openid-connect#authenticationuriparameters
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

var (
	_ OAuthProvider = &GoogleImpl{}
)
