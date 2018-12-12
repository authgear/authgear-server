package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/forgotpwdemail"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/provider/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"

	"github.com/skygeario/skygear-server/pkg/core/audit"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authtoken"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	. "github.com/skygeario/skygear-server/pkg/server/skytest"
	. "github.com/smartystreets/goconvey/convey"
)

func TestForgotPasswordResetHandler(t *testing.T) {
	realTime := timeNow
	timeNow = func() time.Time { return time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC) }
	defer func() {
		timeNow = realTime
	}()

	codeGenerator := &forgotpwdemail.CodeGenerator{MasterKey: "master_key"}

	Convey("Test ForgotPasswordResetHandler", t, func() {
		fh := &ForgotPasswordResetHandler{}
		logger, hook := test.NewNullLogger()
		fh.Logger = logrus.NewEntry(logger)
		fh.AuditTrail = audit.NewMockTrail(t)
		fh.TxContext = db.NewMockTxContext()
		authRecordKeys := [][]string{[]string{"email", "username"}}
		hashedPassword := []byte("$2a$10$/jm/S1sY6ldfL6UZljlJdOAdJojsJfkjg/pqK47Q8WmOLE19tGWQi") // 123456
		fh.PasswordAuthProvider = password.NewMockProviderWithPrincipalMap(
			authRecordKeys,
			map[string]password.Principal{
				"john.doe.principal.id": password.Principal{
					ID:     "john.doe.principal.id",
					UserID: "john.doe.id",
					AuthData: map[string]interface{}{
						"username": "john.doe",
						"email":    "john.doe@example.com",
					},
					HashedPassword: hashedPassword,
				},
			},
		)
		authInfo := authinfo.AuthInfo{
			ID: "john.doe.id",
		}
		fh.AuthInfoStore = authinfo.NewMockStoreWithAuthInfoMap(
			map[string]authinfo.AuthInfo{
				"john.doe.id": authInfo,
			},
		)
		userProfileStore := userprofile.NewMockUserProfileStore()
		userProfileStore.Data["john.doe.id"] = map[string]interface{}{
			"username": "john.doe",
			"email":    "john.doe@example.com",
		}
		fh.UserProfileStore = userProfileStore
		fh.TokenStore = authtoken.NewJWTStore("myApp", "secret", 0)
		fh.CodeGenerator = codeGenerator
		fh.PasswordChecker = &audit.PasswordChecker{}

		Convey("reset password after expiry", func() {
			// expireAt := time.Date(2005, 1, 2, 15, 4, 5, 0, time.UTC)                                // 1104678245
			// expectedCode := codeGenerator.Generate(authInfo, userProfile, hashedPassword, expireAt) // ed3bce0b

			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"user_id": "john.doe.id",
				"code": "ed3bce0b",
				"expire_at": 1104678245,
				"new_password": "xxx"
			}`))
			resp := httptest.NewRecorder()
			h := handler.APIHandlerToHandler(fh, fh.TxContext)
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"error": {
					"code": 110,
					"message": "user not found or code invalid",
					"name": "ResourceNotFound"
				}
			}`)
			So(hook.LastEntry().Level, ShouldEqual, logrus.ErrorLevel)
			So(hook.LastEntry().Message, ShouldEqual, "forgot password code expired")
		})

		Convey("reset password with unmatched code", func() {
			// expireAt := time.Date(2006, 2, 2, 15, 4, 5, 0, time.UTC)                                // 1138892645
			// expectedCode := codeGenerator.Generate(authInfo, userProfile, hashedPassword, expireAt) // 0e0e0776

			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"user_id": "john.doe.id",
				"code": "abcabc",
				"expire_at": 1138892645,
				"new_password": "xxx"
			}`))
			resp := httptest.NewRecorder()
			h := handler.APIHandlerToHandler(fh, fh.TxContext)
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"error": {
					"code": 110,
					"message": "user not found or code invalid",
					"name": "ResourceNotFound"
				}
			}`)
			So(hook.LastEntry().Level, ShouldEqual, logrus.ErrorLevel)
			So(hook.LastEntry().Message, ShouldEqual, "wrong forgot password reset password code")
		})

		Convey("reset password", func() {
			// expireAt := time.Date(2006, 2, 2, 15, 4, 5, 0, time.UTC)                                // 1138892645
			// expectedCode := codeGenerator.Generate(authInfo, userProfile, hashedPassword, expireAt) // 0e0e0776

			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"user_id": "john.doe.id",
				"code": "0e0e0776",
				"expire_at": 1138892645,
				"new_password": "xxx"
			}`))
			resp := httptest.NewRecorder()
			h := handler.APIHandlerToHandler(fh, fh.TxContext)
			h.ServeHTTP(resp, req)
			var respBody map[string]interface{}
			err := json.Unmarshal(resp.Body.Bytes(), &respBody)
			So(err, ShouldBeNil)
			So(respBody, ShouldNotContainKey, "error")
			So(respBody["result"], ShouldContainKey, "access_token")
			So(respBody["result"], ShouldContainKey, "user_id")
		})
	})
}
