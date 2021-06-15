package oauth

type MetadataProvider struct {
	Endpoints EndpointsProvider
}

func (p *MetadataProvider) PopulateMetadata(meta map[string]interface{}) {
	meta["authorization_endpoint"] = p.Endpoints.AuthorizeEndpointURL().String()
	meta["token_endpoint"] = p.Endpoints.TokenEndpointURL().String()
	meta["response_types_supported"] = []string{"code", "none"}
	meta["response_modes_supported"] = []string{"query", "fragment", "form_post"}
	meta["grant_types_supported"] = []string{"authorization_code", "refresh_token"}
	meta["code_challenge_methods_supported"] = []string{"S256"}
	meta["revocation_endpoint"] = p.Endpoints.RevokeEndpointURL().String()
	// We do not support confidential client (i.e. client_secret) at the moment,
	// so we do not perform authentication at all at the token endpoint.
	meta["token_endpoint_auth_methods_supported"] = []string{}
}
