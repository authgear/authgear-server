package siteadmin

import (
	"net/http"

	"github.com/iawaknahc/originmatcher"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/slice"
)

type CORSMatcher struct {
	CORSAllowedOrigins config.CORSAllowedOrigins
}

func (m *CORSMatcher) PrepareOriginMatcher(r *http.Request) (*originmatcher.T, error) {
	allowedOrigins := []string{r.Host}
	allowedOrigins = append(allowedOrigins, m.CORSAllowedOrigins.List()...)
	allowedOrigins = slice.Deduplicate(allowedOrigins)

	matcher, err := originmatcher.New(allowedOrigins)
	if err != nil {
		return nil, err
	}

	return matcher, nil
}
