package modifier

import (
	"net/http"
	"net/url"

	"github.com/google/martian/parse"

	"github.com/authgear/authgear-server/e2e/cmd/proxy/mockoidc"
)

func init() {
	parse.Register("oidcModifier", oidcModifierFromJSON)
}

type OIDCModifier struct {
	Manager *mockoidc.MockOIDCManager
}

type EndpointType int

const (
	DiscoveryEndpoint EndpointType = iota
	AuthorizationEndpoint
	TokenEndpoint
	UserinfoEndpoint
	JWKSEndpoint
)

func (m *OIDCModifier) ModifyRequest(req *http.Request) error {
	var endpointType EndpointType
	var provider *mockoidc.Provider

	reqURL := *req.URL
	reqURL.RawQuery = ""
	reqURLString := reqURL.String()

	for _, p := range mockoidc.SupportedProviders {
		switch reqURLString {
		case p.DiscoveryEndpoint:
			provider = &p
			endpointType = DiscoveryEndpoint
		case p.AuthorizationEndpoint:
			provider = &p
			endpointType = AuthorizationEndpoint
		case p.TokenEndpoint:
			provider = &p
			endpointType = TokenEndpoint
		case p.UserinfoEndpoint:
			provider = &p
			endpointType = UserinfoEndpoint
		case p.JWKSEndpoint:
			provider = &p
			endpointType = JWKSEndpoint
		}

		if provider != nil {
			break
		}
	}

	if provider == nil {
		return nil
	}

	oidc := m.Manager.GetOIDC(provider.Type)
	if oidc == nil {
		return nil
	}

	oidcUrl, err := url.Parse(oidc.Addr)
	if err != nil {
		return err
	}

	// Modify the request to point to the mock OIDC server
	req.Host = oidcUrl.Host
	req.URL.Host = oidcUrl.Host
	req.URL.Scheme = "http"

	switch endpointType {
	case DiscoveryEndpoint:
		req.URL.Path = oidc.DiscoveryEndpoint()
	case AuthorizationEndpoint:
		req.URL.Path = oidc.AuthorizationEndpoint()
	case TokenEndpoint:
		req.URL.Path = oidc.TokenEndpoint()
	case UserinfoEndpoint:
		req.URL.Path = oidc.UserinfoEndpoint()
	case JWKSEndpoint:
		req.URL.Path = oidc.JWKSEndpoint()
	}

	return nil
}

// oidcModifierFromJSON constructs an OAuthMockModifier from JSON.
func oidcModifierFromJSON(b []byte) (*parse.Result, error) {
	modifier := &OIDCModifier{}
	return parse.NewResult(modifier, []parse.ModifierType{parse.Response})
}
