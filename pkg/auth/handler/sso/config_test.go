package sso

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/skygeario/skygear-server/pkg/core/config"
	. "github.com/skygeario/skygear-server/pkg/core/skytest"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	allowedCallbackURLs = []string{
		"http://localhost",
		"http://127.0.0.1",
	}
	sampleConfig = config.TenantConfiguration{
		AppConfig: &config.AppConfiguration{
			SSO: &config.SSOConfiguration{
				OAuth: &config.OAuthConfiguration{
					AllowedCallbackURLs: allowedCallbackURLs,
				},
			},
		},
	}
)

func TestConfigHandler(t *testing.T) {
	Convey("Test ConfigHandler", t, func() {
		Convey("should return tenant SSOSeting AllowedCallbackURLs", func() {
			r, _ := http.NewRequest("POST", "", nil)
			rw := httptest.NewRecorder()
			r = r.WithContext(config.WithTenantConfig(r.Context(), &sampleConfig))

			var testingHandler ConfigHandler
			reqHandler := testingHandler.NewHandler(r)
			reqHandler.ServeHTTP(rw, r)

			So(rw.Body.Bytes(), ShouldEqualJSON, `{
				"result": {
					"authorized_urls": [
						"http://localhost",
						"http://127.0.0.1"
					]
				}
			}`)
		})
	})
}
