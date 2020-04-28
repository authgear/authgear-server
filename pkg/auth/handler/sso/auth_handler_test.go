package sso

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/config"
	coreconfig "github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/crypto"
	"github.com/skygeario/skygear-server/pkg/core/db"

	. "github.com/skygeario/skygear-server/pkg/core/skytest"
	. "github.com/smartystreets/goconvey/convey"
)

func decodeResultInURL(ssoProvider sso.Provider, urlString string) ([]byte, error) {
	u, err := url.Parse(urlString)
	if err != nil {
		return nil, err
	}
	result := u.Query().Get("x-skygear-result")
	bytes, err := base64.StdEncoding.DecodeString(result)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func TestAuthPayload(t *testing.T) {
	Convey("Test AuthRequestPayload", t, func() {
		Convey("validate valid payload", func() {
			payload := AuthRequestPayload{
				Code:  "code",
				State: "state",
			}
			So(payload.Validate(), ShouldBeNil)
		})

		Convey("validate payload without code", func() {
			payload := AuthRequestPayload{
				State: "state",
			}
			So(payload.Validate(), ShouldBeError)
		})

		Convey("validate payload without state", func() {
			payload := AuthRequestPayload{
				Code: "code",
			}
			So(payload.Validate(), ShouldBeError)
		})
	})
}

type MockOAuthHandlerInteractionFlow struct {
	CodeStr string
}

func (m *MockOAuthHandlerInteractionFlow) LoginWithOAuthProvider(clientID string, oauthAuthInfo sso.AuthInfo, codeChallenge string) (string, error) {
	return m.CodeStr, nil
}

func TestAuthHandler(t *testing.T) {
	Convey("Test AuthHandler with login action", t, func() {
		action := "login"
		stateJWTSecret := "secret"
		providerName := "mock"
		providerUserID := "mock_user_id"
		sh := &AuthHandler{}
		sh.TxContext = db.NewMockTxContext()
		accessKey := auth.AccessKey{
			Client: config.OAuthClientConfiguration{
				"client_name":            "client-id",
				"client_id":              "client-id",
				"redirect_uris":          []interface{}{"http://localhost:3000"},
				"access_token_lifetime":  1800.0,
				"refresh_token_lifetime": 86400.0,
			},
		}
		sh.TenantConfiguration = &config.TenantConfiguration{
			AppConfig: &config.AppConfiguration{
				Clients: []config.OAuthClientConfiguration{accessKey.Client},
			},
		}
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
		sh.OAuthProvider = &mockProvider
		sh.SSOProvider = &mockProvider
		sh.AuthHandlerHTMLProvider = sso.NewAuthHandlerHTMLProvider(
			&url.URL{Scheme: "https", Host: "api.example.com"},
		)
		interactions := &MockOAuthHandlerInteractionFlow{}
		sh.Interactions = interactions

		nonce := "nonce"
		hashedNonce := crypto.SHA256String(nonce)

		Convey("should write code in the response body if ux_mode is manual", func() {
			interactions.CodeStr = "code"
			// oauth state
			state := sso.State{
				APIClientID: "client-id",
				Action:      action,
				Extra: AuthAPISSOState{
					"callback_url": "http://localhost:3000",
				},
				UXMode:      sso.UXModeManual,
				HashedNonce: hashedNonce,
			}
			encodedState, _ := mockProvider.EncodeState(state)
			v := url.Values{}
			v.Set("code", "code")
			v.Add("state", encodedState)
			u := url.URL{
				RawQuery: v.Encode(),
			}

			req, _ := http.NewRequest("GET", u.RequestURI(), nil)
			req = req.WithContext(auth.WithAccessKey(req.Context(), accessKey))
			resp := httptest.NewRecorder()
			sh.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, fmt.Sprintf(`
			{
				"result": "%s"
			}`, interactions.CodeStr))
		})

		Convey("should return callback url when ux_mode is web_redirect", func() {
			interactions.CodeStr = "code"
			// oauth state
			state := sso.State{
				APIClientID: "client-id",
				Action:      action,
				Extra: AuthAPISSOState{
					"callback_url": "http://localhost:3000",
				},
				UXMode:      sso.UXModeWebRedirect,
				HashedNonce: hashedNonce,
			}
			encodedState, _ := mockProvider.EncodeState(state)

			v := url.Values{}
			v.Set("code", "code")
			v.Add("state", encodedState)
			u := url.URL{
				RawQuery: v.Encode(),
			}

			req, _ := http.NewRequest("GET", u.RequestURI(), nil)
			req = req.WithContext(auth.WithAccessKey(req.Context(), accessKey))
			resp := httptest.NewRecorder()

			sh.ServeHTTP(resp, req)
			// for web_redirect, it should redirect to original callback url
			So(resp.Code, ShouldEqual, 302)
			location := resp.Result().Header.Get("Location")
			actual, err := decodeResultInURL(sh.SSOProvider, location)
			So(err, ShouldBeNil)

			So(actual, ShouldEqualJSON, fmt.Sprintf(`
			{
				"callback_url": "http://localhost:3000",
				"result": {
					"result": "%s"
				}
			}`, interactions.CodeStr))
		})

		Convey("should return html page when ux_mode is web_popup", func() {
			interactions.CodeStr = "code"
			// oauth state
			state := sso.State{
				APIClientID: "client-id",
				Action:      action,
				Extra: AuthAPISSOState{
					"callback_url": "http://localhost:3000",
				},
				UXMode:      sso.UXModeWebPopup,
				HashedNonce: hashedNonce,
			}
			encodedState, _ := mockProvider.EncodeState(state)

			v := url.Values{}
			v.Set("code", "code")
			v.Add("state", encodedState)
			u := url.URL{
				RawQuery: v.Encode(),
			}

			req, _ := http.NewRequest("GET", u.RequestURI(), nil)
			req = req.WithContext(auth.WithAccessKey(req.Context(), accessKey))
			resp := httptest.NewRecorder()

			sh.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, 200)
			So(resp.Result().Header.Get("Content-Type"), ShouldEqual, "text/html; charset=utf-8")
		})

		Convey("should return callback url with result query parameter when ux_mode is mobile_app", func() {
			interactions.CodeStr = "code"
			// oauth state
			state := sso.State{
				APIClientID: "client-id",
				Action:      action,
				Extra: AuthAPISSOState{
					"callback_url": "http://localhost:3000",
				},
				UXMode:      sso.UXModeMobileApp,
				HashedNonce: hashedNonce,
			}
			encodedState, _ := mockProvider.EncodeState(state)

			v := url.Values{}
			v.Set("code", "code")
			v.Add("state", encodedState)
			u := url.URL{
				RawQuery: v.Encode(),
			}

			req, _ := http.NewRequest("GET", u.RequestURI(), nil)
			req = req.WithContext(auth.WithAccessKey(req.Context(), accessKey))
			resp := httptest.NewRecorder()

			sh.ServeHTTP(resp, req)
			// for mobile app, it should redirect to original callback url
			So(resp.Code, ShouldEqual, 302)
			// check location result query parameter
			actual, err := decodeResultInURL(sh.SSOProvider, resp.Header().Get("Location"))
			So(err, ShouldBeNil)
			So(actual, ShouldEqualJSON, fmt.Sprintf(`{
				"callback_url": "http://localhost:3000",
				"result": {
					"result": "%s"
				}
			}`, interactions.CodeStr))
		})
	})
}
