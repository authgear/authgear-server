package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/auth/task/spec"
	"github.com/skygeario/skygear-server/pkg/core/async"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/validation"
	. "github.com/smartystreets/goconvey/convey"
)

type MockResetPasswordFlow struct{}

func (m *MockResetPasswordFlow) ResetPassword(userID string, password string) error {
	return nil
}

func TestResetPasswordHandler(t *testing.T) {
	Convey("Test ResetPasswordHandler", t, func() {
		// fixture
		authInfoStore := authinfo.NewMockStoreWithAuthInfoMap(
			map[string]authinfo.AuthInfo{
				"john.doe.id": authinfo.AuthInfo{
					ID: "john.doe.id",
				},
			},
		)
		mockTaskQueue := async.NewMockQueue()

		h := &ResetPasswordHandler{}
		validator := validation.NewValidator("http://v2.skygear.io")
		validator.AddSchemaFragments(
			ResetPasswordRequestSchema,
		)
		h.TxContext = db.NewMockTxContext()
		h.Validator = validator
		h.AuthInfoStore = authInfoStore
		h.UserProfileStore = userprofile.NewMockUserProfileStore()
		hookProvider := hook.NewMockProvider()
		h.HookProvider = hookProvider
		h.TaskQueue = mockTaskQueue
		h.Interactions = &MockResetPasswordFlow{}

		Convey("should trigger hook when reset password success", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"user_id": "john.doe.id",
				"password": "234567"
			}`))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, 200)

			// should enqueue pw housekeeper task
			So(mockTaskQueue.TasksName[0], ShouldEqual, spec.PwHousekeeperTaskName)
			So(mockTaskQueue.TasksParam[0], ShouldResemble, spec.PwHousekeeperTaskParam{
				AuthID: "john.doe.id",
			})

			So(hookProvider.DispatchedEvents, ShouldResemble, []event.Payload{
				event.PasswordUpdateEvent{
					Reason: event.PasswordUpdateReasonAdministrative,
					User: model.User{
						ID:         "john.doe.id",
						VerifyInfo: map[string]bool{},
						Metadata:   userprofile.Data{},
					},
				},
			})
		})

	})
}
