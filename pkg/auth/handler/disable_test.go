package handler

import (
	"net/http"
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
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

func TestSetDisableHandler(t *testing.T) {
	Convey("Test setDisableUserPayload", t, func() {
		Convey("validate valid payload", func() {
			payload := setDisableUserPayload{
				UserID:       "john.doe.id",
				Disabled:     true,
				ExpiryString: "2006-01-02T15:04:05Z",
			}
			So(payload.Validate(), ShouldBeNil)
		})

		Convey("validate payload without UserID", func() {
			payload := setDisableUserPayload{}
			err := payload.Validate()
			errResponse := err.(skyerr.Error)
			So(errResponse.Code(), ShouldEqual, skyerr.InvalidArgument)
		})
	})

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
		h.AuthInfoStore = authInfoStore
		h.UserProfileStore = userprofile.NewMockUserProfileStore()
		h.AuditTrail = coreAudit.NewMockTrail(t)
		hookProvider := hook.NewMockProvider()
		h.HookProvider = hookProvider

		Convey("decode valid request", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
				{
					"user_id": "john.doe.id",
					"expiry": "2006-01-02T15:04:05Z",
					"disabled": true,
					"message": "Temporarily disable"
				}
			`))
			req.Header.Set("Content-Type", "application/json")
			payload, err := h.DecodeRequest(req)
			disablePayload, ok := payload.(setDisableUserPayload)
			So(ok, ShouldBeTrue)
			So(disablePayload.UserID, ShouldEqual, "john.doe.id")
			So(disablePayload.expiry.Equal(time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)), ShouldBeTrue)
			So(err, ShouldBeNil)
		})

		Convey("decode invalid expiry time format", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
				{
					"user_id": "john.doe.id",
					"expiry": "Mon Oct 9 15:04:05 HKT 2006",
					"disabled": true,
					"message": "Temporarily disable"
				}
			`))
			req.Header.Set("Content-Type", "application/json")
			_, err := h.DecodeRequest(req)
			errResponse := err.(skyerr.Error)
			So(errResponse.Code(), ShouldEqual, skyerr.InvalidArgument)
		})

		Convey("set user disable", func() {
			expiry := time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)
			userID := "john.doe.id"
			payload := setDisableUserPayload{
				UserID:   userID,
				Disabled: true,
				Message:  "Temporarily disable",
				expiry:   &expiry,
			}

			resp, err := h.Handle(payload)
			So(resp, ShouldResemble, map[string]string{})
			So(err, ShouldBeNil)

			// check the authinfo store data
			a := authinfo.AuthInfo{}
			authInfoStore.GetAuth(userID, &a)
			So(a.ID, ShouldEqual, userID)
			So(a.Disabled, ShouldBeTrue)
			So(a.DisabledMessage, ShouldEqual, "Temporarily disable")
			So(a.DisabledExpiry.Equal(expiry), ShouldBeTrue)

			isDisabled := true
			So(hookProvider.DispatchedEvents, ShouldResemble, []event.Payload{
				event.UserUpdateEvent{
					Reason:     event.UserUpdateReasonAdministrative,
					IsDisabled: &isDisabled,
					User: model.User{
						ID:         userID,
						Disabled:   false,
						VerifyInfo: map[string]bool{},
						Metadata:   userprofile.Data{},
					},
				},
			})
		})

		Convey("should ingore expiry and message when disable is false", func() {
			expiry := time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)
			userID := "john.doe.id"
			payload := setDisableUserPayload{
				UserID:   userID,
				Disabled: false,
				Message:  "Temporarily disable",
				expiry:   &expiry,
			}

			resp, err := h.Handle(payload)
			So(resp, ShouldResemble, map[string]string{})
			So(err, ShouldBeNil)

			// check the authinfo store data
			a := authinfo.AuthInfo{}
			authInfoStore.GetAuth(userID, &a)
			So(a.ID, ShouldEqual, userID)
			So(a.Disabled, ShouldBeFalse)
			So(a.DisabledMessage, ShouldBeEmpty)
			So(a.DisabledExpiry, ShouldBeNil)

			isDisabled := false
			So(hookProvider.DispatchedEvents, ShouldResemble, []event.Payload{
				event.UserUpdateEvent{
					Reason:     event.UserUpdateReasonAdministrative,
					IsDisabled: &isDisabled,
					User: model.User{
						ID:         userID,
						Disabled:   false,
						VerifyInfo: map[string]bool{},
						Metadata:   userprofile.Data{},
					},
				},
			})
		})

		Convey("log audit trail when disable user", func() {
			expiry := time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)
			userID := "john.doe.id"
			payload := setDisableUserPayload{
				UserID:   userID,
				Disabled: true,
				Message:  "Temporarily disable",
				expiry:   &expiry,
			}
			h.Handle(payload)
			mockTrail, _ := h.AuditTrail.(*coreAudit.MockTrail)
			So(mockTrail.Hook.LastEntry().Message, ShouldEqual, "audit_trail")
			So(mockTrail.Hook.LastEntry().Data["event"], ShouldEqual, "disable_user")
		})
	})
}
