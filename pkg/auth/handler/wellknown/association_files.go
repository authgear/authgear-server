package wellknown

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/feature/passkey"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

// The association files list the apps in x_native_apps, so that the operating
// system allows them to perform passkey ceremonies for the public origin's
// host. They are served only on a custom domain; PublicOriginMiddleware has
// already redirected any host other than the public origin's.
// See docs/specs/passkey-webview.md.

func ConfigureAppleAppSiteAssociationRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("GET").
		WithPathPattern("/.well-known/apple-app-site-association")
}

func ConfigureAssetLinksRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("GET").
		WithPathPattern("/.well-known/assetlinks.json")
}

type AppleAppSiteAssociationHandler struct {
	AppHostSuffixes config.AppHostSuffixes
	TrustProxy      config.TrustProxy
	OAuthConfig     *config.OAuthConfig
}

type appleAppSiteAssociation struct {
	WebCredentials appleWebCredentials `json:"webcredentials"`
}

type appleWebCredentials struct {
	Apps []string `json:"apps"`
}

func (h *AppleAppSiteAssociationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.AppHostSuffixes.CheckIsDefaultDomain(httputil.GetHost(r, bool(h.TrustProxy))) {
		http.NotFound(w, r)
		return
	}

	var appIDs []string
	for _, app := range passkey.EffectiveNativeApps(h.OAuthConfig) {
		if app.Platform != config.OAuthClientNativeAppPlatformIOS {
			continue
		}
		appIDs = append(appIDs, app.TeamID+"."+app.BundleID)
	}
	if len(appIDs) == 0 {
		http.NotFound(w, r)
		return
	}

	writeJSON(w, appleAppSiteAssociation{
		WebCredentials: appleWebCredentials{Apps: appIDs},
	})
}

type AssetLinksHandler struct {
	AppHostSuffixes config.AppHostSuffixes
	TrustProxy      config.TrustProxy
	OAuthConfig     *config.OAuthConfig
}

type assetLinksStatement struct {
	Relation []string         `json:"relation"`
	Target   assetLinksTarget `json:"target"`
}

type assetLinksTarget struct {
	Namespace              string   `json:"namespace"`
	PackageName            string   `json:"package_name"`
	SHA256CertFingerprints []string `json:"sha256_cert_fingerprints"`
}

// assetLinksRelations are both required: Google Play services refuses a
// passkey sign-in with get_login_creds alone.
var assetLinksRelations = []string{
	"delegate_permission/common.handle_all_urls",
	"delegate_permission/common.get_login_creds",
}

func (h *AssetLinksHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.AppHostSuffixes.CheckIsDefaultDomain(httputil.GetHost(r, bool(h.TrustProxy))) {
		http.NotFound(w, r)
		return
	}

	var statements []assetLinksStatement
	for _, app := range passkey.EffectiveNativeApps(h.OAuthConfig) {
		if app.Platform != config.OAuthClientNativeAppPlatformAndroid {
			continue
		}
		fingerprints := make([]string, len(app.SHA256CertFingerprints))
		for i, fingerprint := range app.SHA256CertFingerprints {
			fingerprints[i] = strings.ToUpper(fingerprint)
		}
		statements = append(statements, assetLinksStatement{
			Relation: assetLinksRelations,
			Target: assetLinksTarget{
				Namespace:              "android_app",
				PackageName:            app.PackageName,
				SHA256CertFingerprints: fingerprints,
			},
		})
	}
	if len(statements) == 0 {
		http.NotFound(w, r)
		return
	}

	writeJSON(w, statements)
}

func writeJSON(w http.ResponseWriter, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(body)
}
