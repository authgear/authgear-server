package sso

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	coreconfig "github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	. "github.com/skygeario/skygear-server/pkg/core/skytest"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

func TestAuthResultHandler(t *testing.T) {
	stateJWTSecret := "secret"
	providerName := "mock"
	providerUserID := "mock_user_id"

	Convey("AuthResultHandler", t, func() {
		sh := &AuthResultHandler{}
		sh.TxContext = db.NewMockTxContext()
		oauthConfig := &coreconfig.OAuthConfiguration{
			StateJWTSecret: stateJWTSecret,
		}
		providerConfig := coreconfig.OAuthProviderConfiguration{
			ID:           providerName,
			Type:         "google",
			ClientID:     "mock_client_id",
			ClientSecret: "mock_client_secret",
		}
		mockProvider := sso.MockSSOProvider{
			URLPrefix:       &url.URL{Scheme: "https", Host: "api.example.com"},
			RedirectURLFunc: RedirectURIForAPI,
			BaseURL:         "http://mock/auth",
			OAuthConfig:     oauthConfig,
			ProviderConfig:  providerConfig,
			UserInfo: sso.ProviderUserInfo{
				ID:    providerUserID,
				Email: "mock@example.com",
			},
		}
		sh.SSOProvider = &mockProvider
		validator := validation.NewValidator("http://v2.skygear.io")
		validator.AddSchemaFragments(
			AuthResultRequestSchema,
		)
		sh.Validator = validator

		Convey("invalid code verifier", func() {
			codeVerifier := "code_verifier"
			codeChallenge := "nonsense"
			code := &sso.SkygearAuthorizationCode{
				Action:              "login",
				CodeChallenge:       codeChallenge,
				UserID:              "john.doe.id",
				PrincipalID:         "john.doe.id",
				SessionCreateReason: "login",
			}
			encodedCode, err := mockProvider.EncodeSkygearAuthorizationCode(*code)
			So(err, ShouldBeNil)

			reqBody := map[string]interface{}{
				"authorization_code": encodedCode,
				"code_verifier":      codeVerifier,
			}
			reqBodyBytes, err := json.Marshal(reqBody)
			So(err, ShouldBeNil)

			req, _ := http.NewRequest("POST", "", bytes.NewReader(reqBodyBytes))
			req.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()
			sh.ServeHTTP(recorder, req)

			So(recorder.Result().StatusCode, ShouldEqual, 401)
			So(recorder.Body.Bytes(), ShouldEqualJSON, `
			{
				"error": {
					"code": 401,
					"info": {
						"cause": {
							"kind": "InvalidCodeVerifier"
						}
					},
					"message": "invalid code verifier",
					"name": "Unauthorized",
					"reason": "SSOFailed"
				}
			}
			`)
		})
	})
}
