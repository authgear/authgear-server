package httputil

import (
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type ExternalClientOptions struct {
	FollowRedirect bool
	Transport      http.RoundTripper
}

func NewExternalClient(timeout time.Duration) *http.Client {
	return NewExternalClientWithOptions(timeout, ExternalClientOptions{})
}

// NewExternalClientWithOptions returns a client for a destination this
// deployment configures: the Deno hook runner, object storage, an SMS or
// captcha vendor's API. Those legitimately address internal hosts, so no
// address restriction applies here.
//
// For a destination taken from a project's configuration, use
// NewSSRFSafeExternalClient instead -- that is where the SSRF concern lives,
// and it cannot be handled here without breaking the call sites above.
func NewExternalClientWithOptions(timeout time.Duration, opts ExternalClientOptions) *http.Client {

	client := &http.Client{
		Timeout: timeout,
		// It is perfectly fine that Transport is nil.
		Transport: otelhttp.NewTransport(opts.Transport),
	}

	if !opts.FollowRedirect {
		client.CheckRedirect = noFollowRedirectPolicy
	}

	return client
}

func noFollowRedirectPolicy(*http.Request, []*http.Request) error {
	return http.ErrUseLastResponse
}

// SSRFSafeExternalClientOptions configures NewSSRFSafeExternalClient.
type SSRFSafeExternalClientOptions struct {
	// AllowNonPublicAddresses lifts the address restriction entirely. It
	// carries http.insecure_fetch_address_allowed.
	AllowNonPublicAddresses bool
	// AllowedHosts exempts named hosts from the restriction, carrying
	// http.insecure_fetch_address_allowed_hosts. Nothing else may widen the
	// policy.
	AllowedHosts []string
	// Sink names the configuration that chose the URL, e.g.
	// "hook.blocking_handlers". It appears in the log a refusal writes, so
	// that the message says which setting to go and fix.
	Sink string
}

// NewSSRFSafeExternalClient returns a client for fetching a URL whose
// destination comes from a project's configuration rather than from this
// deployment: event webhooks, the custom SMS provider, the account migration
// and phone verification hooks, and an OAuth provider's OIDC discovery
// document.
//
// Whoever administers a project chooses those URLs, and on a deployment with
// open sign-up that is any user who created one, so the destination is not
// trusted to be outside the deployment's own network. SafeDialer is what
// enforces that; see its documentation for why the check cannot live in a
// dialer Control hook alone.
//
// Not for the clients whose destination this deployment configures -- the Deno
// hook runner, object storage, an SMS vendor's API. Those legitimately address
// internal hosts, and must keep using NewExternalClient.
func NewSSRFSafeExternalClient(timeout time.Duration, opts SSRFSafeExternalClientOptions) *http.Client {
	dialer := &SafeDialer{
		AllowNonPublicAddresses: opts.AllowNonPublicAddresses,
		AllowedHosts:            opts.AllowedHosts,
		DialTimeout:             timeout,
		Sink:                    opts.Sink,
	}
	return NewExternalClientWithOptions(timeout, ExternalClientOptions{
		Transport: &http.Transport{
			DialContext:           dialer.DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConnsPerHost:   2,
			IdleConnTimeout:       30 * time.Second,
			TLSHandshakeTimeout:   timeout,
			ResponseHeaderTimeout: timeout,
			// No Proxy. http.ProxyFromEnvironment would route the request
			// through a proxy chosen by the environment, and the proxy --
			// not SafeDialer -- would then resolve the name, silently
			// bypassing every rule it enforces.
			Proxy: nil,
		},
	})
}
