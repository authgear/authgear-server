package admin

import (
	"net/url"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

// FIXME(admin): Refactor to remove the need of this placeholder provider
//     Admin API is unable to trigger code paths that leads to these functions.
//     Implementation of these functions require access to web-app logic.
type WebEndpoints struct {
}

func (WebEndpoints) BaseURL() *url.URL {
	panic("not implemented")
}

func (WebEndpoints) VerifyIdentityURL(code string, webStateID string) *url.URL {
	panic("not implemented")
}

func (WebEndpoints) ResetPasswordURL(code string) *url.URL {
	panic("not implemented")
}

func (WebEndpoints) SSOCallbackURL(providerConfig config.OAuthSSOProviderConfig) *url.URL {
	panic("not implemented")
}
