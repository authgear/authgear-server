package sso

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"

	"github.com/skygeario/skygear-server/pkg/core/skydb"
	. "github.com/skygeario/skygear-server/pkg/core/skytest"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUnlinkHandler(t *testing.T) {
	Convey("Test UnlinkHandler", t, func() {
		providerID := "google"
		providerUserID := "mock_user_id"

		sh := &UnlinkHandler{}
		sh.ProviderID = providerID
		sh.TxContext = db.NewMockTxContext()
		sh.AuthContext = auth.NewMockContextGetterWithDefaultUser()
		sh.ProviderFactory = sso.NewProviderFactory(config.TenantConfiguration{
			UserConfig: config.UserConfiguration{
				SSO: config.SSOConfiguration{
					OAuth: config.OAuthConfiguration{
						Providers: []config.OAuthProviderConfiguration{
							config.OAuthProviderConfiguration{
								Type: "google",
								ID:   "google",
							},
						},
					},
				},
			},
		})
		mockOAuthProvider := oauth.NewMockProvider([]*oauth.Principal{
			&oauth.Principal{
				ProviderType:   "google",
				ProviderKeys:   map[string]interface{}{},
				ProviderUserID: providerUserID,
				UserID:         "faseng.cat.id",
			},
		})
		sh.OAuthAuthProvider = mockOAuthProvider
		h := handler.APIHandlerToHandler(sh, sh.TxContext)

		Convey("should unlink user id with oauth principal", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"access_token": "token"
			}`))
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, 200)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": {}
			}`)

			p, e := sh.OAuthAuthProvider.GetPrincipalByProvider(oauth.GetByProviderOptions{
				ProviderType:   "google",
				ProviderUserID: providerUserID,
			})
			So(e, ShouldEqual, skydb.ErrUserNotFound)
			So(p, ShouldBeNil)
		})
	})
}
