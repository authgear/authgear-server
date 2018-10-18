package handler

import (
	"testing"

	"github.com/skygeario/skygear-server/pkg/core/auth/role"
	"github.com/skygeario/skygear-server/pkg/core/logging"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
	. "github.com/smartystreets/goconvey/convey"
)

func TestRoleAdminHandler(t *testing.T) {
	Convey("Test RoleAdminRequestPayload", t, func() {
		Convey("validate valid payload", func() {
			payload := RoleAdminRequestPayload{
				Roles: []string{
					"role1",
					"role2",
				},
			}
			So(payload.Validate(), ShouldBeNil)
		})

		Convey("validate payload without roles", func() {
			payload := RoleAdminRequestPayload{}
			err := payload.Validate()
			errResponse := err.(skyerr.Error)
			So(errResponse.Code(), ShouldEqual, skyerr.InvalidArgument)
		})

		Convey("validate payload with empty roles", func() {
			payload := RoleAdminRequestPayload{
				Roles: []string{},
			}
			err := payload.Validate()
			errResponse := err.(skyerr.Error)
			So(errResponse.Code(), ShouldEqual, skyerr.InvalidArgument)
		})

		Convey("Test RoleAdminHandler", func() {
			roleStore := role.NewMockStoreWithRoleMap(
				map[string]role.Role{
					"admin": role.Role{
						Name: "admin",
					},
					"user": role.Role{
						Name: "user",
					},
				},
			)
			h := &RoleAdminHandler{}
			h.Logger = logging.LoggerEntry("handler")
			h.RoleStore = roleStore

			Convey("should set admin roles accordingly", func() {
				expectRoles := []string{
					"admin",
				}
				payload := RoleAdminRequestPayload{
					Roles: expectRoles,
				}

				resp, err := h.Handle(payload)
				So(err, ShouldBeNil)
				So(resp, ShouldResemble, expectRoles)
				So(roleStore.RoleMap["admin"].IsAdmin, ShouldBeTrue)
				So(roleStore.RoleMap["user"].IsAdmin, ShouldBeFalse)
			})

			Convey("admin roles should be reset", func() {
				expectRoles := []string{
					"user1",
				}
				payload := RoleAdminRequestPayload{
					Roles: expectRoles,
				}

				resp, err := h.Handle(payload)
				So(err, ShouldBeNil)
				So(resp, ShouldResemble, expectRoles)
				So(roleStore.RoleMap["admin"].IsAdmin, ShouldBeFalse)
				So(roleStore.RoleMap["user"].IsAdmin, ShouldBeFalse)
				So(roleStore.RoleMap["user1"].IsAdmin, ShouldBeTrue)
			})
		})
	})
}
