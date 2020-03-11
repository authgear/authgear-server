package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/core/config"
	. "github.com/skygeario/skygear-server/pkg/core/skytest"
)

var (
	sampleConfig = config.TenantConfiguration{
		AppName: "AppName",
		DatabaseConfig: &config.DatabaseConfiguration{
			DatabaseURL: "DBConnectionStr",
		},
		AppConfig: &config.AppConfiguration{
			Clients: []config.OAuthClientConfiguration{
				config.OAuthClientConfiguration{
					"client_id": "web-client-id",
				},
				config.OAuthClientConfiguration{
					"client_id": "mobile-client-id",
				},
			},
			MasterKey: "MasterKey",
		},
	}
)

func provideConfiguration(r *http.Request) (config.TenantConfiguration, error) {
	return sampleConfig, nil
}

func provideErr(r *http.Request) (config.TenantConfiguration, error) {
	return sampleConfig, http.ErrNotSupported
}

// GetTestHandler returns a http.HandlerFunc for testing http middleware
func GetTestHandler() http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})
}

func TestMiddleware(t *testing.T) {
	newReq := func() (req *http.Request) {
		req, _ = http.NewRequest("POST", "", nil)
		return
	}

	targetMiddleware := WriteTenantConfigMiddleware{
		ConfigurationProvider: ConfigurationProviderFunc(provideConfiguration),
	}
	var cb func(*http.Request)
	handler := targetMiddleware.Handle(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		cb(req)
	}))

	Convey("Test WriteTenantConfigMiddleware", t, func() {
		Convey("should handle request without headers", func() {
			req := newReq()
			cb = func(req *http.Request) {
				// NOTE(louis): msgp v1.1.0 serialize nil empty into empty map
				// so using ShouldResemble will fail the test.
				So(*config.GetTenantConfig(req.Context()), ShouldNonRecursiveDataDeepEqual, sampleConfig)
			}
			handler.ServeHTTP(nil, req)
		})

		targetErrMiddleware := WriteTenantConfigMiddleware{
			ConfigurationProvider: ConfigurationProviderFunc(provideErr),
		}
		errHandler := targetErrMiddleware.Handle(handler)

		Convey("should handle request with error config provider", func() {
			defer func() {
				r := recover()
				err, _ := r.(error)
				So(err.Error(), ShouldEqual, "unable to retrieve configuration: feature not supported")
			}()

			req := newReq()
			resp := httptest.NewRecorder()
			errHandler.ServeHTTP(resp, req)
		})
	})
}
