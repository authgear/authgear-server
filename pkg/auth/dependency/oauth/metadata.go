package oauth

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/urlprefix"
)

type MetadataProvider struct {
	URLPrefix            urlprefix.Provider
	AuthorizeEndpoint    AuthorizeEndpointProvider
	TokenEndpoint        TokenEndpointProvider
	AuthenticateEndpoint AuthenticateEndpointProvider
}

func (p *MetadataProvider) PopulateMetadata(meta map[string]interface{}) {
	meta["issuer"] = p.URLPrefix.Value().String()
	meta["authorization_endpoint"] = p.AuthorizeEndpoint.AuthorizeEndpointURI().String()
	meta["token_endpoint"] = p.TokenEndpoint.TokenEndpointURI().String()
	meta["response_types_supported"] = []string{"code", "none"}
	meta["grant_types_supported"] = []string{"authorization_code", "refresh_token"}
	meta["code_challenge_methods_supported"] = []string{"S256"}
}
