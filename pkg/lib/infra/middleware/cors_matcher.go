package middleware

import (
	"github.com/iawaknahc/originmatcher"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/slice"
)

type CORSMatcher struct {
	Config             *config.HTTPConfig
	OAuthConfig        *config.OAuthConfig
	CORSAllowedOrigins config.CORSAllowedOrigins
}

func (m *CORSMatcher) PrepareOriginMatcher() (*originmatcher.T, error) {
	allowedOrigins := m.Config.AllowedOrigins
	allowedOrigins = append(allowedOrigins, m.CORSAllowedOrigins.List()...)
	for _, oauthClient := range m.OAuthConfig.Clients {
		allowedOrigins = append(allowedOrigins, oauthClient.RedirectURIs...)
	}
	allowedOrigins = slice.Deduplicate(allowedOrigins)
	matcher, err := originmatcher.New(allowedOrigins)
	if err != nil {
		return nil, err
	}

	return matcher, nil
}
