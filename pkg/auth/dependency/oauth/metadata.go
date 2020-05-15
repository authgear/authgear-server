package oauth

type MetadataProvider struct {
	AuthorizeEndpoint AuthorizeEndpointProvider
	TokenEndpoint     TokenEndpointProvider
	RevokeEndpoint    RevokeEndpointProvider
}

func (p *MetadataProvider) PopulateMetadata(meta map[string]interface{}) {
	meta["authorization_endpoint"] = p.AuthorizeEndpoint.AuthorizeEndpointURI().String()
	meta["token_endpoint"] = p.TokenEndpoint.TokenEndpointURI().String()
	meta["response_types_supported"] = []string{"code", "none"}
	meta["grant_types_supported"] = []string{"authorization_code", "refresh_token"}
	meta["code_challenge_methods_supported"] = []string{"S256"}
	meta["revocation_endpoint"] = p.RevokeEndpoint.RevokeEndpointURI().String()
}
