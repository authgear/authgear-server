package sso

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/authnsession"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	coreconfig "github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	. "github.com/skygeario/skygear-server/pkg/core/skytest"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

type MockAuthnSessionProvider struct{}

var _ authnsession.Provider = &MockAuthnSessionProvider{}

func (p *MockAuthnSessionProvider) NewFromToken(token string) (*coreAuth.AuthnSession, error) {
	panic("not mocked")
}

func (p *MockAuthnSessionProvider) NewFromScratch(userID string, prin principal.Principal, reason coreAuth.SessionCreateReason) (*coreAuth.AuthnSession, error) {
	panic("not mocked")
}

func (p *MockAuthnSessionProvider) GenerateResponseAndUpdateLastLoginAt(session coreAuth.AuthnSession) (interface{}, error) {
	panic("not mocked")
}

func (p *MockAuthnSessionProvider) GenerateResponseWithSession(sess *coreAuth.Session, mfaBearerToken string) (interface{}, error) {
	panic("not mocked")
}

func (p *MockAuthnSessionProvider) WriteResponse(w http.ResponseWriter, resp interface{}, err error) {
	if err != nil {
		handler.WriteResponse(w, handler.APIResponse{Error: err})
	} else {
		handler.WriteResponse(w, handler.APIResponse{Result: resp})
	}
}

func (p *MockAuthnSessionProvider) Resolve(authContext coreAuth.ContextGetter, authnSessionToken string, options authnsession.ResolveOptions) (userID string, sess *coreAuth.Session, authnSession *coreAuth.AuthnSession, err error) {
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
			RedirectURIs: []string{
				"http://localhost",
			},
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
		hookProvider := hook.NewMockProvider()
		sh.HookProvider = hookProvider
		sh.AuthnSessionProvider = &MockAuthnSessionProvider{}

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
