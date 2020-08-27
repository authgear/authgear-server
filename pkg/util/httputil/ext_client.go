package httputil

import (
	"net/http"
	"time"
)

type ExternalClientOptions struct {
	FollowRedirect bool
}

func NewExternalClient(timeout time.Duration) *http.Client {
	return NewExternalClientWithOptions(timeout, ExternalClientOptions{})
}

func NewExternalClientWithOptions(timeout time.Duration, opts ExternalClientOptions) *http.Client {
	// SECURITY(http): prevent SSRF
	client := &http.Client{}
	if !opts.FollowRedirect {
		client.CheckRedirect = noFollowRedirectPolicy
	}
	client.Timeout = timeout
	return client
}

func noFollowRedirectPolicy(*http.Request, []*http.Request) error {
	return http.ErrUseLastResponse
}
