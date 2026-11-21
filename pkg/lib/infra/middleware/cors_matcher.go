package middleware

import (
	"net/http"

	"github.com/iawaknahc/originmatcher"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/slice"
)

type CORSMatcher struct {
	Config             *config.HTTPConfig
	OAuthConfig        *config.OAuthConfig
	UIConfig           *config.UIConfig
	SAMLConfig         *config.SAMLConfig
	CORSAllowedOrigins config.CORSAllowedOrigins
}

func (m *CORSMatcher) PrepareOriginMatcher(r *http.Request) (*originmatcher.T, error) {
	// The host header is always allowed.
	allowedOrigins := []string{r.Host}

	// Allow the allowed_origins.
	allowedOrigins = append(allowedOrigins, m.Config.AllowedOrigins...)

	// Allow the origins in environment variable.
	allowedOrigins = append(allowedOrigins, m.CORSAllowedOrigins.List()...)

	// Allow the origins listed in redirect_uris, x_pre_authenticated_url_allowed_origins.
	for _, oauthClient := range m.OAuthConfig.Clients {
		allowedOrigins = append(allowedOrigins, oauthClient.RedirectURIs...)
		allowedOrigins = append(allowedOrigins, oauthClient.PreAuthenticatedURLAllowedOrigins...)
	}

	// Allow the origins of the Custom UI pages.
	allowedOrigins = append(allowedOrigins, config.CustomUIAllURIs(m.UIConfig, m.OAuthConfig)...)

	// Allow the origins listed in acs_urls, slo_callback_url
	for _, samlSP := range m.SAMLConfig.ServiceProviders {
		allowedOrigins = append(allowedOrigins, samlSP.AcsURLs...)
		allowedOrigins = append(allowedOrigins, samlSP.SLOCallbackURL)
	}

	allowedOrigins = slice.Deduplicate(allowedOrigins)

	matcher, err := originmatcher.New(allowedOrigins)
	if err != nil {
		return nil, err
	}

	return matcher, nil
}
