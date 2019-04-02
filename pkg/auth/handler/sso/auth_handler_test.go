package sso

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
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
	"github.com/skygeario/skygear-server/pkg/core/skyerr"

	. "github.com/skygeario/skygear-server/pkg/core/skytest"
	. "github.com/smartystreets/goconvey/convey"
)

func TestAuthPayload(t *testing.T) {
	Convey("Test AuthRequestPayload", t, func() {
		// callback URL and ux_mode is required
		Convey("validate valid payload", func() {
			payload := AuthRequestPayload{
				Code:         "code",
				EncodedState: "state",
			}
			So(payload.Validate(), ShouldBeNil)
		})

		Convey("validate payload without code", func() {
			payload := AuthRequestPayload{
				EncodedState: "state",
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
	timeNow = func() time.Time { return zeroTime }
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
		sh.UserProfileStore = userprofile.NewMockUserProfileStore()
		sh.AuthHandlerHTMLProvider = sso.NewAuthHandlerHTMLProvider(
			"https://api.example.com",
			"https://api.example.com/skygear.js",
		)
		sh.SSOSetting = setting
		loginIDsKeyWhitelist := []string{"email"}
		passwordAuthProvider := password.NewMockProviderWithPrincipalMap(
			loginIDsKeyWhitelist,
			map[string]password.Principal{},
		)
		sh.PasswordAuthProvider = passwordAuthProvider

		Convey("should return callback url when ux_mode is web_redirect", func() {
			UXMode := "web_redirect"

			// oauth state
			state := sso.State{
				CallbackURL: "http://localhost:3000",
				UXMode:      UXMode,
				Action:      action,
			}
			encodedState, _ := sso.EncodeState(stateJWTSecret, state)

			v := url.Values{}
			v.Set("code", "code")
			v.Add("state", encodedState)
			u := url.URL{
				RawQuery: v.Encode(),
			}

			req, _ := http.NewRequest("GET", u.RequestURI(), nil)
			resp := httptest.NewRecorder()

			sh.ServeHTTP(resp, req)
			// for web_redirect, it should redirect to original callback url
			So(resp.Code, ShouldEqual, 302)
			So(resp.Header().Get("Location"), ShouldEqual, "http://localhost:3000")

			// check cookies
			// it should have following format
			// {
			// 	"callback_url": "callback_url"
			// 	"result": authResp
			// }
			cookies := resp.Result().Cookies()
			So(cookies, ShouldNotBeEmpty)
			var ssoDataCookie *http.Cookie
			for _, c := range cookies {
				if c.Name == "sso_data" {
					ssoDataCookie = c
					break
				}
			}
			So(ssoDataCookie, ShouldNotBeNil)

			// decoded it first
			decoded, err := base64.StdEncoding.DecodeString(ssoDataCookie.Value)
			So(err, ShouldBeNil)
			So(decoded, ShouldNotBeNil)

			// Unmarshal to map
			data := make(map[string]interface{})
			err = json.Unmarshal(decoded, &data)
			So(err, ShouldBeNil)

			// check callback_url
			So(data["callback_url"], ShouldEqual, "http://localhost:3000")

			// check result(authResp)
			authResp, err := json.Marshal(data["result"])
			So(err, ShouldBeNil)
			p, err := sh.OAuthAuthProvider.GetPrincipalByProviderUserID(providerName, providerUserID)
			So(err, ShouldBeNil)
			token := mockTokenStore.GetTokensByAuthInfoID(p.UserID)[0]
			So(authResp, ShouldEqualJSON, fmt.Sprintf(`{
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

		Convey("should return html page when ux_mode is web_popup", func() {
			UXMode := "web_popup"

			// oauth state
			state := sso.State{
				CallbackURL: "http://localhost:3000",
				UXMode:      UXMode,
				Action:      action,
			}
			encodedState, _ := sso.EncodeState(stateJWTSecret, state)

			v := url.Values{}
			v.Set("code", "code")
			v.Add("state", encodedState)
			u := url.URL{
				RawQuery: v.Encode(),
			}

			req, _ := http.NewRequest("GET", u.RequestURI(), nil)
			resp := httptest.NewRecorder()

			sh.ServeHTTP(resp, req)
			// for web_redirect, it should redirect to original callback url
			So(resp.Code, ShouldEqual, 200)
			JSSKDURLPattern := `<script type="text/javascript" src="https://api.example.com/skygear.js"></script>`
			matched, err := regexp.MatchString(JSSKDURLPattern, resp.Body.String())
			So(err, ShouldBeNil)
			So(matched, ShouldBeTrue)
		})

		Convey("should return callback url with result query parameter when ux_mode is ios or android", func() {
			UXMode := "ios"

			// oauth state
			state := sso.State{
				CallbackURL: "http://localhost:3000",
				UXMode:      UXMode,
				Action:      action,
			}
			encodedState, _ := sso.EncodeState(stateJWTSecret, state)

			v := url.Values{}
			v.Set("code", "code")
			v.Add("state", encodedState)
			u := url.URL{
				RawQuery: v.Encode(),
			}

			req, _ := http.NewRequest("GET", u.RequestURI(), nil)
			resp := httptest.NewRecorder()

			sh.ServeHTTP(resp, req)
			// for ios or android, it should redirect to original callback url
			So(resp.Code, ShouldEqual, 302)
			// check location result query parameter
			location, _ := url.Parse(resp.Header().Get("Location"))
			q := location.Query()
			result := q.Get("result")
			decoded, _ := base64.StdEncoding.DecodeString(result)
			p, _ := sh.OAuthAuthProvider.GetPrincipalByProviderUserID(providerName, providerUserID)
			token := mockTokenStore.GetTokensByAuthInfoID(p.UserID)[0]
			So(decoded, ShouldEqualJSON, fmt.Sprintf(`{
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

	Convey("Test AuthHandler with link action", t, func() {
		action := "link"
		stateJWTSecret := "secret"
		sh := &AuthHandler{}
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
			Name:         "mock",
			ClientID:     "mock_client_id",
			ClientSecret: "mock_client_secret",
		}
		mockProvider := sso.MockSSOProverImpl{
			BaseURL: "http://mock/auth",
			Setting: setting,
			Config:  config,
		}
		sh.Provider = &mockProvider
		mockOAuthProvider := oauth.NewMockProvider(
			map[string]string{
				"jane.doe.id": "jane.doe.id",
			},
			map[string]oauth.Principal{
				"jane.doe.id": oauth.Principal{},
			},
		)
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
		sh.AuthHandlerHTMLProvider = sso.NewAuthHandlerHTMLProvider(
			"https://api.example.com",
			"https://api.example.com/skygear.js",
		)
		sh.SSOSetting = setting
		loginIDsKeyWhitelist := []string{"email"}
		passwordAuthProvider := password.NewMockProviderWithPrincipalMap(
			loginIDsKeyWhitelist,
			map[string]password.Principal{},
		)
		sh.PasswordAuthProvider = passwordAuthProvider

		Convey("should return callback url when ux_mode is web_redirect", func() {
			UXMode := "web_redirect"

			// oauth state
			state := sso.State{
				CallbackURL: "http://localhost:3000",
				UXMode:      UXMode,
				Action:      action,
				UserID:      "john.doe.id",
			}
			encodedState, _ := sso.EncodeState(stateJWTSecret, state)

			v := url.Values{}
			v.Set("code", "code")
			v.Add("state", encodedState)
			u := url.URL{
				RawQuery: v.Encode(),
			}

			req, _ := http.NewRequest("GET", u.RequestURI(), nil)
			resp := httptest.NewRecorder()

			sh.ServeHTTP(resp, req)
			// for web_redirect, it should redirect to original callback url
			So(resp.Code, ShouldEqual, 302)
			So(resp.Header().Get("Location"), ShouldEqual, "http://localhost:3000")

			// check cookies
			// it should have following format
			// {
			// 	"callback_url": "callback_url"
			// 	"result": authResp
			// }
			cookies := resp.Result().Cookies()
			So(cookies, ShouldNotBeEmpty)
			var ssoDataCookie *http.Cookie
			for _, c := range cookies {
				if c.Name == "sso_data" {
					ssoDataCookie = c
					break
				}
			}
			So(ssoDataCookie, ShouldNotBeNil)

			// decoded it first
			decoded, err := base64.StdEncoding.DecodeString(ssoDataCookie.Value)
			So(err, ShouldBeNil)
			So(decoded, ShouldNotBeNil)

			// Unmarshal to map
			data := make(map[string]interface{})
			err = json.Unmarshal(decoded, &data)
			So(err, ShouldBeNil)

			// check callback_url
			So(data["callback_url"], ShouldEqual, "http://localhost:3000")

			// check result(resp)
			result, err := json.Marshal(data["result"])
			So(err, ShouldBeNil)
			So(string(result), ShouldEqualJSON, `{
				"result": "OK"
			}`)
		})

		Convey("should get err if user is already linked", func() {
			UXMode := "web_redirect"

			// oauth state
			state := sso.State{
				CallbackURL: "http://localhost:3000",
				UXMode:      UXMode,
				Action:      action,
				UserID:      "jane.doe.id",
			}
			encodedState, _ := sso.EncodeState(stateJWTSecret, state)

			v := url.Values{}
			v.Set("code", "code")
			v.Add("state", encodedState)
			u := url.URL{
				RawQuery: v.Encode(),
			}

			req, _ := http.NewRequest("GET", u.RequestURI(), nil)
			resp := httptest.NewRecorder()

			sh.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, 302)
			So(resp.Header().Get("Location"), ShouldEqual, "http://localhost:3000")

			// check cookies
			// it should have following format
			// {
			// 	"callback_url": "callback_url"
			// 	"result": errorAuthResp
			// }
			cookies := resp.Result().Cookies()
			So(cookies, ShouldNotBeEmpty)
			var ssoDataCookie *http.Cookie
			for _, c := range cookies {
				if c.Name == "sso_data" {
					ssoDataCookie = c
					break
				}
			}
			So(ssoDataCookie, ShouldNotBeNil)

			// decoded it first
			decoded, err := base64.StdEncoding.DecodeString(ssoDataCookie.Value)
			So(err, ShouldBeNil)
			So(decoded, ShouldNotBeNil)

			// Unmarshal to map
			data := make(map[string]interface{})
			err = json.Unmarshal(decoded, &data)
			So(err, ShouldBeNil)

			// check callback_url
			So(data["callback_url"], ShouldEqual, "http://localhost:3000")

			// check result(resp)
			result, err := json.Marshal(data["result"])
			So(err, ShouldBeNil)
			So(string(result), ShouldEqualJSON, `{
				"error":{"code":108,"message":"provider account already linked with existing user","name":"InvalidArgument"}
			}`)
		})
	})

	Convey("Test AuthHandler's auto link procedure", t, func() {
		action := "login"
		UXMode := "web_redirect"

		stateJWTSecret := "secret"
		providerName := "mock"
		providerUserID := "mock_user_id"
		sh := &AuthHandler{}
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
			AuthData: map[string]string{
				"email": "john.doe@example.com",
			},
		}
		sh.Provider = &mockProvider
		mockOAuthProvider := oauth.NewMockProvider(
			map[string]string{},
			map[string]oauth.Principal{},
		)
		sh.OAuthAuthProvider = mockOAuthProvider
		authInfoStore := authinfo.NewMockStoreWithAuthInfoMap(
			map[string]authinfo.AuthInfo{
				"john.doe.id": authinfo.AuthInfo{
					ID: "john.doe.id",
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
			"https://api.example.com/skygear.js",
		)
		sh.SSOSetting = setting
		loginIDsKeyWhitelist := []string{"email"}
		passwordAuthProvider := password.NewMockProviderWithPrincipalMap(
			loginIDsKeyWhitelist,
			map[string]password.Principal{
				"john.doe.principal.id": password.Principal{
					ID:             "john.doe.principal.id",
					UserID:         "john.doe.id",
					LoginIDKey:     "email",
					LoginID:        "john.doe@example.com",
					HashedPassword: []byte("$2a$10$/jm/S1sY6ldfL6UZljlJdOAdJojsJfkjg/pqK47Q8WmOLE19tGWQi"), // 123456
				},
			},
		)
		sh.PasswordAuthProvider = passwordAuthProvider

		Convey("should auto-link password principal", func() {
			// oauth state
			state := sso.State{
				CallbackURL: "http://localhost:3000",
				UXMode:      UXMode,
				Action:      action,
			}
			encodedState, _ := sso.EncodeState(stateJWTSecret, state)

			v := url.Values{}
			v.Set("code", "code")
			v.Add("state", encodedState)
			u := url.URL{
				RawQuery: v.Encode(),
			}

			req, _ := http.NewRequest("GET", u.RequestURI(), nil)
			resp := httptest.NewRecorder()

			sh.ServeHTTP(resp, req)

			oauthPrincipal, err := sh.OAuthAuthProvider.GetPrincipalByProviderUserID(providerName, providerUserID)
			So(err, ShouldBeNil)
			So(oauthPrincipal.UserID, ShouldEqual, "john.doe.id")
		})
	})

	Convey("Test AuthHandler's auto link procedure", t, func() {
		action := "login"
		UXMode := "web_redirect"

		stateJWTSecret := "secret"
		providerName := "mock"
		providerUserID := "mock_user_id"
		sh := &AuthHandler{}
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
			AuthData: map[string]string{
				"email": "john.doe@example.com",
			},
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
		sh.UserProfileStore = userprofile.NewMockUserProfileStore()
		sh.AuthHandlerHTMLProvider = sso.NewAuthHandlerHTMLProvider(
			"https://api.example.com",
			"https://api.example.com/skygear.js",
		)
		sh.SSOSetting = setting
		loginIDsKeyWhitelist := []string{"email"}
		passwordAuthProvider := password.NewMockProviderWithPrincipalMap(
			loginIDsKeyWhitelist,
			map[string]password.Principal{},
		)
		sh.PasswordAuthProvider = passwordAuthProvider

		Convey("should also create an empty password principal", func() {
			// oauth state
			state := sso.State{
				CallbackURL: "http://localhost:3000",
				UXMode:      UXMode,
				Action:      action,
			}
			encodedState, _ := sso.EncodeState(stateJWTSecret, state)

			v := url.Values{}
			v.Set("code", "code")
			v.Add("state", encodedState)
			u := url.URL{
				RawQuery: v.Encode(),
			}

			req, _ := http.NewRequest("GET", u.RequestURI(), nil)
			resp := httptest.NewRecorder()

			sh.ServeHTTP(resp, req)

			oauthPrincipal, _ := sh.OAuthAuthProvider.GetPrincipalByProviderUserID(providerName, providerUserID)
			_, err := sh.PasswordAuthProvider.GetPrincipalsByUserID(oauthPrincipal.UserID)
			So(err, ShouldBeNil)
		})
	})

	Convey("Test AuthHandler's auto link procedure", t, func() {
		action := "login"
		UXMode := "web_redirect"

		stateJWTSecret := "secret"
		providerName := "mock"
		providerUserID := "mock_user_id"
		sh := &AuthHandler{}
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
			AuthData: map[string]string{
				"email": "john.doe@example.com",
			},
		}
		sh.Provider = &mockProvider
		mockOAuthProvider := oauth.NewMockProvider(
			map[string]string{},
			map[string]oauth.Principal{},
		)
		sh.OAuthAuthProvider = mockOAuthProvider
		authInfoStore := authinfo.NewMockStoreWithAuthInfoMap(
			map[string]authinfo.AuthInfo{
				"john.doe.id": authinfo.AuthInfo{
					ID: "john.doe.id",
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
			"https://api.example.com/skygear.js",
		)
		sh.SSOSetting = setting
		// providerLoginID wouldn't match loginIDsKeyWhitelist "["username"]"
		loginIDsKeyWhitelist := []string{"username"}
		passwordAuthProvider := password.NewMockProviderWithPrincipalMap(
			loginIDsKeyWhitelist,
			map[string]password.Principal{
				"john.doe.principal.id": password.Principal{
					ID:             "john.doe.principal.id",
					UserID:         "john.doe.id",
					LoginIDKey:     "username",
					LoginID:        "john.doe",
					HashedPassword: []byte("$2a$10$/jm/S1sY6ldfL6UZljlJdOAdJojsJfkjg/pqK47Q8WmOLE19tGWQi"), // 123456
				},
			},
		)
		sh.PasswordAuthProvider = passwordAuthProvider

		Convey("shouldn't auto-link password principal if loginIDsKeyWhitelist not matched", func() {
			// oauth state
			state := sso.State{
				CallbackURL: "http://localhost:3000",
				UXMode:      UXMode,
				Action:      action,
			}
			encodedState, _ := sso.EncodeState(stateJWTSecret, state)

			v := url.Values{}
			v.Set("code", "code")
			v.Add("state", encodedState)
			u := url.URL{
				RawQuery: v.Encode(),
			}

			req, _ := http.NewRequest("GET", u.RequestURI(), nil)
			resp := httptest.NewRecorder()

			sh.ServeHTTP(resp, req)

			oauthPrincipal, err := sh.OAuthAuthProvider.GetPrincipalByProviderUserID(providerName, providerUserID)
			So(err, ShouldBeNil)
			// should signup a new user
			So(oauthPrincipal.UserID, ShouldNotEqual, "john.doe.id")
			// empty password should not be created
			_, err = sh.PasswordAuthProvider.GetPrincipalsByUserID(oauthPrincipal.UserID)
			So(err, ShouldNotBeNil)
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
