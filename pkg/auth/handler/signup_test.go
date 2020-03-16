package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/authnsession"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/mfa"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/core/async"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/metadata"
	"github.com/skygeario/skygear-server/pkg/core/auth/session"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	. "github.com/skygeario/skygear-server/pkg/core/skytest"
	coreTime "github.com/skygeario/skygear-server/pkg/core/time"
	"github.com/skygeario/skygear-server/pkg/core/validation"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSignupHandler(t *testing.T) {
	Convey("Test SignupHandler", t, func() {
		two := 2
		loginIDsKeys := []config.LoginIDKeyConfiguration{
			config.LoginIDKeyConfiguration{
				Key:     "email",
				Type:    config.LoginIDKeyType(metadata.Email),
				Maximum: &two,
			},
			config.LoginIDKeyConfiguration{
				Key:     "username",
				Type:    config.LoginIDKeyTypeRaw,
				Maximum: &two,
			},
		}
		allowedRealms := []string{password.DefaultRealm, "admin"}
		authInfoStore := authinfo.NewMockStore()
		passwordAuthProvider := password.NewMockProvider(loginIDsKeys, allowedRealms)
		sh := &SignupHandler{}
		timeProvider := &coreTime.MockProvider{TimeNowUTC: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)}
		validator := validation.NewValidator("http://v2.skygear.io")
		validator.AddSchemaFragments(
			SignupRequestSchema,
		)
		sh.Validator = validator
		sh.Logger = logrus.NewEntry(logrus.New())
		mockTaskQueue := async.NewMockQueue()
		sh.TaskQueue = mockTaskQueue
		sh.TxContext = db.NewMockTxContext()
		hookProvider := hook.NewMockProvider()
		sh.HookProvider = hookProvider

		mfaStore := mfa.NewMockStore(timeProvider)
		mfaConfiguration := &config.MFAConfiguration{
			Enabled:     false,
			Enforcement: config.MFAEnforcementOptional,
		}
		mfaSender := mfa.NewMockSender()
		mfaProvider := mfa.NewProvider(mfaStore, mfaConfiguration, timeProvider, mfaSender)
		sessionProvider := session.NewMockProvider()
		sessionWriter := session.NewMockWriter()
		userProfileStore := userprofile.NewMockUserProfileStore()
		identityProvider := principal.NewMockIdentityProvider(passwordAuthProvider)
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

		Convey("should reject request without login ID", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
			{
				"login_ids": [],
				"password": "123456"
			}`))
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
								"kind": "EntryAmount",
								"pointer": "/login_ids",
								"message": "Array must have at least 1 items",
								"details": { "gte": 1 }
							}
						]
					}
				}
			}`)
		})
	})
}
