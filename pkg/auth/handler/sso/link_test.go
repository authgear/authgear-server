package sso

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/provider/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"

	. "github.com/skygeario/skygear-server/pkg/server/skytest"
	. "github.com/smartystreets/goconvey/convey"
)

func TestLinkPayload(t *testing.T) {
	Convey("Test LinkRequestPayload", t, func() {
		// callback URL and ux_mode is required
		Convey("validate valid payload", func() {
			payload := LinkRequestPayload{
				AccessTokenResp: sso.AccessTokenResp{
					AccessToken: "token",
				},
			}
			So(payload.Validate(), ShouldBeNil)
		})

		Convey("validate payload without access token", func() {
			payload := LinkRequestPayload{}
			err := payload.Validate()
			errResponse := err.(skyerr.Error)
			So(errResponse.Code(), ShouldEqual, skyerr.InvalidArgument)
		})
	})
}

func TestLinkHandler(t *testing.T) {
	realTime := timeNow
	timeNow = func() time.Time { return zeroTime }
	defer func() {
		timeNow = realTime
	}()

	Convey("Test LinkHandler", t, func() {
		stateJWTSecret := "secret"
		providerName := "mock"
		providerUserID := "mock_user_id"
		providerAuthData := map[string]interface{}{
			"email": "john.doe@example.com",
		}

		sh := &LinkHandler{}
		sh.TxContext = db.NewMockTxContext()
		sh.AuthContext = auth.NewMockContextGetterWithDefaultUser()
		setting := sso.Setting{
			URLPrefix:      "http://localhost:3000",
			StateJWTSecret: stateJWTSecret,
			AllowedCallbackURLs: []string{
				"http://localhost",
			},
		}
		config := sso.Config{
			Name:         providerName,
			ClientID:     "mock_client_id",
			ClientSecret: "mock_client_secret",
		}
		mockProvider := sso.MockSSOProverImpl{
			BaseURL:      "http://mock/auth",
			Setting:      setting,
			Config:       config,
			MockUserID:   providerUserID,
			MockAuthData: providerAuthData,
		}
		sh.Provider = &mockProvider
		ssoProviderFactory := sso.ProviderFactory{}
		sh.AuthInfoProcessor = ssoProviderFactory.NewAuthInfoProcessor(providerName)
		mockOAuthProvider := oauth.NewMockProvider(
			map[string]string{},
			map[string]oauth.Principal{},
		)
		sh.OAuthAuthProvider = mockOAuthProvider
		authInfoStore := authinfo.NewMockStoreWithAuthInfoMap(
			map[string]authinfo.AuthInfo{
				"faseng.cat.id": authinfo.AuthInfo{
					ID: "faseng.cat.id",
				},
			},
		)
		sh.AuthInfoStore = authInfoStore
		h := handler.APIHandlerToHandler(sh, sh.TxContext)

		Convey("should link user id with oauth principal", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"access_token": "token"
			}`))
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, 200)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": "OK"
			}`)

			p, _ := sh.OAuthAuthProvider.GetPrincipalByProviderUserID(providerName, providerUserID)
			So(p.UserID, ShouldEqual, "faseng.cat.id")
		})
	})
}
