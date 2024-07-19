package oauth

import (
	"github.com/authgear/authgear-server/pkg/lib/dpop"
	"github.com/authgear/authgear-server/pkg/util/pkce"
)

type MetadataProvider struct {
	Endpoints EndpointsProvider
}

func (p *MetadataProvider) PopulateMetadata(meta map[string]interface{}) {
	meta["authorization_endpoint"] = p.Endpoints.AuthorizeEndpointURL().String()
	meta["token_endpoint"] = p.Endpoints.TokenEndpointURL().String()
	meta["response_types_supported"] = []string{"code", "urn:authgear:params:oauth:response-type:settings-action", "none"}
	meta["response_modes_supported"] = []string{"query", "fragment", "form_post"}
	meta["grant_types_supported"] = []string{"authorization_code", "refresh_token"}
	meta["code_challenge_methods_supported"] = []string{pkce.CodeChallengeMethodS256}
	meta["revocation_endpoint"] = p.Endpoints.RevokeEndpointURL().String()
	// The default is client_secret_basic if this key is omitted.
	// See https://openid.net/specs/openid-connect-discovery-1_0.html#:~:text=passed%20by%20reference.-,token_endpoint_auth_methods_supported,-OPTIONAL.%20JSON%20array
	// See https://openid.net/specs/openid-connect-core-1_0.html#ClientAuthentication:~:text=The%20Client%20does%20not%20authenticate%20itself%20at%20the%20Token%20Endpoint
	meta["token_endpoint_auth_methods_supported"] = []string{"none", "client_secret_post"}
	meta["dpop_signing_alg_values_supported"] = dpop.SupportedAlgorithms
}
