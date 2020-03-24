package oidc

import "github.com/skygeario/skygear-server/pkg/auth/dependency/urlprefix"

type MetadataProvider struct {
	URLPrefix        urlprefix.Provider
	JWKSEndpoint     JWKSEndpointProvider
	UserInfoEndpoint UserInfoEndpointProvider
}

func (p *MetadataProvider) PopulateMetadata(meta map[string]interface{}) {
	meta["issuer"] = p.URLPrefix.Value().String()
	meta["scopes_supported"] = AllowedScopes
	meta["subject_types_supported"] = []string{"public"}
	meta["id_token_signing_alg_values_supported"] = []string{"RS256"}
	meta["claims_supported"] = []string{
		"iss",
		"aud",
		"iat",
		"exp",
		"sub",
	}
	meta["jwks_uri"] = p.JWKSEndpoint.JWKSEndpointURI().String()
	meta["userinfo_endpoint"] = p.UserInfoEndpoint.UserInfoEndpointURI().String()
}
