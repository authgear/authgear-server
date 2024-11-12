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

func NewExternalClientWithOptions(timeout time.Duration, opts ExternalClientOptions) *http.Client {
	// SECURITY(http): prevent SSRF

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
