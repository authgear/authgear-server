package handler

import (
	"net/url"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
)

type oauthRequest interface {
	ClientID() string
	RedirectURI() string
}

func resolveClient(config *config.OAuthConfig, r oauthRequest) config.OAuthClientConfig {
	if client, ok := config.GetClient(r.ClientID()); ok {
		return client
	}
	return nil
}

func parseRedirectURI(client config.OAuthClientConfig, r oauthRequest) (*url.URL, protocol.ErrorResponse) {
	allowedURIs := client.RedirectURIs()
	redirectURIString := r.RedirectURI()
	if len(allowedURIs) == 1 && redirectURIString == "" {
		// Redirect URI is default to the only allowed URI if possible.
		redirectURIString = allowedURIs[0]
	}

	redirectURI, err := url.Parse(redirectURIString)
	if err != nil {
		return nil, protocol.NewErrorResponse("invalid_request", "invalid redirect URI")
	}

	allowed := false
	for _, u := range allowedURIs {
		if u == redirectURIString {
			allowed = true
			break
		}
	}
	if !allowed {
		return nil, protocol.NewErrorResponse("invalid_request", "redirect URI is not allowed")
	}

	return redirectURI, nil
}
