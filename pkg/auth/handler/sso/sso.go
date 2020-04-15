package sso

import (
	"fmt"
	"net/url"
	"path"

	"github.com/skygeario/skygear-server/pkg/core/config"
)

// nolint: deadcode
/*
	@ID SSOProviderID
	@Parameter provider_id path
		ID of SSO provider
		@JSONSchema
			{ "type": "string" }
*/
type ssoProviderParameter string

func RedirectURIForAPI(urlPrefix *url.URL, providerConfig config.OAuthProviderConfiguration) string {
	u := *urlPrefix
	u.Path = path.Join(u.Path, fmt.Sprintf("_auth/sso/%s/auth_handler", url.PathEscape(providerConfig.ID)))
	return u.String()
}
