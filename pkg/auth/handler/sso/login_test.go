package sso

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authtoken"
	coreconfig "github.com/skygeario/skygear-server/pkg/core/config"
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
	timeNow = func() time.Time { return time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC) }
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
		mockProvider := sso.MockSSOProvider{
			BaseURL:  "http://mock/auth",
			Setting:  setting,
			Config:   config,
			UserInfo: sso.ProviderUserInfo{ID: providerUserID},
		}
		sh.Provider = &mockProvider
		mockOAuthProvider := oauth.NewMockProvider(
			map[string]string{},
			map[string]oauth.Principal{},
		)
		sh.OAuthAuthProvider = mockOAuthProvider
		sh.IdentityProvider = principal.NewMockIdentityProvider(sh.OAuthAuthProvider)
		authInfoStore := authinfo.NewMockStoreWithAuthInfoMap(
			map[string]authinfo.AuthInfo{},
		)
		sh.AuthInfoStore = authInfoStore
		mockTokenStore := authtoken.NewMockStore()
		sh.TokenStore = mockTokenStore
		loginIDsKeys := map[string]coreconfig.LoginIDKeyConfiguration{}
		allowedRealms := []string{password.DefaultRealm}
		passwordAuthProvider := password.NewMockProviderWithPrincipalMap(
			loginIDsKeys,
			allowedRealms,
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
					"user": {
						"id": "%s",
						"is_verified": false,
						"is_disabled": false,
						"last_login_at": "2006-01-02T15:04:05Z",
						"created_at": "0001-01-01T00:00:00Z",
						"verify_info": {},
						"metadata": {},
						"identity": {
							"id": "%s",
							"type": "oauth",
							"provider_id": "mock",
							"provider_user_id": "mock_user_id",
							"raw_profile": {},
							"claims": {}
						}
					},
					"access_token": "%s"
				}
			}`,
				p.UserID,
				p.ID,
				token.AccessToken))
		})
	})
}
