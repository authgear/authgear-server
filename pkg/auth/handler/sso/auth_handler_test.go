package sso

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"testing"
	"time"

	coreconfig "github.com/skygeario/skygear-server/pkg/core/config"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authtoken"
	"github.com/skygeario/skygear-server/pkg/core/auth/metadata"
	"github.com/skygeario/skygear-server/pkg/core/crypto"
	"github.com/skygeario/skygear-server/pkg/core/db"
	coreHttp "github.com/skygeario/skygear-server/pkg/core/http"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"

	. "github.com/skygeario/skygear-server/pkg/core/skytest"
	. "github.com/smartystreets/goconvey/convey"
)

func decodeCookie(resp *httptest.ResponseRecorder) ([]byte, error) {
	cookies := resp.Result().Cookies()
	for _, c := range cookies {
		if c.Name == "sso_data" {
			decoded, err := base64.StdEncoding.DecodeString(c.Value)
			if err != nil {
				return nil, err
			}
			return decoded, nil
		}
	}
	return nil, errors.New("not_found")
}

func TestAuthPayload(t *testing.T) {
	Convey("Test AuthRequestPayload", t, func() {
		Convey("validate valid payload", func() {
			payload := AuthRequestPayload{
				Code:  "code",
				State: "state",
				Nonce: "nonce",
			}
			So(payload.Validate(), ShouldBeNil)
		})

		Convey("validate payload without code", func() {
			payload := AuthRequestPayload{
				State: "state",
			}
			err := payload.Validate()
			errResponse := err.(skyerr.Error)
			So(errResponse.Code(), ShouldEqual, skyerr.InvalidArgument)
		})

		Convey("validate payload without state", func() {
			payload := AuthRequestPayload{
				Code: "code",
			}
			err := payload.Validate()
			errResponse := err.(skyerr.Error)
			So(errResponse.Code(), ShouldEqual, skyerr.InvalidArgument)
		})
	})
}

func TestAuthHandler(t *testing.T) {
	realTime := timeNow
	timeNow = func() time.Time { return time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC) }
	defer func() {
		timeNow = realTime
	}()

	Convey("Test AuthHandler with login action", t, func() {
		action := "login"
		stateJWTSecret := "secret"
		providerName := "mock"
		providerUserID := "mock_user_id"
		sh := &AuthHandler{}
		sh.TxContext = db.NewMockTxContext()
		sh.AuthContext = auth.NewMockContextGetterWithDefaultUser()
		oauthConfig := coreconfig.OAuthConfiguration{
			URLPrefix:      "http://localhost:3000",
			StateJWTSecret: stateJWTSecret,
			AllowedCallbackURLs: []string{
				"http://localhost",
			},
		}
		providerConfig := coreconfig.OAuthProviderConfiguration{
			ID:           providerName,
			Type:         "google",
			ClientID:     "mock_client_id",
			ClientSecret: "mock_client_secret",
		}
		mockProvider := sso.MockSSOProvider{
			BaseURL:        "http://mock/auth",
			OAuthConfig:    oauthConfig,
			ProviderConfig: providerConfig,
			UserInfo: sso.ProviderUserInfo{
				ID:    providerUserID,
				Email: "mock@example.com",
			},
		}
		sh.Provider = &mockProvider
		mockOAuthProvider := oauth.NewMockProvider(nil)
		sh.OAuthAuthProvider = mockOAuthProvider
		authInfoStore := authinfo.NewMockStoreWithAuthInfoMap(
			map[string]authinfo.AuthInfo{},
		)
		sh.AuthInfoStore = authInfoStore
		mockTokenStore := authtoken.NewMockStore()
		sh.TokenStore = mockTokenStore
		sh.UserProfileStore = userprofile.NewMockUserProfileStore()
		sh.AuthHandlerHTMLProvider = sso.NewAuthHandlerHTMLProvider(
			"https://api.example.com",
		)
		sh.OAuthConfiguration = oauthConfig
		zero := 0
		one := 1
		loginIDsKeys := map[string]coreconfig.LoginIDKeyConfiguration{
			"email": coreconfig.LoginIDKeyConfiguration{Minimum: &zero, Maximum: &one},
		}
		allowedRealms := []string{password.DefaultRealm}
		passwordAuthProvider := password.NewMockProviderWithPrincipalMap(
			loginIDsKeys,
			allowedRealms,
			map[string]password.Principal{},
		)
		sh.IdentityProvider = principal.NewMockIdentityProvider(sh.OAuthAuthProvider, passwordAuthProvider)
		sh.HookProvider = hook.NewMockProvider()
		nonce := "nonce"
		hashedNonce := crypto.SHA256String(nonce)
		nonceCookie := &http.Cookie{
			Name:  coreHttp.CookieNameOpenIDConnectNonce,
			Value: nonce,
		}

		Convey("should return callback url when ux_mode is web_redirect", func() {
			uxMode := sso.UXModeWebRedirect

			// oauth state
			state := sso.State{
				OAuthAuthorizationCodeFlowState: sso.OAuthAuthorizationCodeFlowState{
					CallbackURL: "http://localhost:3000",
					UXMode:      uxMode,
					Action:      action,
				},
				Nonce: hashedNonce,
			}
			encodedState, _ := sso.EncodeState(stateJWTSecret, state)

			v := url.Values{}
			v.Set("code", "code")
			v.Add("state", encodedState)
			u := url.URL{
				RawQuery: v.Encode(),
			}

			req, _ := http.NewRequest("GET", u.RequestURI(), nil)
			req.AddCookie(nonceCookie)
			resp := httptest.NewRecorder()

			sh.ServeHTTP(resp, req)
			// for web_redirect, it should redirect to original callback url
			So(resp.Code, ShouldEqual, 302)
			So(resp.Header().Get("Location"), ShouldEqual, "http://localhost:3000")

			actual, err := decodeCookie(resp)
			So(err, ShouldBeNil)
			p, err := sh.OAuthAuthProvider.GetPrincipalByProvider(oauth.GetByProviderOptions{
				ProviderType:   "google",
				ProviderUserID: providerUserID,
			})
			So(err, ShouldBeNil)
			token := mockTokenStore.GetTokensByAuthInfoID(p.UserID)[0]
			So(actual, ShouldEqualJSON, fmt.Sprintf(`
			{
				"callback_url": "http://localhost:3000",
				"result": {
					"result": {
						"user": {
							"id": "%s",
							"is_verified": false,
							"is_disabled": false,
							"last_login_at": "2006-01-02T15:04:05Z",
							"created_at": "0001-01-01T00:00:00Z",
							"verify_info": {},
							"metadata": {}
						},
						"identity": {
							"id": "%s",
							"type": "oauth",
							"provider_type": "google",
							"provider_user_id": "mock_user_id",
							"raw_profile": {
								"id": "mock_user_id",
								"email": "mock@example.com"
							},
							"claims": {
								"email": "mock@example.com"
							}
						},
						"access_token": "%s"
					}
				}
			}`,
				p.UserID,
				p.ID,
				token.AccessToken))
		})

		Convey("should return html page when ux_mode is web_popup", func() {
			uxMode := sso.UXModeWebPopup

			// oauth state
			state := sso.State{
				OAuthAuthorizationCodeFlowState: sso.OAuthAuthorizationCodeFlowState{
					CallbackURL: "http://localhost:3000",
					UXMode:      uxMode,
					Action:      action,
				},
				Nonce: hashedNonce,
			}
			encodedState, _ := sso.EncodeState(stateJWTSecret, state)

			v := url.Values{}
			v.Set("code", "code")
			v.Add("state", encodedState)
			u := url.URL{
				RawQuery: v.Encode(),
			}

			req, _ := http.NewRequest("GET", u.RequestURI(), nil)
			req.AddCookie(nonceCookie)
			resp := httptest.NewRecorder()

			sh.ServeHTTP(resp, req)
			// for web_redirect, it should redirect to original callback url
			So(resp.Code, ShouldEqual, 200)
			apiEndpointPattern := `"https://api.example.com/_auth/sso/config"`
			matched, err := regexp.MatchString(apiEndpointPattern, resp.Body.String())
			So(err, ShouldBeNil)
			So(matched, ShouldBeTrue)
		})

		Convey("should return callback url with result query parameter when ux_mode is mobile_app", func() {
			uxMode := sso.UXModeMobileApp

			// oauth state
			state := sso.State{
				OAuthAuthorizationCodeFlowState: sso.OAuthAuthorizationCodeFlowState{
					CallbackURL: "http://localhost:3000",
					UXMode:      uxMode,
					Action:      action,
				},
				Nonce: hashedNonce,
			}
			encodedState, _ := sso.EncodeState(stateJWTSecret, state)

			v := url.Values{}
			v.Set("code", "code")
			v.Add("state", encodedState)
			u := url.URL{
				RawQuery: v.Encode(),
			}

			req, _ := http.NewRequest("GET", u.RequestURI(), nil)
			req.AddCookie(nonceCookie)
			resp := httptest.NewRecorder()

			sh.ServeHTTP(resp, req)
			// for mobile app, it should redirect to original callback url
			So(resp.Code, ShouldEqual, 302)
			// check location result query parameter
			location, _ := url.Parse(resp.Header().Get("Location"))
			q := location.Query()
			result := q.Get("result")
			decoded, _ := base64.StdEncoding.DecodeString(result)
			p, _ := sh.OAuthAuthProvider.GetPrincipalByProvider(oauth.GetByProviderOptions{
				ProviderType:   "google",
				ProviderUserID: providerUserID,
			})
			token := mockTokenStore.GetTokensByAuthInfoID(p.UserID)[0]
			So(decoded, ShouldEqualJSON, fmt.Sprintf(`{
				"result": {
					"user": {
						"id": "%s",
						"is_verified": false,
						"is_disabled": false,
						"last_login_at": "2006-01-02T15:04:05Z",
						"created_at": "0001-01-01T00:00:00Z",
						"verify_info": {},
						"metadata": {}
					},
					"identity": {
						"id": "%s",
						"type": "oauth",
						"provider_type": "google",
						"provider_user_id": "mock_user_id",
						"raw_profile": {
							"id": "mock_user_id",
							"email": "mock@example.com"
						},
						"claims": {
							"email": "mock@example.com"
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

	Convey("Test AuthHandler with link action", t, func() {
		action := "link"
		stateJWTSecret := "secret"
		sh := &AuthHandler{}
		sh.TxContext = db.NewMockTxContext()
		sh.AuthContext = auth.NewMockContextGetterWithDefaultUser()
		oauthConfig := coreconfig.OAuthConfiguration{
			URLPrefix:      "http://localhost:3000",
			StateJWTSecret: stateJWTSecret,
			AllowedCallbackURLs: []string{
				"http://localhost",
			},
		}
		providerConfig := coreconfig.OAuthProviderConfiguration{
			ID:           "mock",
			Type:         "google",
			ClientID:     "mock_client_id",
			ClientSecret: "mock_client_secret",
		}
		mockProvider := sso.MockSSOProvider{
			BaseURL:        "http://mock/auth",
			OAuthConfig:    oauthConfig,
			ProviderConfig: providerConfig,
		}
		sh.Provider = &mockProvider
		mockOAuthProvider := oauth.NewMockProvider([]*oauth.Principal{
			&oauth.Principal{
				ID:           "jane.doe.id",
				UserID:       "jane.doe.id",
				ProviderType: "google",
				ProviderKeys: map[string]interface{}{},
			},
		})
		sh.OAuthAuthProvider = mockOAuthProvider
		authInfoStore := authinfo.NewMockStoreWithAuthInfoMap(
			map[string]authinfo.AuthInfo{
				"john.doe.id": authinfo.AuthInfo{
					ID: "john.doe.id",
				},
				"jane.doe.id": authinfo.AuthInfo{
					ID: "jane.doe.id",
				},
			},
		)
		sh.AuthInfoStore = authInfoStore
		mockTokenStore := authtoken.NewMockStore()
		sh.TokenStore = mockTokenStore
		sh.UserProfileStore = userprofile.NewMockUserProfileStore()
		sh.AuthHandlerHTMLProvider = sso.NewAuthHandlerHTMLProvider(
			"https://api.example.com",
		)
		sh.OAuthConfiguration = oauthConfig
		zero := 0
		one := 1
		loginIDsKeys := map[string]coreconfig.LoginIDKeyConfiguration{
			"email": coreconfig.LoginIDKeyConfiguration{Minimum: &zero, Maximum: &one},
		}
		allowedRealms := []string{password.DefaultRealm}
		passwordAuthProvider := password.NewMockProviderWithPrincipalMap(
			loginIDsKeys,
			allowedRealms,
			map[string]password.Principal{},
		)
		sh.IdentityProvider = principal.NewMockIdentityProvider(sh.OAuthAuthProvider, passwordAuthProvider)
		sh.HookProvider = hook.NewMockProvider()
		nonce := "nonce"
		hashedNonce := crypto.SHA256String(nonce)
		nonceCookie := &http.Cookie{
			Name:  coreHttp.CookieNameOpenIDConnectNonce,
			Value: nonce,
		}

		Convey("should return callback url when ux_mode is web_redirect", func() {
			mockOAuthProvider := oauth.NewMockProvider(nil)
			sh.OAuthAuthProvider = mockOAuthProvider
			uxMode := sso.UXModeWebRedirect

			// oauth state
			state := sso.State{
				OAuthAuthorizationCodeFlowState: sso.OAuthAuthorizationCodeFlowState{
					CallbackURL: "http://localhost:3000",
					UXMode:      uxMode,
					Action:      action,
				},
				LinkState: sso.LinkState{
					UserID: "john.doe.id",
				},
				Nonce: hashedNonce,
			}
			encodedState, _ := sso.EncodeState(stateJWTSecret, state)

			v := url.Values{}
			v.Set("code", "code")
			v.Add("state", encodedState)
			u := url.URL{
				RawQuery: v.Encode(),
			}

			req, _ := http.NewRequest("GET", u.RequestURI(), nil)
			req.AddCookie(nonceCookie)
			resp := httptest.NewRecorder()

			sh.ServeHTTP(resp, req)
			// for web_redirect, it should redirect to original callback url
			So(resp.Code, ShouldEqual, 302)
			So(resp.Header().Get("Location"), ShouldEqual, "http://localhost:3000")

			actual, err := decodeCookie(resp)
			So(err, ShouldBeNil)
			So(actual, ShouldEqualJSON, `
			{
				"callback_url": "http://localhost:3000",
				"result": {
					"result": {}
				}
			}
			`)
		})

		Convey("should get err if user is already linked", func() {
			uxMode := sso.UXModeWebRedirect
			mockOAuthProvider := oauth.NewMockProvider([]*oauth.Principal{
				&oauth.Principal{
					ID:           "jane.doe.id",
					UserID:       "jane.doe.id",
					ProviderType: "google",
					ProviderKeys: map[string]interface{}{},
				},
			})
			sh.OAuthAuthProvider = mockOAuthProvider

			// oauth state
			state := sso.State{
				OAuthAuthorizationCodeFlowState: sso.OAuthAuthorizationCodeFlowState{
					CallbackURL: "http://localhost:3000",
					UXMode:      uxMode,
					Action:      action,
				},
				LinkState: sso.LinkState{
					UserID: "jane.doe.id",
				},
				Nonce: hashedNonce,
			}
			encodedState, _ := sso.EncodeState(stateJWTSecret, state)

			v := url.Values{}
			v.Set("code", "code")
			v.Add("state", encodedState)
			u := url.URL{
				RawQuery: v.Encode(),
			}

			req, _ := http.NewRequest("GET", u.RequestURI(), nil)
			req.AddCookie(nonceCookie)
			resp := httptest.NewRecorder()

			sh.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, 302)
			So(resp.Header().Get("Location"), ShouldEqual, "http://localhost:3000")

			actual, err := decodeCookie(resp)
			So(err, ShouldBeNil)
			So(actual, ShouldEqualJSON, `
			{
				"callback_url": "http://localhost:3000",
				"result": {
					"error": {
						"code": 108,
						"message": "the provider user is already linked",
						"name": "InvalidArgument"
					}
				}
			}
			`)
		})
	})

	Convey("Test OnUserDuplicate", t, func() {
		action := "login"
		UXMode := sso.UXModeWebRedirect
		stateJWTSecret := "secret"
		providerName := "mock"
		providerUserID := "mock_user_id"

		sh := &AuthHandler{}
		sh.TxContext = db.NewMockTxContext()
		sh.AuthContext = auth.NewMockContextGetterWithDefaultUser()
		oauthConfig := coreconfig.OAuthConfiguration{
			URLPrefix:      "http://localhost:3000",
			StateJWTSecret: stateJWTSecret,
			AllowedCallbackURLs: []string{
				"http://localhost",
			},
		}
		providerConfig := coreconfig.OAuthProviderConfiguration{
			ID:           providerName,
			Type:         "google",
			ClientID:     "mock_client_id",
			ClientSecret: "mock_client_secret",
		}
		mockProvider := sso.MockSSOProvider{
			BaseURL:        "http://mock/auth",
			OAuthConfig:    oauthConfig,
			ProviderConfig: providerConfig,
			UserInfo: sso.ProviderUserInfo{ID: providerUserID,
				Email: "john.doe@example.com"},
		}
		sh.Provider = &mockProvider
		mockOAuthProvider := oauth.NewMockProvider(nil)
		sh.OAuthAuthProvider = mockOAuthProvider
		authInfoStore := authinfo.NewMockStoreWithAuthInfoMap(
			map[string]authinfo.AuthInfo{
				"john.doe.id": authinfo.AuthInfo{
					ID:         "john.doe.id",
					VerifyInfo: map[string]bool{},
				},
			},
		)
		sh.AuthInfoStore = authInfoStore
		mockTokenStore := authtoken.NewMockStore()
		sh.TokenStore = mockTokenStore
		profileData := map[string]map[string]interface{}{
			"john.doe.id": map[string]interface{}{},
		}
		sh.UserProfileStore = userprofile.NewMockUserProfileStoreByData(profileData)
		sh.AuthHandlerHTMLProvider = sso.NewAuthHandlerHTMLProvider(
			"https://api.example.com",
		)
		sh.OAuthConfiguration = oauthConfig
		zero := 0
		one := 1
		loginIDsKeys := map[string]coreconfig.LoginIDKeyConfiguration{
			"email": coreconfig.LoginIDKeyConfiguration{
				Type:    coreconfig.LoginIDKeyType(metadata.Email),
				Minimum: &zero,
				Maximum: &one,
			},
		}
		allowedRealms := []string{password.DefaultRealm}
		passwordAuthProvider := password.NewMockProviderWithPrincipalMap(
			loginIDsKeys,
			allowedRealms,
			map[string]password.Principal{
				"john.doe.principal.id": password.Principal{
					ID:             "john.doe.principal.id",
					UserID:         "john.doe.id",
					LoginIDKey:     "email",
					LoginID:        "john.doe@example.com",
					Realm:          "default",
					HashedPassword: []byte("$2a$10$/jm/S1sY6ldfL6UZljlJdOAdJojsJfkjg/pqK47Q8WmOLE19tGWQi"), // 123456
					ClaimsValue: map[string]interface{}{
						"email": "john.doe@example.com",
					},
				},
			},
		)
		sh.IdentityProvider = principal.NewMockIdentityProvider(sh.OAuthAuthProvider, passwordAuthProvider)
		sh.HookProvider = hook.NewMockProvider()
		nonce := "nonce"
		hashedNonce := crypto.SHA256String(nonce)
		nonceCookie := &http.Cookie{
			Name:  coreHttp.CookieNameOpenIDConnectNonce,
			Value: nonce,
		}

		Convey("OnUserDuplicate == abort", func() {
			state := sso.State{
				OAuthAuthorizationCodeFlowState: sso.OAuthAuthorizationCodeFlowState{
					CallbackURL: "http://localhost:3000",
					UXMode:      UXMode,
					Action:      action,
				},
				LoginState: sso.LoginState{
					MergeRealm:      password.DefaultRealm,
					OnUserDuplicate: sso.OnUserDuplicateAbort,
				},
				Nonce: hashedNonce,
			}
			encodedState, _ := sso.EncodeState(stateJWTSecret, state)

			v := url.Values{}
			v.Set("code", "code")
			v.Add("state", encodedState)
			u := url.URL{
				RawQuery: v.Encode(),
			}
			req, _ := http.NewRequest("GET", u.RequestURI(), nil)
			req.AddCookie(nonceCookie)
			resp := httptest.NewRecorder()
			sh.ServeHTTP(resp, req)

			So(resp.Code, ShouldEqual, 302)

			actual, err := decodeCookie(resp)
			So(err, ShouldBeNil)
			So(actual, ShouldEqualJSON, `
			{
				"callback_url": "http://localhost:3000",
				"result": {
					"error": {
						"code": 109,
						"message": "Aborted due to duplicate user",
						"name": "Duplicated"
					}
				}
			}
			`)
		})

		Convey("OnUserDuplicate == merge", func() {
			state := sso.State{
				OAuthAuthorizationCodeFlowState: sso.OAuthAuthorizationCodeFlowState{
					CallbackURL: "http://localhost:3000",
					UXMode:      UXMode,
					Action:      action,
				},
				LoginState: sso.LoginState{
					MergeRealm:      password.DefaultRealm,
					OnUserDuplicate: sso.OnUserDuplicateMerge,
				},
				Nonce: hashedNonce,
			}
			encodedState, _ := sso.EncodeState(stateJWTSecret, state)

			v := url.Values{}
			v.Set("code", "code")
			v.Add("state", encodedState)
			u := url.URL{
				RawQuery: v.Encode(),
			}
			req, _ := http.NewRequest("GET", u.RequestURI(), nil)
			req.AddCookie(nonceCookie)
			resp := httptest.NewRecorder()
			sh.ServeHTTP(resp, req)

			So(resp.Code, ShouldEqual, 302)

			actual, err := decodeCookie(resp)
			So(err, ShouldBeNil)
			p, _ := sh.OAuthAuthProvider.GetPrincipalByProvider(oauth.GetByProviderOptions{
				ProviderType:   "google",
				ProviderUserID: providerUserID,
			})
			token := mockTokenStore.GetTokensByAuthInfoID(p.UserID)[0]
			So(actual, ShouldEqualJSON, fmt.Sprintf(`
			{
				"callback_url": "http://localhost:3000",
				"result": {
					"result": {
						"user": {
							"created_at": "0001-01-01T00:00:00Z",
							"id": "%s",
							"is_disabled": false,
							"is_verified": false,
							"metadata": {},
							"verify_info": {}
						},
						"identity": {
							"type": "oauth",
							"id": "%s",
							"provider_type": "google",
							"provider_user_id": "%s",
							"raw_profile": {
								"id": "%s",
								"email": "john.doe@example.com"
							},
							"claims": {
								"email": "john.doe@example.com"
							}
						},
						"access_token": "%s"
					}
				}
			}
			`, p.UserID,
				p.ID,
				providerUserID,
				providerUserID,
				token.AccessToken))
		})

		Convey("OnUserDuplicate == create", func() {
			state := sso.State{
				OAuthAuthorizationCodeFlowState: sso.OAuthAuthorizationCodeFlowState{
					CallbackURL: "http://localhost:3000",
					UXMode:      UXMode,
					Action:      action,
				},
				LoginState: sso.LoginState{
					MergeRealm:      password.DefaultRealm,
					OnUserDuplicate: sso.OnUserDuplicateCreate,
				},
				Nonce: hashedNonce,
			}
			encodedState, _ := sso.EncodeState(stateJWTSecret, state)

			v := url.Values{}
			v.Set("code", "code")
			v.Add("state", encodedState)
			u := url.URL{
				RawQuery: v.Encode(),
			}
			req, _ := http.NewRequest("GET", u.RequestURI(), nil)
			req.AddCookie(nonceCookie)
			resp := httptest.NewRecorder()
			sh.ServeHTTP(resp, req)

			So(resp.Code, ShouldEqual, 302)

			actual, err := decodeCookie(resp)
			So(err, ShouldBeNil)
			p, _ := sh.OAuthAuthProvider.GetPrincipalByProvider(oauth.GetByProviderOptions{
				ProviderType:   "google",
				ProviderUserID: providerUserID,
			})
			So(p.UserID, ShouldNotEqual, "john.doe.id")
			token := mockTokenStore.GetTokensByAuthInfoID(p.UserID)[0]
			So(actual, ShouldEqualJSON, fmt.Sprintf(`
			{
				"callback_url": "http://localhost:3000",
				"result": {
					"result": {
						"user": {
							"created_at": "0001-01-01T00:00:00Z",
							"id": "%s",
							"is_disabled": false,
							"is_verified": false,
							"last_login_at": "2006-01-02T15:04:05Z",
							"metadata": {},
							"verify_info": {}
						},
						"identity": {
							"type": "oauth",
							"id": "%s",
							"provider_type": "google",
							"provider_user_id": "%s",
							"raw_profile": {
								"id": "%s",
								"email": "john.doe@example.com"
							},
							"claims": {
								"email": "john.doe@example.com"
							}
						},
						"access_token": "%s"
					}
				}
			}
			`, p.UserID,
				p.ID,
				providerUserID,
				providerUserID,
				token.AccessToken))
		})
	})
}

func TestValidateCallbackURL(t *testing.T) {
	Convey("Test ValidateCallbackURL", t, func() {
		sh := &AuthHandler{}
		callbackURL := "http://localhost:3000"
		allowedCallbackURLs := []string{
			"http://localhost",
			"http://127.0.0.1",
		}

		e := sh.validateCallbackURL(allowedCallbackURLs, callbackURL)
		So(e, ShouldBeNil)

		callbackURL = "http://oursky"
		e = sh.validateCallbackURL(allowedCallbackURLs, callbackURL)
		So(e, ShouldNotBeNil)

		e = sh.validateCallbackURL(allowedCallbackURLs, "")
		So(e, ShouldNotBeNil)
	})
}
