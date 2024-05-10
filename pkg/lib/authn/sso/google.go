package sso

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/oauthrelyingparty"
	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
	"github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/oauthrelyingpartyutil"
)

const (
	googleOIDCDiscoveryDocumentURL string = "https://accounts.google.com/.well-known/openid-configuration"
)

type GoogleImpl struct{}

func (f *GoogleImpl) GetAuthorizationURL(deps oauthrelyingparty.Dependencies, param oauthrelyingparty.GetAuthorizationURLOptions) (string, error) {
	d, err := FetchOIDCDiscoveryDocument(deps.HTTPClient, googleOIDCDiscoveryDocumentURL)
	if err != nil {
		return "", err
	}
	return d.MakeOAuthURL(oauthrelyingpartyutil.AuthorizationURLParams{
		ClientID:     deps.ProviderConfig.ClientID(),
		RedirectURI:  param.RedirectURI,
		Scope:        deps.ProviderConfig.Scope(),
		ResponseType: oauthrelyingparty.ResponseTypeCode,
		ResponseMode: param.ResponseMode,
		State:        param.State,
		Nonce:        param.Nonce,
		Prompt:       f.getPrompt(param.Prompt),
	}), nil
}

func (f *GoogleImpl) GetUserProfile(deps oauthrelyingparty.Dependencies, param oauthrelyingparty.GetUserProfileOptions) (authInfo oauthrelyingparty.UserProfile, err error) {
	d, err := FetchOIDCDiscoveryDocument(deps.HTTPClient, googleOIDCDiscoveryDocumentURL)
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

func (f *GoogleImpl) getPrompt(prompt []string) []string {
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

var (
	_ OAuthProvider = &GoogleImpl{}
)
