package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/model"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	sampleConfig = config.TenantConfiguration{
		DBConnectionStr: "DBConnectionStr",
		APIKey:          "APIKey",
		MasterKey:       "MasterKey",
		AppName:         "AppName",
		TokenStore: config.TokenStoreConfiguration{
			Secret: "Secret",
			Expiry: 1000,
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

	Convey("handle request without headers", t, func() {
		req := newReq()
		handler.ServeHTTP(nil, req)
		So(model.GetAccessKeyType(req), ShouldEqual, model.NoAccessKey)
		So(config.GetTenantConfig(req), ShouldResemble, sampleConfig)
	})

	Convey("handle request with apikey", t, func() {
		req := newReq()
		req.Header.Set("X-Skygear-Api-Key", "APIKey")
		handler.ServeHTTP(nil, req)
		So(model.GetAccessKeyType(req), ShouldEqual, model.APIAccessKey)
		So(config.GetTenantConfig(req), ShouldResemble, sampleConfig)
	})

	Convey("handle request with masterkey", t, func() {
		req := newReq()
		req.Header.Set("X-Skygear-Api-Key", "MasterKey")
		handler.ServeHTTP(nil, req)
		So(model.GetAccessKeyType(req), ShouldEqual, model.MasterAccessKey)
		So(config.GetTenantConfig(req), ShouldResemble, sampleConfig)
	})

	targetErrMiddleware := TenantConfigurationMiddleware{
		ConfigurationProvider: ConfigurationProviderFunc(provideErr),
	}
	errHandler := targetErrMiddleware.Handle(GetTestHandler())

	Convey("handle request with error config provider", t, func() {
		req := newReq()
		resp := httptest.NewRecorder()
		errHandler.ServeHTTP(resp, req)
		So(resp.Code, ShouldEqual, http.StatusInternalServerError)
		So(resp.Body.String(), ShouldEqual, "Unable to retrieve configuration\n")
	})
}
