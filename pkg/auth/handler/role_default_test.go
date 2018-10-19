package handler

import (
	"testing"

	"github.com/skygeario/skygear-server/pkg/core/auth/role"
	"github.com/skygeario/skygear-server/pkg/core/logging"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
	. "github.com/smartystreets/goconvey/convey"
)

func TestRoleDefaultRequestPayload(t *testing.T) {
	Convey("RoleDefaultRequestPayload", t, func() {
		Convey("should validate valid payload", func() {
			payload := RoleDefaultRequestPayload{
				Roles: []string{
					"role1",
					"role2",
				},
			}
			So(payload.Validate(), ShouldBeNil)
		})

		Convey("should reject payload without roles", func() {
			payload := RoleDefaultRequestPayload{}
			err := payload.Validate()
			errResponse := err.(skyerr.Error)
			So(errResponse.Code(), ShouldEqual, skyerr.InvalidArgument)
		})

		Convey("should reject payload with empty roles", func() {
			payload := RoleDefaultRequestPayload{
				Roles: []string{},
			}
			err := payload.Validate()
			errResponse := err.(skyerr.Error)
			So(errResponse.Code(), ShouldEqual, skyerr.InvalidArgument)
		})
	})
}

func TestRoleDefaultHandler(t *testing.T) {
	Convey("RoleDefaultHandler", t, func() {
		Convey("Test RoleDefaultHandler", func() {
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
			h := &RoleDefaultHandler{}
			h.Logger = logging.LoggerEntry("handler")
			h.RoleStore = roleStore

			Convey("should set default roles accordingly", func() {
				expectRoles := []string{
					"admin",
				}
				payload := RoleDefaultRequestPayload{
					Roles: expectRoles,
				}

				resp, err := h.Handle(payload)
				So(err, ShouldBeNil)
				So(resp, ShouldResemble, expectRoles)
				So(roleStore.RoleMap["admin"].IsDefault, ShouldBeTrue)
				So(roleStore.RoleMap["user"].IsDefault, ShouldBeFalse)
			})

			Convey("should reset default roles", func() {
				expectRoles := []string{
					"human",
				}
				payload := RoleDefaultRequestPayload{
					Roles: expectRoles,
				}

				resp, err := h.Handle(payload)
				So(err, ShouldBeNil)
				So(resp, ShouldResemble, expectRoles)
				So(roleStore.RoleMap["admin"].IsDefault, ShouldBeFalse)
				So(roleStore.RoleMap["user"].IsDefault, ShouldBeFalse)
				So(roleStore.RoleMap["human"].IsDefault, ShouldBeTrue)
			})
		})
	})
}
