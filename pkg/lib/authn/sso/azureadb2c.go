package sso

import (
	"context"
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/oauthrelyingparty"
	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
	"github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/azureadb2c"
	"github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/oauthrelyingpartyutil"
)

type Azureadb2cImpl struct{}

func (f *Azureadb2cImpl) getOpenIDConfiguration(deps oauthrelyingparty.Dependencies) (*oauthrelyingpartyutil.OIDCDiscoveryDocument, error) {
	azureadb2cConfig := azureadb2c.ProviderConfig(deps.ProviderConfig)
	tenant := azureadb2cConfig.Tenant()
	policy := azureadb2cConfig.Policy()

	endpoint := fmt.Sprintf(
		"https://%s.b2clogin.com/%s.onmicrosoft.com/%s/v2.0/.well-known/openid-configuration",
		tenant,
		tenant,
		policy,
	)

	return oauthrelyingpartyutil.FetchOIDCDiscoveryDocument(deps.HTTPClient, endpoint)
}

func (f *Azureadb2cImpl) GetAuthorizationURL(deps oauthrelyingparty.Dependencies, param oauthrelyingparty.GetAuthorizationURLOptions) (string, error) {
	c, err := f.getOpenIDConfiguration(deps)
	if err != nil {
		return "", err
	}
	return c.MakeOAuthURL(oauthrelyingpartyutil.AuthorizationURLParams{
		ClientID:     deps.ProviderConfig.ClientID(),
		RedirectURI:  param.RedirectURI,
		Scope:        deps.ProviderConfig.Scope(),
		ResponseType: oauthrelyingparty.ResponseTypeCode,
		ResponseMode: param.ResponseMode,
		State:        param.State,
		Prompt:       f.getPrompt(param.Prompt),
		Nonce:        param.Nonce,
	}), nil
}

func (f *Azureadb2cImpl) GetUserProfile(deps oauthrelyingparty.Dependencies, param oauthrelyingparty.GetUserProfileOptions) (authInfo oauthrelyingparty.UserProfile, err error) {
	c, err := f.getOpenIDConfiguration(deps)
	if err != nil {
		return
	}
	// OPTIMIZE(sso): Cache JWKs
	keySet, err := c.FetchJWKs(deps.HTTPClient)
	if err != nil {
		return
	}

	var tokenResp oauthrelyingpartyutil.AccessTokenResp
	jwtToken, err := c.ExchangeCode(
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

	iss, ok := claims["iss"].(string)
	if !ok {
		err = OAuthProtocolError.New("iss not found in ID token")
		return
	}
	if iss != c.Issuer {
		err = OAuthProtocolError.New(
			fmt.Sprintf("iss: %v != %v", iss, c.Issuer),
		)
		return
	}

	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		err = OAuthProtocolError.New("sub not found in ID Token")
		return
	}

	authInfo.ProviderRawProfile = claims
	authInfo.ProviderUserID = sub

	stdAttrs, err := f.Extract(deps, claims)
	if err != nil {
		return
	}
	authInfo.StandardAttributes = stdAttrs

	return
}

func (f *Azureadb2cImpl) Extract(deps oauthrelyingparty.Dependencies, claims map[string]interface{}) (stdattrs.T, error) {
	// Here is the list of possible builtin claims of user flows
	// https://learn.microsoft.com/en-us/azure/active-directory-b2c/user-flow-overview#user-flows
	// city: free text
	// country: free text
	// jobTitle: free text
	// legalAgeGroupClassification: a enum with undocumented variants
	// postalCode: free text
	// state: free text
	// streetAddress: free text
	// newUser: true means the user signed up newly
	// oid: sub is identical to it by default.
	// emails: if non-empty, the first value corresponds to standard claim
	// name: correspond to standard claim
	// given_name: correspond to standard claim
	// family_name: correspond to standard claim

	// For custom policy we further recognize the following claims.
	// https://learn.microsoft.com/en-us/azure/active-directory-b2c/user-profile-attributes
	// signInNames.emailAddress: string

	extractString := func(input map[string]interface{}, output stdattrs.T, key string) {
		if value, ok := input[key].(string); ok && value != "" {
			output[key] = value
		}
	}

	out := stdattrs.T{}

	extractString(claims, out, stdattrs.Name)
	extractString(claims, out, stdattrs.GivenName)
	extractString(claims, out, stdattrs.FamilyName)

	var email string
	if email == "" {
		if ifaceSlice, ok := claims["emails"].([]interface{}); ok {
			for _, iface := range ifaceSlice {
				if str, ok := iface.(string); ok && str != "" {
					email = str
				}
			}
		}
	}
	if email == "" {
		if str, ok := claims["signInNames.emailAddress"].(string); ok {
			if str != "" {
				email = str
			}
		}
	}
	out[stdattrs.Email] = email

	emailRequired := deps.ProviderConfig.EmailClaimConfig().Required()
	return stdattrs.Extract(out, stdattrs.ExtractOptions{
		EmailRequired: emailRequired,
	})
}

func (f *Azureadb2cImpl) getPrompt(prompt []string) []string {
	// The only supported value is login.
	// See https://docs.microsoft.com/en-us/azure/active-directory-b2c/openid-connect
	for _, p := range prompt {
		if p == "login" {
			return []string{"login"}
		}
	}
	return []string{}
}

var (
	_ OAuthProvider = &Azureadb2cImpl{}
)
