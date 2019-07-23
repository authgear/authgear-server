package userverify

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	gotime "time"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/time"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userverify"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	. "github.com/skygeario/skygear-server/pkg/core/skytest"
	. "github.com/smartystreets/goconvey/convey"
)

func TestForgotPasswordResetHandler(t *testing.T) {
	Convey("Test VerifyCodeHandler", t, func() {
		time := time.MockProvider{TimeNowUTC: gotime.Date(2006, 1, 2, 15, 4, 5, 0, gotime.UTC)}

		vh := &VerifyCodeHandler{}
		logger, _ := test.NewNullLogger()
		vh.Logger = logrus.NewEntry(logger)
		vh.TxContext = db.NewMockTxContext()
		vh.AuthContext = auth.NewMockContextGetterWithUnverifiedUser(map[string]bool{
			"faseng.cat.id@example.com": false,
		})

		zero := 0
		one := 1
		loginIDsKeys := map[string]config.LoginIDKeyConfiguration{
			"email": config.LoginIDKeyConfiguration{Minimum: &zero, Maximum: &one},
		}
		vh.PasswordAuthProvider = password.NewMockProviderWithPrincipalMap(
			loginIDsKeys,
			[]string{password.DefaultRealm},
			map[string]password.Principal{
				"faseng1": password.Principal{
					ID:             "id1",
					UserID:         "faseng.cat.id",
					LoginIDKey:     "email",
					LoginID:        "faseng.cat.id@example.com",
					Realm:          "default",
					HashedPassword: []byte("$2a$10$/jm/S1sY6ldfL6UZljlJdOAdJojsJfkjg/pqK47Q8WmOLE19tGWQi"), // 123456
				},
				"faseng2": password.Principal{
					ID:             "id2",
					UserID:         "faseng.cat.id",
					LoginIDKey:     "phone",
					LoginID:        "+85299999999",
					Realm:          "default",
					HashedPassword: []byte("$2a$10$/jm/S1sY6ldfL6UZljlJdOAdJojsJfkjg/pqK47Q8WmOLE19tGWQi"), // 123456
				},
				"chima1": password.Principal{
					ID:             "id2",
					UserID:         "chima.cat.id",
					LoginIDKey:     "email",
					LoginID:        "chima.cat.id@example.com",
					Realm:          "default",
					HashedPassword: []byte("$2a$10$/jm/S1sY6ldfL6UZljlJdOAdJojsJfkjg/pqK47Q8WmOLE19tGWQi"), // 123456
				},
			},
		)

		authInfo := authinfo.AuthInfo{
			ID: "faseng.cat.id",
		}
		authInfoStore := authinfo.NewMockStoreWithAuthInfoMap(
			map[string]authinfo.AuthInfo{
				"faseng.cat.id": authInfo,
			},
		)
		vh.AuthInfoStore = authInfoStore
		vh.UserProfileStore = userprofile.NewMockUserProfileStore()
		vh.HookProvider = hook.NewMockProvider()

		verifyConfig := config.UserVerificationConfiguration{
			Criteria: config.UserVerificationCriteriaAll,
			LoginIDKeys: map[string]config.UserVerificationKeyConfiguration{
				"email": config.UserVerificationKeyConfiguration{
					Expiry: 12 * 60 * 60,
				},
			},
		}
		store := userverify.MockStore{
			CodeByID: []userverify.VerifyCode{
				userverify.VerifyCode{
					ID:         "code",
					UserID:     "faseng.cat.id",
					LoginIDKey: "email",
					LoginID:    "faseng.cat.id@example.com",
					Code:       "C0DE1",
					Consumed:   false,
					CreatedAt:  time.NowUTC(),
				},
				userverify.VerifyCode{
					ID:         "code",
					UserID:     "faseng.cat.id",
					LoginIDKey: "email",
					LoginID:    "faseng.cat.id@example.com",
					Code:       "C0DE2",
					Consumed:   false,
					CreatedAt:  time.NowUTC().Add(-gotime.Duration(1) * gotime.Hour),
				},
				userverify.VerifyCode{
					ID:         "code1",
					UserID:     "chima.cat.id",
					LoginIDKey: "email",
					LoginID:    "chima.cat.id@example.com",
					Code:       "C0DE3",
					Consumed:   false,
					CreatedAt:  time.NowUTC(),
				},
			},
		}
		vh.UserVerificationProvider = userverify.NewProvider(nil, &store, verifyConfig, time)

		Convey("verify with correct code and auto update", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"code": "code1"
			}`))
			resp := httptest.NewRecorder()
			h := handler.APIHandlerToHandler(vh, vh.TxContext)
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": {}
			}`)
			So(authInfoStore.AuthInfoMap["faseng.cat.id"].Verified, ShouldBeTrue)
		})

		Convey("verify with correct code but not all verified", func() {
			newVerifyConfig := verifyConfig
			newVerifyConfig.LoginIDKeys = map[string]config.UserVerificationKeyConfiguration{
				"email": config.UserVerificationKeyConfiguration{Expiry: 12 * 60 * 60},
				"phone": config.UserVerificationKeyConfiguration{Expiry: 12 * 60 * 60},
			}
			provider := userverify.NewProvider(nil, &store, newVerifyConfig, time)
			oldProvider := vh.UserVerificationProvider
			vh.UserVerificationProvider = provider
			defer func() {
				vh.UserVerificationProvider = oldProvider
			}()

			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"code": "code1"
			}`))
			resp := httptest.NewRecorder()
			h := handler.APIHandlerToHandler(vh, vh.TxContext)
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": {}
			}`)
			So(authInfoStore.AuthInfoMap["faseng.cat.id"].Verified, ShouldBeFalse)
		})

		Convey("verify with expired code", func() {
			code := store.CodeByID[0]
			code.CreatedAt = time.NowUTC().Add(-gotime.Duration(100) * gotime.Hour)
			store.CodeByID[0] = code

			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"code": "code1"
			}`))
			resp := httptest.NewRecorder()
			h := handler.APIHandlerToHandler(vh, vh.TxContext)
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"error": {
					"code": 108,
					"message": "the code has expired",
					"name": "InvalidArgument"
				}
			}`)
			So(authInfoStore.AuthInfoMap["faseng.cat.id"].Verified, ShouldBeFalse)
		})

		Convey("verify with past generated code", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"code": "code2"
			}`))
			resp := httptest.NewRecorder()
			h := handler.APIHandlerToHandler(vh, vh.TxContext)
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"error": {
					"code": 108,
					"message": "invalid verification code",
					"name": "InvalidArgument"
				}
			}`)
			So(authInfoStore.AuthInfoMap["faseng.cat.id"].Verified, ShouldBeFalse)
		})

		Convey("verify with someone else code", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"code": "code3"
			}`))
			resp := httptest.NewRecorder()
			h := handler.APIHandlerToHandler(vh, vh.TxContext)
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"error": {
					"code": 108,
					"message": "invalid verification code",
					"name": "InvalidArgument"
				}
			}`)
			So(authInfoStore.AuthInfoMap["faseng.cat.id"].Verified, ShouldBeFalse)
		})

		Convey("verify with consumed code", func() {
			code := store.CodeByID[0]
			code.Consumed = true
			store.CodeByID[0] = code

			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"code": "code1"
			}`))
			resp := httptest.NewRecorder()
			h := handler.APIHandlerToHandler(vh, vh.TxContext)
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"error": {
					"code": 108,
					"message": "invalid verification code",
					"name": "InvalidArgument"
				}
			}`)
			So(authInfoStore.AuthInfoMap["faseng.cat.id"].Verified, ShouldBeFalse)
		})

		Convey("verify with incorrect code", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"code": "incorrect"
			}`))
			resp := httptest.NewRecorder()
			h := handler.APIHandlerToHandler(vh, vh.TxContext)
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"error": {
					"code": 108,
					"message": "invalid verification code",
					"name": "InvalidArgument"
				}
			}`)
			So(authInfoStore.AuthInfoMap["faseng.cat.id"].Verified, ShouldBeFalse)
		})
	})
}
