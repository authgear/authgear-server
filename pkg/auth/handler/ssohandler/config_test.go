package ssohandler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/middleware"
	. "github.com/skygeario/skygear-server/pkg/server/skytest"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	allowedCallbackURLs = []string{
		"http://localhost",
		"http://127.0.0.1",
	}
	sampleConfig = config.TenantConfiguration{
		SSOSetting: config.SSOSetting{
			AllowedCallbackURLs: allowedCallbackURLs,
		},
	}
)

func provideConfiguration(r *http.Request) (config.TenantConfiguration, error) {
	return sampleConfig, nil
}

func TestConfigHandler(t *testing.T) {
	Convey("Test ConfigHandler", t, func() {
		targetMiddleware := middleware.TenantConfigurationMiddleware{
			ConfigurationProvider: middleware.ConfigurationProviderFunc(provideConfiguration),
		}

		Convey("should return tenant SSOSeting AllowedCallbackURLs", func() {
			r, _ := http.NewRequest("POST", "", nil)
			rw := httptest.NewRecorder()

			var testingHandler ConfigHandler
			reqHandler := targetMiddleware.Handle(testingHandler.NewHandler(r))
			reqHandler.ServeHTTP(rw, r)

			So(rw.Body.Bytes(), ShouldEqualJSON, `{
				"result": [
					"http://localhost",
					"http://127.0.0.1"
				]
			}`)
		})
	})
}
