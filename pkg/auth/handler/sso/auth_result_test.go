package sso

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/session"
	coreconfig "github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	coreTime "github.com/skygeario/skygear-server/pkg/core/time"
	"github.com/skygeario/skygear-server/pkg/core/validation"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/authnsession"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/mfa"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"

	. "github.com/skygeario/skygear-server/pkg/core/skytest"
	. "github.com/smartystreets/goconvey/convey"
)

func codeVerifierToCodeChallenge(codeVerifier string) string {
	sha256Arr := sha256.Sum256([]byte(codeVerifier))
	sha256Slice := sha256Arr[:]
	codeChallenge := base64.RawURLEncoding.EncodeToString(sha256Slice)
	return codeChallenge
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
		sh.SSOProvider = &mockProvider
		mockOAuthProvider := oauth.NewMockProvider([]*oauth.Principal{
			&oauth.Principal{
				ID:             "john.doe.id",
				UserID:         "john.doe.id",
				ProviderType:   "google",
				ProviderKeys:   map[string]interface{}{},
				ProviderUserID: providerUserID,
			},
		})
		sh.OAuthAuthProvider = mockOAuthProvider
		authInfoStore := authinfo.NewMockStoreWithAuthInfoMap(
			map[string]authinfo.AuthInfo{
				"john.doe.id": authinfo.AuthInfo{
					ID: "john.doe.id",
				},
			},
		)
		sh.AuthInfoStore = authInfoStore
		sessionProvider := session.NewMockProvider()
		sessionWriter := session.NewMockWriter()
		userProfileStore := userprofile.NewMockUserProfileStore()
		sh.UserProfileStore = userProfileStore
		one := 1
		loginIDsKeys := []coreconfig.LoginIDKeyConfiguration{
			coreconfig.LoginIDKeyConfiguration{Key: "email", Maximum: &one},
		}
		allowedRealms := []string{password.DefaultRealm}
		passwordAuthProvider := password.NewMockProviderWithPrincipalMap(
			loginIDsKeys,
			allowedRealms,
			map[string]password.Principal{},
		)
		identityProvider := principal.NewMockIdentityProvider(sh.OAuthAuthProvider, passwordAuthProvider)
		sh.IdentityProvider = identityProvider
		hookProvider := hook.NewMockProvider()
		sh.HookProvider = hookProvider
		timeProvider := &coreTime.MockProvider{TimeNowUTC: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)}
		sh.TimeProvider = timeProvider
		mfaStore := mfa.NewMockStore(timeProvider)
		mfaConfiguration := &coreconfig.MFAConfiguration{
			Enabled:     false,
			Enforcement: coreconfig.MFAEnforcementOptional,
		}
		mfaSender := mfa.NewMockSender()
		mfaProvider := mfa.NewProvider(mfaStore, mfaConfiguration, timeProvider, mfaSender)
		sh.AuthnSessionProvider = authnsession.NewMockProvider(
			mfaConfiguration,
			timeProvider,
			mfaProvider,
			authInfoStore,
			sessionProvider,
			sessionWriter,
			identityProvider,
			hookProvider,
			userProfileStore,
		)
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

		Convey("action = login", func() {
			codeVerifier := "code_verifier"
			codeChallenge := codeVerifierToCodeChallenge(codeVerifier)
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

			So(recorder.Result().StatusCode, ShouldEqual, 200)
			So(recorder.Body.Bytes(), ShouldEqualJSON, `
			{
				"result": {
					"access_token": "access-token-john.doe.id-john.doe.id-0",
					"identity": {
						"claims": null,
						"id": "john.doe.id",
						"provider_keys": {},
						"provider_type": "google",
						"provider_user_id": "mock_user_id",
						"raw_profile": null,
						"type": "oauth"
					},
					"session_id": "john.doe.id-john.doe.id-0",
					"user": {
						"created_at": "0001-01-01T00:00:00Z",
						"id": "john.doe.id",
						"is_disabled": false,
						"is_manually_verified": false,
						"is_verified": false,
						"metadata": {},
						"verify_info": {}
					}
				}
			}
			`)
		})

		Convey("action = link", func() {
			codeVerifier := "code_verifier"
			codeChallenge := codeVerifierToCodeChallenge(codeVerifier)
			code := &sso.SkygearAuthorizationCode{
				Action:        "link",
				CodeChallenge: codeChallenge,
				UserID:        "john.doe.id",
				PrincipalID:   "john.doe.id",
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

			So(recorder.Result().StatusCode, ShouldEqual, 200)
			So(recorder.Body.Bytes(), ShouldEqualJSON, `
			{
				"result": {
					"user": {
						"created_at": "0001-01-01T00:00:00Z",
						"id": "john.doe.id",
						"is_disabled": false,
						"is_manually_verified": false,
						"is_verified": false,
						"metadata": {},
						"verify_info": {}
					}
				}
			}
			`)
		})
	})
}
