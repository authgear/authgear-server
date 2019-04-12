package sso

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/provider/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/provider/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authtoken"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"

	. "github.com/skygeario/skygear-server/pkg/core/skytest"
	. "github.com/smartystreets/goconvey/convey"
)

func TestLoginPayload(t *testing.T) {
	Convey("Test LoginRequestPayload", t, func() {
		// callback URL and ux_mode is required
		Convey("validate valid payload", func() {
			payload := LoginRequestPayload{
				AccessToken: "token",
			}
			So(payload.Validate(), ShouldBeNil)
		})

		Convey("validate payload without access token", func() {
			payload := LoginRequestPayload{}
			err := payload.Validate()
			errResponse := err.(skyerr.Error)
			So(errResponse.Code(), ShouldEqual, skyerr.InvalidArgument)
		})
	})
}

func TestLoginHandler(t *testing.T) {
	realTime := timeNow
	timeNow = func() time.Time { return zeroTime }
	defer func() {
		timeNow = realTime
	}()

	Convey("Test LoginHandler", t, func() {
		stateJWTSecret := "secret"
		providerName := "mock"
		providerUserID := "mock_user_id"

		sh := &LoginHandler{}
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
			BaseURL: "http://mock/auth",
			Setting: setting,
			Config:  config,
			UserID:  providerUserID,
		}
		sh.Provider = &mockProvider
		mockOAuthProvider := oauth.NewMockProvider(
			map[string]string{},
			map[string]oauth.Principal{},
		)
		sh.OAuthAuthProvider = mockOAuthProvider
		authInfoStore := authinfo.NewMockStoreWithAuthInfoMap(
			map[string]authinfo.AuthInfo{},
		)
		sh.AuthInfoStore = authInfoStore
		mockTokenStore := authtoken.NewMockStore()
		sh.TokenStore = mockTokenStore
		loginIDsKeyWhitelist := []string{}
		passwordAuthProvider := password.NewMockProviderWithPrincipalMap(
			loginIDsKeyWhitelist,
			map[string]password.Principal{},
		)
		sh.PasswordAuthProvider = passwordAuthProvider
		sh.UserProfileStore = userprofile.NewMockUserProfileStore()
		h := handler.APIHandlerToHandler(sh, sh.TxContext)

		Convey("should get auth response", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"access_token": "token"
			}`))
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, 200)
			p, _ := sh.OAuthAuthProvider.GetPrincipalByProviderUserID(providerName, providerUserID)
			token := mockTokenStore.GetTokensByAuthInfoID(p.UserID)[0]
			So(resp.Body.Bytes(), ShouldEqualJSON, fmt.Sprintf(`{
				"result": {
					"user_id": "%s",
					"access_token": "%s",
					"verified": false,
					"verify_info": null,
					"created_at": "0001-01-01T00:00:00Z",
					"created_by": "%s",
					"updated_at": "0001-01-01T00:00:00Z",
					"updated_by": "%s",
					"metadata": {}
				}
			}`,
				p.UserID,
				token.AccessToken,
				p.UserID,
				p.UserID))
		})
	})
}
