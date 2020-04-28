package sso

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authn"
	interactionflows "github.com/skygeario/skygear-server/pkg/auth/dependency/interaction/flows"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/core/config"
	coreconfig "github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/crypto"
	"github.com/skygeario/skygear-server/pkg/core/db"
	. "github.com/skygeario/skygear-server/pkg/core/skytest"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

type MockOAuthResultInteractionFlow struct {
}

func (p *MockOAuthResultInteractionFlow) ExchangeCode(codeHash string, verifier string) (*interactionflows.AuthResult, error) {
	return nil, errors.New("not mocked")
}

type MockAuthResultAuthnProvider struct {
	code *sso.SkygearAuthorizationCode
}

func (p *MockAuthResultAuthnProvider) OAuthConsumeCode(hashCode string) (*sso.SkygearAuthorizationCode, error) {
	if hashCode == p.code.CodeHash {
		code := p.code
		p.code = nil
		return code, nil
	}
	return nil, sso.ErrCodeNotFound
}

func (p *MockAuthResultAuthnProvider) OAuthExchangeCode(
	client config.OAuthClientConfiguration,
	session auth.AuthSession,
	code *sso.SkygearAuthorizationCode,
) (authn.Result, error) {
	panic("not mocked")
}

func (p *MockAuthResultAuthnProvider) WriteAPIResult(rw http.ResponseWriter, result authn.Result) {
	panic("not mocked")
}

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
		codeVerifier := "code_verifier"
		codeChallenge := "nonsense"
		codeStr := "code"
		mockAuthnProvider := &MockAuthResultAuthnProvider{
			code: &sso.SkygearAuthorizationCode{
				CodeHash:            crypto.SHA256String(codeStr),
				Action:              "login",
				CodeChallenge:       codeChallenge,
				UserID:              "john.doe.id",
				PrincipalID:         "john.doe.id",
				SessionCreateReason: "login",
			},
		}

		sh.SSOProvider = &mockProvider
		sh.AuthnProvider = mockAuthnProvider
		validator := validation.NewValidator("http://v2.skygear.io")
		validator.AddSchemaFragments(
			AuthResultRequestSchema,
		)
		sh.Validator = validator
		sh.Interactions = &MockOAuthResultInteractionFlow{}

		Convey("invalid code verifier", func() {
			// this test is testing the old flow that
			// code verifier is validation in handler
			reqBody := map[string]interface{}{
				"authorization_code": codeStr,
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
