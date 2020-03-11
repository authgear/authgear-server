package sso

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/skygeario/skygear-server/pkg/core/apiclientconfig"
	"github.com/skygeario/skygear-server/pkg/core/config"
	. "github.com/skygeario/skygear-server/pkg/core/skytest"
	. "github.com/smartystreets/goconvey/convey"
)

func TestConfigHandler(t *testing.T) {
	Convey("Test ConfigHandler", t, func() {
		Convey("should return tenant SSOSeting AllowedCallbackURLs", func() {
			r, _ := http.NewRequest("POST", "", nil)
			rw := httptest.NewRecorder()

			var testingHandler ConfigHandler
			testingHandler.ClientProvider = &apiclientconfig.MockProvider{
				ClientID: "client_id",
				APIClientConfig: config.OAuthClientConfiguration{
					"redirect_uris": []interface{}{
						"http://localhost",
						"http://127.0.0.1",
					},
				},
			}
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
