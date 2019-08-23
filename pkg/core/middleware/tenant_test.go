package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/skygeario/skygear-server/pkg/core/config"
	coreHttp "github.com/skygeario/skygear-server/pkg/core/http"
	"github.com/skygeario/skygear-server/pkg/core/model"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	sampleConfig = config.TenantConfiguration{
		AppName: "AppName",
		AppConfig: config.AppConfiguration{
			DatabaseURL: "DBConnectionStr",
		},
		UserConfig: config.UserConfiguration{
			APIKey:    "APIKey",
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

	targetMiddleware := TenantConfigurationMiddleware{
		ConfigurationProvider: ConfigurationProviderFunc(provideConfiguration),
	}
	handler := targetMiddleware.Handle(GetTestHandler())

	Convey("Test TenantConfigurationMiddleware", t, func() {
		Convey("should handle request without headers", func() {
			req := newReq()
			handler.ServeHTTP(nil, req)
			So(model.GetAccessKeyType(req), ShouldEqual, model.NoAccessKey)
			So(config.GetTenantConfig(req), ShouldResemble, sampleConfig)
		})

		Convey("should handle request with apikey", func() {
			req := newReq()
			req.Header.Set(coreHttp.HeaderAPIKey, "APIKey")
			handler.ServeHTTP(nil, req)
			So(model.GetAccessKeyType(req), ShouldEqual, model.APIAccessKey)
			So(config.GetTenantConfig(req), ShouldResemble, sampleConfig)
		})

		Convey("should handle request with masterkey", func() {
			req := newReq()
			req.Header.Set(coreHttp.HeaderAPIKey, "MasterKey")
			handler.ServeHTTP(nil, req)
			So(model.GetAccessKeyType(req), ShouldEqual, model.MasterAccessKey)
			So(config.GetTenantConfig(req), ShouldResemble, sampleConfig)
		})

		targetErrMiddleware := TenantConfigurationMiddleware{
			ConfigurationProvider: ConfigurationProviderFunc(provideErr),
		}
		errHandler := targetErrMiddleware.Handle(GetTestHandler())

		Convey("should handle request with error config provider", func() {
			defer func() {
				r := recover()
				err, _ := r.(error)
				So(err.Error(), ShouldEqual, "Unable to retrieve configuration: feature not supported")
			}()

			req := newReq()
			resp := httptest.NewRecorder()
			errHandler.ServeHTTP(resp, req)
		})
	})
}
