package sso

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authn"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/core/config"
	coreconfig "github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	. "github.com/skygeario/skygear-server/pkg/core/skytest"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

type MockLoginAuthnProvider struct{}

func (p *MockLoginAuthnProvider) OAuthAuthenticate(
	authInfo sso.AuthInfo,
	codeChallenge string,
	loginState sso.LoginState,
) (*sso.SkygearAuthorizationCode, error) {
	panic("not mocked")
}

func (p *MockLoginAuthnProvider) OAuthExchangeCode(
	client config.OAuthClientConfiguration,
	session auth.Session,
	code *sso.SkygearAuthorizationCode,
) (authn.Result, error) {
	panic("not mocked")
}

func (p *MockLoginAuthnProvider) WriteResult(http.ResponseWriter, authn.Result) {
	panic("not mocked")
}

func TestLoginHandler(t *testing.T) {
	Convey("Test LoginHandler", t, func() {
		stateJWTSecret := "secret"
		providerName := "mock"
		providerUserID := "mock_user_id"

		sh := &LoginHandler{}
		sh.TxContext = db.NewMockTxContext()
		validator := validation.NewValidator("http://v2.skygear.io")
		validator.AddSchemaFragments(
			LoginRequestSchema,
		)
		sh.Validator = validator
		oauthConfig := &coreconfig.OAuthConfiguration{
			StateJWTSecret:                 stateJWTSecret,
			ExternalAccessTokenFlowEnabled: true,
		}
		providerConfig := coreconfig.OAuthProviderConfiguration{
			ID:           providerName,
			Type:         "google",
			ClientID:     "mock_client_id",
			ClientSecret: "mock_client_secret",
		}
		mockProvider := sso.MockSSOProvider{
			URLPrefix:      &url.URL{Scheme: "https", Host: "api.example.com"},
			BaseURL:        "http://mock/auth",
			OAuthConfig:    oauthConfig,
			ProviderConfig: providerConfig,
			UserInfo: sso.ProviderUserInfo{
				ID:    providerUserID,
				Email: "mock@example.com",
			},
		}
		sh.OAuthProvider = &mockProvider
		sh.SSOProvider = &mockProvider
		sh.AuthnProvider = &MockLoginAuthnProvider{}

		Convey("should reject payload without access token", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{}`))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			sh.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, 400)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"error": {
					"name": "Invalid",
					"reason": "ValidationFailed",
					"message": "invalid request body",
					"code": 400,
					"info": {
						"causes": [
							{
								"kind": "Required",
								"message": "access_token is required",
								"pointer": "/access_token"
							}
						]
					}
				}
			}`)
		})

		Convey("should return error if disabled", func() {
			mockProvider.OAuthConfig.ExternalAccessTokenFlowEnabled = false
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"access_token": "token"
			}`))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			sh.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, 404)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"error": {
					"name": "NotFound",
					"reason": "NotFound",
					"message": "external access token flow is disabled",
					"code": 404
				}
			}`)
		})
	})
}
