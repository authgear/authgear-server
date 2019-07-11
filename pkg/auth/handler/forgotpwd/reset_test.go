package forgotpwd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	authAudit "github.com/skygeario/skygear-server/pkg/auth/dependency/audit"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/forgotpwdemail"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"

	"github.com/skygeario/skygear-server/pkg/auth/task"
	"github.com/skygeario/skygear-server/pkg/core/async"
	"github.com/skygeario/skygear-server/pkg/core/audit"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authtoken"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	. "github.com/skygeario/skygear-server/pkg/core/skytest"
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
		mockTaskQueue := async.NewMockQueue()

		fh := &ForgotPasswordResetHandler{}
		logger, hook := test.NewNullLogger()
		fh.Logger = logrus.NewEntry(logger)
		fh.AuditTrail = audit.NewMockTrail(t)
		fh.TxContext = db.NewMockTxContext()
		hashedPassword := []byte("$2a$10$/jm/S1sY6ldfL6UZljlJdOAdJojsJfkjg/pqK47Q8WmOLE19tGWQi") // 123456
		fh.PasswordAuthProvider = password.NewMockProviderWithPrincipalMap(
			map[string]config.LoginIDKeyConfiguration{},
			[]string{password.DefaultRealm},
			map[string]password.Principal{
				"john.doe.principal.id1": password.Principal{
					ID:             "john.doe.principal.id1",
					UserID:         "john.doe.id",
					LoginIDKey:     "username",
					LoginID:        "john.doe",
					HashedPassword: hashedPassword,
				},
				"john.doe.principal.id2": password.Principal{
					ID:             "john.doe.principal.id2",
					UserID:         "john.doe.id",
					LoginIDKey:     "email",
					LoginID:        "john.doe@example.com",
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
		fh.TokenStore = authtoken.NewJWTStore("myApp", "secret", 0)
		fh.CodeGenerator = codeGenerator
		fh.PasswordChecker = &authAudit.PasswordChecker{}
		fh.TaskQueue = mockTaskQueue

		Convey("reset password after expiry", func() {
			// expireAt := time.Date(2005, 1, 2, 15, 4, 5, 0, time.UTC)                                // 1104678245
			// expectedCode := codeGenerator.Generate(authInfo, email, hashedPassword, expireAt)       // ed3bce0b

			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"user_id": "john.doe.id",
				"code": "54edc977",
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
			// expectedCode := codeGenerator.Generate(authInfo, email, hashedPassword, expireAt)       // 0e0e0776

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
			// expectedCode := codeGenerator.Generate(authInfo, email, hashedPassword, expireAt)       // 0e0e0776

			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"user_id": "john.doe.id",
				"code": "1398d567",
				"expire_at": 1138892645,
				"new_password": "xxx"
			}`))
			resp := httptest.NewRecorder()
			h := handler.APIHandlerToHandler(fh, fh.TxContext)
			h.ServeHTTP(resp, req)
			var respBody map[string]interface{}
			err := json.Unmarshal(resp.Body.Bytes(), &respBody)
			So(err, ShouldBeNil)
			So(respBody["result"], ShouldResemble, map[string]interface{}{})

			// should enqueue pw housekeeper task
			So(mockTaskQueue.TasksName[0], ShouldEqual, task.PwHousekeeperTaskName)
			So(mockTaskQueue.TasksParam[0], ShouldResemble, task.PwHousekeeperTaskParam{
				AuthID: "john.doe.id",
			})
		})
	})
}
