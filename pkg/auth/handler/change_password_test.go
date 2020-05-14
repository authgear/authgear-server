package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	authtesting "github.com/skygeario/skygear-server/pkg/auth/dependency/auth/testing"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/auth/task/spec"
	"github.com/skygeario/skygear-server/pkg/core/async"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/db"
	. "github.com/skygeario/skygear-server/pkg/core/skytest"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

type MockPasswordFlow struct {
}

func (m *MockPasswordFlow) ChangePassword(userID string, OldPassword string, newPassword string) error {
	return nil
}

func TestChangePasswordHandler(t *testing.T) {
	Convey("Test ChangePasswordHandler", t, func() {
		// fixture
		userID := "john.doe.id"
		mockTaskQueue := async.NewMockQueue()

		lh := &ChangePasswordHandler{}
		validator := validation.NewValidator("http://v2.skygear.io")
		validator.AddSchemaFragments(
			ChangePasswordRequestSchema,
		)
		lh.Validator = validator
		lh.AuthInfoStore = authinfo.NewMockStoreWithAuthInfoMap(map[string]authinfo.AuthInfo{
			userID: {
				ID:       userID,
				Verified: true,
			},
		})
		profileData := map[string]map[string]interface{}{
			"john.doe.id": map[string]interface{}{},
		}
		lh.UserProfileStore = userprofile.NewMockUserProfileStoreByData(profileData)
		lh.TxContext = db.NewMockTxContext()
		lh.TaskQueue = mockTaskQueue
		hookProvider := hook.NewMockProvider()
		lh.HookProvider = hookProvider
		lh.Interactions = &MockPasswordFlow{}

		Convey("should trigger hook when update password success", func(c C) {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
			{
				"old_password": "123456",
				"password": "1234567"
			}`))
			req = authtesting.WithAuthn().
				UserID(userID).
				Verified(true).
				IdentityType("login_id").
				ToRequest(req)
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
						"is_manually_verified": false,
						"is_verified": true,
						"metadata": {},
						"verify_info": {}
					},
					"identity": {
						"claims": {},
						"type": "login_id"
					}
				}
			}`)

			// should enqueue pw housekeeper task
			So(mockTaskQueue.TasksName[0], ShouldEqual, spec.PwHousekeeperTaskName)
			So(mockTaskQueue.TasksParam[0], ShouldResemble, spec.PwHousekeeperTaskParam{
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
	})
}
