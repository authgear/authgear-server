package sso

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"

	authtesting "github.com/skygeario/skygear-server/pkg/auth/dependency/auth/testing"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/urlprefix"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	coreTime "github.com/skygeario/skygear-server/pkg/core/time"

	. "github.com/skygeario/skygear-server/pkg/core/skytest"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUnlinkHandler(t *testing.T) {
	Convey("Test UnlinkHandler", t, func() {
		req, _ := http.NewRequest("POST", "https://api.example.com", nil)

		sh := &UnlinkHandler{}
		sh.TxContext = db.NewMockTxContext()
		timeProvider := &coreTime.MockProvider{}
		sh.ProviderFactory = sso.NewOAuthProviderFactory(config.TenantConfiguration{
			AppConfig: &config.AppConfiguration{
				Identity: &config.IdentityConfiguration{
					OAuth: &config.OAuthConfiguration{
						Providers: []config.OAuthProviderConfiguration{
							config.OAuthProviderConfiguration{
								Type: "google",
								ID:   "google",
							},
						},
					},
				},
			},
		}, urlprefix.NewProvider(req), timeProvider, nil, nil, RedirectURIForAPI)

		router := mux.NewRouter()
		router.Handle("/sso/{provider}/unlink", sh)

		Convey("should return unknown sso provider", func() {
			req, _ := http.NewRequest("POST", "/sso/unknown/unlink", strings.NewReader(`{
			}`))
			req = authtesting.WithAuthn().
				UserID("faseng.cat.id").
				ToRequest(req)
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, 404)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"error": {
					"code": 404,
					"message": "unknown SSO provider",
					"name": "NotFound",
					"reason": "NotFound"
				}
			}
			`)
		})
	})
}
