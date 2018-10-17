package handler

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

func TestRoleRevokePayload(t *testing.T) {
	Convey("RoleRevokePayload", t, func() {
		Convey("should validate payload", func() {
			payload := RoleRevokeRequestPayload{
				Roles:   []string{"admin", "developer"},
				UserIDs: []string{"john", "jane"},
			}
			So(payload.Validate(), ShouldBeNil)
		})

		Convey("should reject payload without users", func() {
			payload := RoleRevokeRequestPayload{
				Roles:   []string{"admin", "developer"},
				UserIDs: []string{},
			}
			err := payload.Validate()
			errResponse := err.(skyerr.Error)
			So(errResponse.Code(), ShouldEqual, skyerr.InvalidArgument)
		})

		Convey("validate payload without roles", func() {
			payload := RoleRevokeRequestPayload{
				Roles:   []string{},
				UserIDs: []string{"john", "jane"},
			}
			err := payload.Validate()
			errResponse := err.(skyerr.Error)
			So(errResponse.Code(), ShouldEqual, skyerr.InvalidArgument)
		})
	})
}

func TestRoleRevokeHandler(t *testing.T) {
	Convey("RoleRevokeHandler", t, func() {
		realTime := timeNow
		timeNow = func() time.Time { return time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC) }
		defer func() {
			timeNow = realTime
		}()

		// fixture
		authInfoStore := authinfo.NewMockStoreWithAuthInfoMap(
			map[string]authinfo.AuthInfo{
				"john.doe.id": authinfo.AuthInfo{
					ID: "john.doe.id",
					Roles: []string{
						"admin",
						"developer",
						"human",
					},
				},
				"jane.doe.id": authinfo.AuthInfo{
					ID: "jane.doe.id",
					Roles: []string{
						"human",
					},
				},
			},
		)

		h := &RoleRevokeHandler{}
		h.AuthInfoStore = authInfoStore

		Convey("should revoke existing user and roles", func() {
			payload := RoleRevokeRequestPayload{
				Roles: []string{
					"admin",
					"developer",
				},
				UserIDs: []string{
					"john.doe.id",
				},
			}

			resp, err := h.Handle(payload)
			So(resp, ShouldEqual, "OK")
			So(err, ShouldBeNil)

			// Asset the authinfo store data
			a := authinfo.AuthInfo{}
			authInfoStore.GetAuth("john.doe.id", &a)
			So(a.Roles, ShouldResemble, []string{"human"})
		})
	})
}
