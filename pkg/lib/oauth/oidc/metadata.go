package oidc

import (
	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
)

var claimsSupported []string

func init() {
	claimsSupported = append(
		claimsSupported,
		[]string{
			"iss",
			"aud",
			"iat",
			"exp",
			"sub",
		}...,
	)
	claimsSupported = append(
		claimsSupported,
		[]string{
			stdattrs.Email,
			stdattrs.EmailVerified,
			stdattrs.PhoneNumber,
			stdattrs.PhoneNumberVerified,
			stdattrs.PreferredUsername,
			stdattrs.FamilyName,
			stdattrs.GivenName,
			stdattrs.MiddleName,
			stdattrs.Name,
			stdattrs.Nickname,
			stdattrs.Picture,
			stdattrs.Profile,
			stdattrs.Website,
			stdattrs.Gender,
			stdattrs.Birthdate,
			stdattrs.Zoneinfo,
			stdattrs.Locale,
			stdattrs.Address,
			stdattrs.UpdatedAt,
		}...,
	)
}

type MetadataProvider struct {
	Endpoints EndpointsProvider
}

func (p *MetadataProvider) PopulateMetadata(meta map[string]interface{}) {
	meta["issuer"] = p.Endpoints.Origin().String()
	meta["scopes_supported"] = AllowedScopes
	meta["subject_types_supported"] = []string{"public"}
	meta["id_token_signing_alg_values_supported"] = []string{"RS256"}
	meta["claims_supported"] = claimsSupported
	meta["jwks_uri"] = p.Endpoints.JWKSEndpointURL().String()
	meta["userinfo_endpoint"] = p.Endpoints.UserInfoEndpointURL().String()
	meta["end_session_endpoint"] = p.Endpoints.EndSessionEndpointURL().String()
	// TODO(mfa): Declare acr_values_supported and support acr_values in authorization request.
}
