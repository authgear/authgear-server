package oidc

type MetadataProvider struct{}

func (p *MetadataProvider) PopulateMetadata(meta map[string]interface{}) {
	meta["scopes_supported"] = AllowedScopes
	meta["subject_types_supported"] = []string{"public"}
	meta["id_token_signing_alg_values_supported"] = []string{"RS256"}
	// TODO(oauth): userinfo_endpoint
	// TODO(oauth): jwks_uri
	// TODO(oauth): revocation_endpoint
	// TODO(oauth): claims_supported
}
