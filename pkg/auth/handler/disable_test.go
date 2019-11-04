package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	coreAudit "github.com/skygeario/skygear-server/pkg/core/audit"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/db"
	. "github.com/skygeario/skygear-server/pkg/core/skytest"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

func TestSetDisableHandler(t *testing.T) {
	Convey("Test SetDisableHandler", t, func() {
		// fixture
		authInfoStore := authinfo.NewMockStoreWithAuthInfoMap(
			map[string]authinfo.AuthInfo{
				"john.doe.id": authinfo.AuthInfo{
					ID: "john.doe.id",
				},
			},
		)
		h := &SetDisableHandler{}
		validator := validation.NewValidator("http://v2.skygear.io")
		validator.AddSchemaFragments(
			SetDisableRequestSchema,
		)
		h.Validator = validator
		h.AuthInfoStore = authInfoStore
		h.UserProfileStore = userprofile.NewMockUserProfileStore()
		h.AuditTrail = coreAudit.NewMockTrail(t)
		hookProvider := hook.NewMockProvider()
		h.HookProvider = hookProvider
		h.TxContext = db.NewMockTxContext()

		Convey("reject invalid expiry time format", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
				{
					"user_id": "john.doe.id",
					"expiry": "Mon Oct 9 15:04:05 HKT 2006",
					"disabled": true,
					"message": "Temporarily disable"
				}
			`))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			h.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, 400)
			So(w.Body.Bytes(), ShouldEqualJSON, `
			{
				"error": {
					"name": "Invalid",
					"reason": "ValidationFailed",
					"message": "invalid request body",
					"code": 400,
					"info": {
						"causes": [
							{
								"kind": "StringFormat",
								"message": "Does not match format 'date-time'",
								"pointer": "/expiry",
								"details": { "format": "date-time" }
							}
						]
					}
				}
			}`)
		})

		Convey("set user disable", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
				{
					"user_id": "john.doe.id",
					"expiry": "2006-01-02T15:04:05Z",
					"disabled": true,
					"message": "Temporarily disable"
				}
			`))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			h.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, 200)

			// check the authinfo store data
			a := authinfo.AuthInfo{}
			authInfoStore.GetAuth("john.doe.id", &a)
			So(a.ID, ShouldEqual, "john.doe.id")
			So(a.Disabled, ShouldBeTrue)
			So(a.DisabledMessage, ShouldEqual, "Temporarily disable")
			So(a.DisabledExpiry.Equal(time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)), ShouldBeTrue)

			isDisabled := true
			So(hookProvider.DispatchedEvents, ShouldResemble, []event.Payload{
				event.UserUpdateEvent{
					Reason:     event.UserUpdateReasonAdministrative,
					IsDisabled: &isDisabled,
					User: model.User{
						ID:         "john.doe.id",
						Disabled:   false,
						VerifyInfo: map[string]bool{},
						Metadata:   userprofile.Data{},
					},
				},
			})
		})

		Convey("should ingore expiry and message when disable is false", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
				{
					"user_id": "john.doe.id",
					"expiry": "2006-01-02T15:04:05Z",
					"disabled": false,
					"message": "Temporarily disable"
				}
			`))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			h.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, 200)

			// check the authinfo store data
			a := authinfo.AuthInfo{}
			authInfoStore.GetAuth("john.doe.id", &a)
			So(a.ID, ShouldEqual, "john.doe.id")
			So(a.Disabled, ShouldBeFalse)
			So(a.DisabledMessage, ShouldBeEmpty)
			So(a.DisabledExpiry, ShouldBeNil)

			isDisabled := false
			So(hookProvider.DispatchedEvents, ShouldResemble, []event.Payload{
				event.UserUpdateEvent{
					Reason:     event.UserUpdateReasonAdministrative,
					IsDisabled: &isDisabled,
					User: model.User{
						ID:         "john.doe.id",
						Disabled:   false,
						VerifyInfo: map[string]bool{},
						Metadata:   userprofile.Data{},
					},
				},
			})
		})
	})
}
