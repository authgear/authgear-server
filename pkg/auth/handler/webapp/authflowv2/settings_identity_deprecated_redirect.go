package authflowv2

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureAuthflowV2SettingsIdentityDeprecatedRedirectRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("GET").
		WithPathPattern(AuthflowV2RouteSettingsIdentityDeprecated)
}

func ConfigureAuthflowV2SettingsIdentitiesDeprecatedRedirectRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("GET").
		WithPathPattern(AuthflowV2RouteSettingsIdentitiesDeprecated)
}

// AuthflowV2SettingsIdentityDeprecatedRedirectHandler redirects the removed
// /settings/identity and /settings/identities pages to /settings, so that
// SDKs and bookmarks linking to the old paths keep working.
type AuthflowV2SettingsIdentityDeprecatedRedirectHandler struct{}

func (h *AuthflowV2SettingsIdentityDeprecatedRedirectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, SettingsV2RouteSettings, http.StatusFound)
}
