package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	authAudit "github.com/skygeario/skygear-server/pkg/auth/dependency/audit"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/auth/task"
	"github.com/skygeario/skygear-server/pkg/core/async"
	"github.com/skygeario/skygear-server/pkg/core/audit"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/session"
	authtest "github.com/skygeario/skygear-server/pkg/core/auth/testing"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	. "github.com/skygeario/skygear-server/pkg/core/skytest"
)

func TestChangePasswordHandler(t *testing.T) {
	var zeroTime time.Time
	realTime := timeNow
	timeNow = func() time.Time { return zeroTime }
	defer func() {
		timeNow = realTime
	}()

	Convey("Test ChangePasswordHandler", t, func() {
		// fixture
		userID := "john.doe.id"
		mockTaskQueue := async.NewMockQueue()

		lh := &ChangePasswordHandler{}
		lh.AuditTrail = audit.NewMockTrail(t)
		lh.AuthContext = authtest.NewMockContext().UseUser(userID, "john.doe.principal.id0").MarkVerified()
		lh.AuthInfoStore = authinfo.NewMockStoreWithUser(userID)
		lh.SessionProvider = session.NewMockProvider()
		lh.SessionWriter = session.NewMockWriter()
		profileData := map[string]map[string]interface{}{
			"john.doe.id": map[string]interface{}{},
		}
		lh.UserProfileStore = userprofile.NewMockUserProfileStoreByData(profileData)
		lh.TxContext = db.NewMockTxContext()
		lh.PasswordChecker = &authAudit.PasswordChecker{
			PwMinLength: 6,
		}
		lh.PasswordAuthProvider = password.NewMockProviderWithPrincipalMap(
			map[string]config.LoginIDKeyConfiguration{},
			[]string{password.DefaultRealm},
			map[string]password.Principal{
				"john.doe.principal.id0": password.Principal{
					ID:             "john.doe.principal.id0",
					UserID:         userID,
					LoginIDKey:     "username",
					LoginID:        "john.doe",
					HashedPassword: []byte("$2a$10$/jm/S1sY6ldfL6UZljlJdOAdJojsJfkjg/pqK47Q8WmOLE19tGWQi"), // 123456
					ClaimsValue:    map[string]interface{}{},
				},
				"john.doe.principal.id1": password.Principal{
					ID:             "john.doe.principal.id1",
					UserID:         userID,
					LoginIDKey:     "email",
					LoginID:        "john.doe@example.com",
					HashedPassword: []byte("$2a$10$/jm/S1sY6ldfL6UZljlJdOAdJojsJfkjg/pqK47Q8WmOLE19tGWQi"), // 123456
					ClaimsValue: map[string]interface{}{
						"email": "john.doe@example.com",
					},
				},
			},
		)
		lh.IdentityProvider = principal.NewMockIdentityProvider(lh.PasswordAuthProvider)
		lh.TaskQueue = mockTaskQueue
		hookProvider := hook.NewMockProvider()
		lh.HookProvider = hookProvider

		Convey("change password success", func(c C) {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
			{
				"old_password": "123456",
				"password": "1234567"
			}`))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			lh.ServeHTTP(resp, req)

			So(resp.Code, ShouldEqual, 200)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": {
					"user": {
						"id": "john.doe.id",
						"created_at": "0001-01-01T00:00:00Z",
						"is_disabled": false,
						"is_verified": true,
						"metadata": {},
						"verify_info": {}
					},
					"identity": {
						"claims": {},
						"id": "john.doe.principal.id0",
						"login_id": "john.doe",
						"login_id_key": "username",
						"realm": "",
						"type": "password"
					},
					"access_token": "access-token-john.doe.id-john.doe.principal.id0-0",
					"session_id": "john.doe.id-john.doe.principal.id0"
				}
			}`)

			// should enqueue pw housekeeper task
			So(mockTaskQueue.TasksName[0], ShouldEqual, task.PwHousekeeperTaskName)
			So(mockTaskQueue.TasksParam[0], ShouldResemble, task.PwHousekeeperTaskParam{
				AuthID: userID,
			})

			So(hookProvider.DispatchedEvents, ShouldResemble, []event.Payload{
				event.PasswordUpdateEvent{
					Reason: event.PasswordUpdateReasonChangePassword,
					User: model.User{
						ID:         userID,
						Verified:   true,
						VerifyInfo: map[string]bool{},
						Metadata:   userprofile.Data{},
					},
				},
			})
		})

		Convey("change to a weak password", func(c C) {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
			{
				"old_password": "123456",
				"password": "1234"
			}`))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			lh.ServeHTTP(resp, req)

			So(resp.Body.Bytes(), ShouldEqualJSON, `
				{
					"error": {
						"name": "Invalid",
						"reason": "PasswordPolicyViolated",
						"message": "password policy violated",
						"code": 400,
						"info": {
							"causes": [
								{ "kind": "PasswordTooShort", "min_length": 6, "pw_length": 4 }
							]
						}
					}
				}
			`)
			So(resp.Code, ShouldEqual, 400)
		})

		Convey("old password incorrect", func(c C) {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
			{
				"old_password": "wrong_password",
				"password": "123456"
			}`))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			lh.ServeHTTP(resp, req)

			So(resp.Body.Bytes(), ShouldEqualJSON, `
				{
					"error": {
						"name": "Unauthorized",
						"reason": "InvalidCredentials",
						"message": "invalid credentials",
						"code": 401
					}
				}
			`)
			So(resp.Code, ShouldEqual, 401)
		})
	})
}
