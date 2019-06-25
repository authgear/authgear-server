package handler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	authAudit "github.com/skygeario/skygear-server/pkg/auth/dependency/audit"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/provider/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/task"
	"github.com/skygeario/skygear-server/pkg/core/async"
	"github.com/skygeario/skygear-server/pkg/core/audit"
	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authtoken"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
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
		mockTokenStore := authtoken.NewMockStore()
		mockTaskQueue := async.NewMockQueue()

		lh := &ChangePasswordHandler{}
		lh.AuditTrail = audit.NewMockTrail(t)
		lh.AuthContext = auth.NewMockContextGetterWithUser(userID, true, map[string]bool{})
		lh.AuthInfoStore = authinfo.NewMockStoreWithUser(userID)
		lh.TokenStore = mockTokenStore
		profileData := map[string]map[string]interface{}{
			"john.doe.id": map[string]interface{}{},
		}
		lh.UserProfileStore = userprofile.NewMockUserProfileStoreByData(profileData)
		lh.TxContext = db.NewMockTxContext()
		lh.PasswordChecker = &authAudit.PasswordChecker{
			PwMinLength: 6,
		}
		lh.PasswordAuthProvider = password.NewMockProviderWithPrincipalMap(
			[]string{},
			map[string]password.Principal{
				"john.doe.principal.id0": password.Principal{
					ID:             "john.doe.principal.id0",
					UserID:         userID,
					LoginIDKey:     "username",
					LoginID:        "john.doe",
					HashedPassword: []byte("$2a$10$/jm/S1sY6ldfL6UZljlJdOAdJojsJfkjg/pqK47Q8WmOLE19tGWQi"), // 123456
				},
				"john.doe.principal.id1": password.Principal{
					ID:             "john.doe.principal.id1",
					UserID:         userID,
					LoginIDKey:     "email",
					LoginID:        "john.doe@example.com",
					HashedPassword: []byte("$2a$10$/jm/S1sY6ldfL6UZljlJdOAdJojsJfkjg/pqK47Q8WmOLE19tGWQi"), // 123456
				},
			},
		)
		lh.TaskQueue = mockTaskQueue
		h := handler.APIHandlerToHandler(lh, lh.TxContext)

		Convey("change password success", func(c C) {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
			{
				"old_password": "123456",
				"password": "1234567"
			}`))
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)

			token := mockTokenStore.GetTokensByAuthInfoID(userID)[0]

			So(resp.Code, ShouldEqual, 200)
			So(resp.Body.Bytes(), ShouldEqualJSON, fmt.Sprintf(`{
				"result": {
					"user_id": "%s",
					"access_token": "%s",
					"verified":true,
					"verify_info":{},
					"created_at": "0001-01-01T00:00:00Z",
					"created_by": "%s",
					"updated_at": "0001-01-01T00:00:00Z",
					"updated_by": "%s",
					"metadata": {}
				}
			}`,
				userID,
				token.AccessToken,
				userID,
				userID))

			// should enqueue pw housekeeper task
			So(mockTaskQueue.TasksName[0], ShouldEqual, task.PwHousekeeperTaskName)
			So(mockTaskQueue.TasksParam[0], ShouldResemble, task.PwHousekeeperTaskParam{
				AuthID: userID,
			})
		})

		Convey("change to a weak password", func(c C) {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
			{
				"old_password": "123456",
				"password": "1234"
			}`))
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)

			So(resp.Body.Bytes(), ShouldEqualJSON, `
				{
					"error": {
						"code": 126,
						"name": "PasswordPolicyViolated",
						"message": "password too short",
						"info": {
							"reason": "PasswordTooShort",
							"min_length": 6,
							"pw_length": 4
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
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)

			So(resp.Body.Bytes(), ShouldEqualJSON, `
				{
					"error": {
						"code": 105,
						"message": "Incorrect old password",
						"name": "InvalidCredentials"
					}
				}
			`)
			So(resp.Code, ShouldEqual, 401)
		})
	})
}
