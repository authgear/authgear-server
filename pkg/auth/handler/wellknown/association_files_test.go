package wellknown

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

const testFingerprint = "8B:BF:39:60:61:89:30:A4:45:F3:D7:09:1E:7B:1B:05:0F:8A:FD:AF:24:EB:F1:EB:2E:3D:13:88:09:FC:79:59"

func TestAssociationFiles(t *testing.T) {
	Convey("Association files", t, func() {
		ios := func(teamID string, bundleID string) *config.OAuthClientNativeAppConfig {
			return &config.OAuthClientNativeAppConfig{Platform: config.OAuthClientNativeAppPlatformIOS, TeamID: teamID, BundleID: bundleID}
		}
		android := func(packageName string, fingerprints ...string) *config.OAuthClientNativeAppConfig {
			return &config.OAuthClientNativeAppConfig{Platform: config.OAuthClientNativeAppPlatformAndroid, PackageName: packageName, SHA256CertFingerprints: fingerprints}
		}
		suffixes := config.AppHostSuffixes{".authgear.cloud"}
		serveAt := func(h http.Handler, host string) *httptest.ResponseRecorder {
			w := httptest.NewRecorder()
			h.ServeHTTP(w, httptest.NewRequest("GET", "https://"+host+"/", nil))
			return w
		}
		serve := func(h http.Handler) *httptest.ResponseRecorder {
			return serveAt(h, "auth.example.com")
		}
		aasa := func(suffixes config.AppHostSuffixes, oauthConfig *config.OAuthConfig) http.Handler {
			return &AppleAppSiteAssociationHandler{AppHostSuffixes: suffixes, OAuthConfig: oauthConfig}
		}
		assetLinks := func(suffixes config.AppHostSuffixes, oauthConfig *config.OAuthConfig) http.Handler {
			return &AssetLinksHandler{AppHostSuffixes: suffixes, OAuthConfig: oauthConfig}
		}
		twoClients := &config.OAuthConfig{
			Clients: []config.OAuthClientConfig{
				{ClientID: "a", NativeApps: []*config.OAuthClientNativeAppConfig{ios("ABCDE12345", "com.example.a"), android("com.example.a", testFingerprint)}},
				{ClientID: "b", NativeApps: []*config.OAuthClientNativeAppConfig{ios("ABCDE12345", "com.example.b"), android("com.example.b", "8b:bf:39:60:61:89:30:a4:45:f3:d7:09:1e:7b:1b:05:0f:8a:fd:af:24:eb:f1:eb:2e:3d:13:88:09:fc:79:59")}},
			},
		}

		Convey("apple-app-site-association", func() {
			Convey("lists every iOS app across clients", func() {
				w := serve(aasa(suffixes, twoClients))
				So(w.Code, ShouldEqual, http.StatusOK)
				So(w.Header().Get("Content-Type"), ShouldEqual, "application/json")
				So(w.Body.String(), ShouldEqualJSON, `{
					"webcredentials": {
						"apps": ["ABCDE12345.com.example.a", "ABCDE12345.com.example.b"]
					}
				}`)
			})

			Convey("returns 404 without iOS apps", func() {
				oauthConfig := &config.OAuthConfig{Clients: []config.OAuthClientConfig{
					{ClientID: "a", NativeApps: []*config.OAuthClientNativeAppConfig{android("com.example.a", testFingerprint)}},
				}}
				w := serve(aasa(suffixes, oauthConfig))
				So(w.Code, ShouldEqual, http.StatusNotFound)
			})
		})

		Convey("on a default domain", func() {
			for name, h := range map[string]func(config.AppHostSuffixes, *config.OAuthConfig) http.Handler{
				"apple-app-site-association": aasa,
				"assetlinks.json":            assetLinks,
			} {
				Convey(name+" returns 404 on a default domain", func() {
					w := serveAt(h(suffixes, twoClients), "myapp.authgear.cloud")
					So(w.Code, ShouldEqual, http.StatusNotFound)
				})

				Convey(name+" is served when no default domain is configured", func() {
					w := serveAt(h(nil, twoClients), "myapp.authgear.cloud")
					So(w.Code, ShouldEqual, http.StatusOK)
				})
			}
		})

		Convey("assetlinks.json", func() {
			Convey("lists every Android app across clients with both relations and upper-cased fingerprints", func() {
				w := serve(assetLinks(suffixes, twoClients))
				So(w.Code, ShouldEqual, http.StatusOK)
				So(w.Header().Get("Content-Type"), ShouldEqual, "application/json")
				So(w.Body.String(), ShouldEqualJSON, `[
					{
						"relation": [
							"delegate_permission/common.handle_all_urls",
							"delegate_permission/common.get_login_creds"
						],
						"target": {
							"namespace": "android_app",
							"package_name": "com.example.a",
							"sha256_cert_fingerprints": ["`+testFingerprint+`"]
						}
					},
					{
						"relation": [
							"delegate_permission/common.handle_all_urls",
							"delegate_permission/common.get_login_creds"
						],
						"target": {
							"namespace": "android_app",
							"package_name": "com.example.b",
							"sha256_cert_fingerprints": ["`+testFingerprint+`"]
						}
					}
				]`)
			})

			Convey("returns 404 without Android apps", func() {
				oauthConfig := &config.OAuthConfig{Clients: []config.OAuthClientConfig{
					{ClientID: "a", NativeApps: []*config.OAuthClientNativeAppConfig{ios("ABCDE12345", "com.example.a")}},
					{ClientID: "b"},
				}}
				w := serve(assetLinks(suffixes, oauthConfig))
				So(w.Code, ShouldEqual, http.StatusNotFound)
			})
		})
	})
}
