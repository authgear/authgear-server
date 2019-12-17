package sso

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	coreconfig "github.com/skygeario/skygear-server/pkg/core/config"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAuthRedirectHandler(t *testing.T) {
	Convey("AuthRedirectHandler", t, func() {
		action := "action"
		stateJWTSecret := "secret"
		providerName := "mock"
		providerUserID := "mock_user_id"
		oauthConfig := &coreconfig.OAuthConfiguration{
			StateJWTSecret: stateJWTSecret,
			AllowedCallbackURLs: []string{
				"http://localhost:3000",
			},
		}
		providerConfig := coreconfig.OAuthProviderConfiguration{
			ID:           providerName,
			Type:         "google",
			ClientID:     "mock_client_id",
			ClientSecret: "mock_client_secret",
		}
		mockProvider := &sso.MockSSOProvider{
			URLPrefix:      &url.URL{Scheme: "https", Host: "api.example.com"},
			BaseURL:        "http://mock/auth",
			OAuthConfig:    oauthConfig,
			ProviderConfig: providerConfig,
			UserInfo: sso.ProviderUserInfo{
				ID:    providerUserID,
				Email: "mock@example.com",
			},
		}
		h := &AuthRedirectHandler{}
		h.SSOProvider = mockProvider
		h.OAuthProvider = mockProvider

		Convey("write JSON when ux_mode is manual", func() {
			uxMode := sso.UXModeManual
			// oauth state
			state := sso.State{
				Action: action,
				OAuthAuthorizationCodeFlowState: sso.OAuthAuthorizationCodeFlowState{
					CallbackURL: "http://localhost:3000",
					UXMode:      uxMode,
				},
			}
			encodedState, _ := mockProvider.EncodeState(state)
			v := url.Values{}
			v.Set("state", encodedState)
			u := url.URL{
				RawQuery: v.Encode(),
			}
			req, _ := http.NewRequest("GET", u.RequestURI(), nil)
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)

			So(resp.Result().StatusCode, ShouldEqual, 200)
			So(resp.Result().Header.Get("content-type"), ShouldEqual, "application/json")
		})

		Convey("redirect when ux_mode is not manual", func() {
			uxMode := sso.UXModeWebRedirect
			// oauth state
			state := sso.State{
				Action: action,
				OAuthAuthorizationCodeFlowState: sso.OAuthAuthorizationCodeFlowState{
					CallbackURL: "http://localhost:3000",
					UXMode:      uxMode,
				},
			}
			encodedState, _ := mockProvider.EncodeState(state)
			v := url.Values{}
			v.Set("state", encodedState)
			u := url.URL{
				RawQuery: v.Encode(),
			}
			req, _ := http.NewRequest("GET", u.RequestURI(), nil)
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)

			So(resp.Result().StatusCode, ShouldEqual, 302)
			So(resp.Result().Header.Get("location"), ShouldNotBeEmpty)
		})
	})
}
