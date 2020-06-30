package oidc

type MetadataProvider struct {
	Endpoints EndpointsProvider
}

func (p *MetadataProvider) PopulateMetadata(meta map[string]interface{}) {
	meta["issuer"] = p.Endpoints.BaseURL().String()
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
	meta["jwks_uri"] = p.Endpoints.JWKSEndpointURL().String()
	meta["userinfo_endpoint"] = p.Endpoints.UserInfoEndpointURL().String()
	meta["end_session_endpoint"] = p.Endpoints.EndSessionEndpointURL().String()
}
