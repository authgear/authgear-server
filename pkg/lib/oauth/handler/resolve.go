package handler

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type oauthRequest interface {
	ClientID() string
	RedirectURI() string
}

type OAuthClientResolver interface {
	ResolveClient(clientID string) *config.OAuthClientConfig
}

func resolveClient(resolver OAuthClientResolver, r oauthRequest) *config.OAuthClientConfig {
	return resolver.ResolveClient(r.ClientID())
}

func parseRedirectURI(
	client *config.OAuthClientConfig,
	httpProto httputil.HTTPProto,
	httpOrigin httputil.HTTPOrigin,
	domainWhitelist []string,
	originWhitelist []string,
	r oauthRequest,
) (*url.URL, protocol.ErrorResponse) {
	allowedURIs := client.RedirectURIs
	redirectURIString := r.RedirectURI()
	if len(allowedURIs) == 1 && redirectURIString == "" {
		// Redirect URI is default to the only allowed URI if possible.
		redirectURIString = allowedURIs[0]
	}

	redirectURI, err := url.Parse(redirectURIString)
	if err != nil {
		return nil, protocol.NewErrorResponse("invalid_request", "invalid redirect URI")
	}

	err = validateRedirectURI(client, httpProto, httpOrigin, domainWhitelist, originWhitelist, redirectURI)
	if err != nil {
		return nil, protocol.NewErrorResponse("invalid_request", err.Error())
	}

	return redirectURI, nil
}

func validateRedirectURI(
	client *config.OAuthClientConfig,
	httpProto httputil.HTTPProto,
	httpOrigin httputil.HTTPOrigin,
	domainWhitelist []string,
	originWhitelist []string,
	redirectURI *url.URL,
) error {
	allowed := false
	redirectURIString := redirectURI.String()

	for _, u := range client.RedirectURIs {
		if u == redirectURIString {
			allowed = true
			break
		}
	}

	// Implicitly allow URIs at same origin as the AS.
	// NOTE: this is a willful violation of OAuth spec, since first-party apps
	//       would often want to open pages on AS using OAuth mechanism.
	redirectURIOrigin := fmt.Sprintf("%s://%s", redirectURI.Scheme, redirectURI.Host)
	if redirectURIOrigin == string(httpOrigin) {
		allowed = true
	}

	// Implicitly allow URIs at same origin as the custom ui uri.
	if client.CustomUIURI != "" {
		customUIURI, err := url.Parse(client.CustomUIURI)
		if err != nil {
			return errors.New("invalid custom ui URI")
		}
		customUIURIOrigin := fmt.Sprintf("%s://%s", customUIURI.Scheme, customUIURI.Host)
		if customUIURIOrigin == redirectURIOrigin {
			allowed = true
		}
	}

	// Implicitly allow URIs for all whitelisted domains in httpProto
	for _, domain := range domainWhitelist {
		origin := fmt.Sprintf("%s://%s", httpProto, domain)
		if redirectURIOrigin == string(origin) {
			allowed = true
		}
	}

	for _, originStr := range originWhitelist {
		originURL, parseErr := url.Parse(originStr)
		if parseErr != nil {
			continue
		}
		origin := fmt.Sprintf("%s://%s", originURL.Scheme, originURL.Host)
		if redirectURIOrigin == origin {
			allowed = true
		}
	}

	if !allowed {
		return errors.New("redirect URI is not allowed")
	}

	return nil
}
