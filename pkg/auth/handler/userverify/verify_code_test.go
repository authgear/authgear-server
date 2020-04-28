package userverify

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	gotime "time"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"

	authtesting "github.com/skygeario/skygear-server/pkg/auth/dependency/auth/testing"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity/loginid"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userverify"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	. "github.com/skygeario/skygear-server/pkg/core/skytest"
	"github.com/skygeario/skygear-server/pkg/core/time"
	"github.com/skygeario/skygear-server/pkg/core/validation"
	. "github.com/smartystreets/goconvey/convey"
)

func TestVerifyCodeHandler(t *testing.T) {
	Convey("Test VerifyCodeHandler", t, func() {
		time := time.MockProvider{TimeNowUTC: gotime.Date(2006, 1, 2, 15, 4, 5, 0, gotime.UTC)}

		vh := &VerifyCodeHandler{}
		validator := validation.NewValidator("http://v2.skygear.io")
		validator.AddSchemaFragments(
			VerifyCodeRequestSchema,
		)
		vh.Validator = validator
		logger, _ := test.NewNullLogger()
		vh.Logger = logrus.NewEntry(logger)
		vh.TxContext = db.NewMockTxContext()

		vh.LoginIDProvider = &mockLoginIDProvider{
			Identities: []loginid.Identity{
				{
					ID:         "id1",
					UserID:     "faseng.cat.id",
					LoginIDKey: "email",
					LoginID:    "faseng.cat.id@example.com",
				},
				{
					ID:         "id2",
					UserID:     "faseng.cat.id",
					LoginIDKey: "phone",
					LoginID:    "+85299999999",
				},
				{
					ID:         "id3",
					UserID:     "chima.cat.id",
					LoginIDKey: "email",
					LoginID:    "chima.cat.id@example.com",
				},
			},
		}

		authInfo := authinfo.AuthInfo{
			ID:         "faseng.cat.id",
			VerifyInfo: map[string]bool{},
		}
		authInfoStore := authinfo.NewMockStoreWithAuthInfoMap(
			map[string]authinfo.AuthInfo{
				"faseng.cat.id": authInfo,
			},
		)
		vh.AuthInfoStore = authInfoStore
		vh.UserProfileStore = userprofile.NewMockUserProfileStore()
		hookProvider := hook.NewMockProvider()
		vh.HookProvider = hookProvider

		verifyConfig := &config.UserVerificationConfiguration{
			Criteria: config.UserVerificationCriteriaAll,
			LoginIDKeys: []config.UserVerificationKeyConfiguration{
				config.UserVerificationKeyConfiguration{
					Key:    "email",
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
		vh.UserVerificationProvider = userverify.NewProvider(nil, &store, verifyConfig, &time)

		Convey("verify with correct code and auto update", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"code": "code1"
			}`))
			req = authtesting.WithAuthn().
				UserID("faseng.cat.id").
				ToRequest(req)
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			vh.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": {}
			}`)
			So(authInfoStore.AuthInfoMap["faseng.cat.id"].Verified, ShouldBeTrue)

			isVerified := true
			So(hookProvider.DispatchedEvents, ShouldResemble, []event.Payload{
				event.UserUpdateEvent{
					Reason:     event.UserUpdateReasonVerification,
					IsVerified: &isVerified,
					VerifyInfo: &map[string]bool{
						"faseng.cat.id@example.com": true,
					},
					User: model.User{
						ID:         "faseng.cat.id",
						Verified:   false,
						Disabled:   false,
						VerifyInfo: map[string]bool{},
						Metadata:   userprofile.Data{},
					},
				},
			})
		})

		Convey("verify with correct code but not all verified", func() {
			newVerifyConfig := verifyConfig
			newVerifyConfig.LoginIDKeys = []config.UserVerificationKeyConfiguration{
				config.UserVerificationKeyConfiguration{Key: "email", Expiry: 12 * 60 * 60},
				config.UserVerificationKeyConfiguration{Key: "phone", Expiry: 12 * 60 * 60},
			}
			provider := userverify.NewProvider(nil, &store, newVerifyConfig, &time)
			oldProvider := vh.UserVerificationProvider
			vh.UserVerificationProvider = provider
			defer func() {
				vh.UserVerificationProvider = oldProvider
			}()

			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"code": "code1"
			}`))
			req = authtesting.WithAuthn().
				UserID("faseng.cat.id").
				ToRequest(req)
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			vh.ServeHTTP(resp, req)
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
			req = authtesting.WithAuthn().
				UserID("faseng.cat.id").
				ToRequest(req)
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			vh.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"error": {
					"name": "Invalid",
					"reason": "UserVerificationFailed",
					"message": "verification code has expired",
					"code": 400,
					"info": { "cause": { "kind": "ExpiredCode" } }
				}
			}`)
			So(authInfoStore.AuthInfoMap["faseng.cat.id"].Verified, ShouldBeFalse)
		})

		Convey("verify with past generated code", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"code": "code2"
			}`))
			req = authtesting.WithAuthn().
				UserID("faseng.cat.id").
				ToRequest(req)
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			vh.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"error": {
					"name": "Invalid",
					"reason": "UserVerificationFailed",
					"message": "invalid verification code",
					"code": 400,
					"info": { "cause": { "kind": "InvalidCode" } }
				}
			}`)
			So(authInfoStore.AuthInfoMap["faseng.cat.id"].Verified, ShouldBeFalse)
		})

		Convey("verify with someone else code", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"code": "code3"
			}`))
			req = authtesting.WithAuthn().
				UserID("faseng.cat.id").
				ToRequest(req)
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			vh.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"error": {
					"name": "Invalid",
					"reason": "UserVerificationFailed",
					"message": "invalid verification code",
					"code": 400,
					"info": { "cause": { "kind": "InvalidCode" } }
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
			req = authtesting.WithAuthn().
				UserID("faseng.cat.id").
				ToRequest(req)
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			vh.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"error": {
					"name": "Invalid",
					"reason": "UserVerificationFailed",
					"message": "verification code is used",
					"code": 400,
					"info": { "cause": { "kind": "UsedCode" } }
				}
			}`)
			So(authInfoStore.AuthInfoMap["faseng.cat.id"].Verified, ShouldBeFalse)
		})

		Convey("verify with incorrect code", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"code": "incorrect"
			}`))
			req = authtesting.WithAuthn().
				UserID("faseng.cat.id").
				ToRequest(req)
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			vh.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"error": {
					"name": "Invalid",
					"reason": "UserVerificationFailed",
					"message": "invalid verification code",
					"code": 400,
					"info": { "cause": { "kind": "InvalidCode" } }
				}
			}`)
			So(authInfoStore.AuthInfoMap["faseng.cat.id"].Verified, ShouldBeFalse)
		})
	})
}
