package handler

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/lib/otelauthgear"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type oauthRequest interface {
	ClientID() string
	RedirectURI() string
}

type OAuthClientResolver interface {
	ResolveClient(ctx context.Context, clientID string) *config.OAuthClientConfig
}

func resolveClient(ctx context.Context, resolver OAuthClientResolver, clientID string) (context.Context, *config.OAuthClientConfig) {
	client := resolver.ResolveClient(ctx, clientID)
	if client != nil {
		otelauthgear.SetClientID(ctx, clientID)
	}
	return ctx, client
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

// isLoopbackRedirectURI reports whether u is a loopback redirect URI in the
// sense of RFC 8252 §7.3, that is, the "http" scheme on a loopback host
// (httputil.IsLoopbackHost -- the same set pkg/lib/cimd and pkg/lib/dcr
// accept when validating registered URIs). url.Parse already lowercases the
// scheme; the host is lowercased here since RFC 3986 §3.2.2 makes it
// case-insensitive.
func isLoopbackRedirectURI(u *url.URL) bool {
	if u.Scheme != "http" {
		return false
	}
	return httputil.IsLoopbackHost(strings.ToLower(u.Hostname()))
}

// matchLoopbackRedirectURI implements RFC 8252 §7.3, which says the
// authorization server "MUST allow any port to be specified at the time of the
// request for loopback IP redirect URIs, to accommodate clients that obtain an
// available ephemeral port from the operating system at the time of the
// request". Native and CLI apps bind an OS-assigned ephemeral port, so the port
// of their redirect URI is not known when the client is registered.
//
// The exception applies only when both the registered URI and the incoming URI
// are loopback URIs. The port is then the only component allowed to differ:
// scheme, host, path, query and fragment must still match. Everything else
// keeps requiring strict equality, so a client that registers no loopback URI
// sees no change in behaviour.
//
// Confining the exception to the "http" scheme on a loopback host keeps it
// safe: such a redirect never leaves the end-user's machine, so an arbitrary
// port cannot be turned into a cross-origin redirection to an attacker. PKCE
// continues to protect the code against other local processes.
func matchLoopbackRedirectURI(allowedURIString string, redirectURI *url.URL) bool {
	allowedURI, err := url.Parse(allowedURIString)
	if err != nil {
		return false
	}
	if !isLoopbackRedirectURI(allowedURI) || !isLoopbackRedirectURI(redirectURI) {
		return false
	}
	return redirectURIWithoutPort(allowedURI) == redirectURIWithoutPort(redirectURI)
}

// redirectURIWithoutPort serializes u with the port component removed, so that
// two loopback URIs can be compared on every component except the port.
func redirectURIWithoutPort(u *url.URL) string {
	withoutPort := *u
	withoutPort.Host = strings.ToLower(u.Hostname())
	return withoutPort.String()
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

	for _, allowedURIString := range client.RedirectURIs {
		if allowedURIString == redirectURIString {
			allowed = true
			break
		}
		if matchLoopbackRedirectURI(allowedURIString, redirectURI) {
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
