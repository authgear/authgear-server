package customattrs

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
)

func TestT(t *testing.T) {
	Convey("T", t, func() {
		Convey("Clone", func() {
			a := T{
				"a": "a",
			}
			cloned := a.Clone()
			cloned["a"] = "b"
			So(a["a"], ShouldEqual, "a")
			So(cloned["a"], ShouldEqual, "b")
		})

		Convey("Update", func() {
			// Can update when level is enough.
			actual, err := T{
				"a": "a",
				"b": "b",
			}.Update(accesscontrol.T{
				accesscontrol.Subject("/a"): map[accesscontrol.Role]accesscontrol.Level{
					config.RoleEndUser: config.AccessControlLevelReadwrite,
				},
			}, config.RoleEndUser, []string{"/a"}, T{
				"a": "aa",
			})
			So(err, ShouldBeNil)
			So(actual, ShouldResemble, T{
				"a": "aa",
				"b": "b",
			})

			// Can delete when level is enough
			actual, err = T{
				"a": "a",
				"b": "b",
			}.Update(accesscontrol.T{
				accesscontrol.Subject("/a"): map[accesscontrol.Role]accesscontrol.Level{
					config.RoleEndUser: config.AccessControlLevelReadwrite,
				},
			}, config.RoleEndUser, []string{"/a"}, T{})
			So(err, ShouldBeNil)
			So(actual, ShouldResemble, T{
				"b": "b",
			})

			// Only consider given pointers.
			actual, err = T{
				"a": "a",
				"b": "b",
			}.Update(accesscontrol.T{
				accesscontrol.Subject("/a"): map[accesscontrol.Role]accesscontrol.Level{
					config.RoleEndUser: config.AccessControlLevelReadwrite,
				},
			}, config.RoleEndUser, []string{"/a"}, T{
				"a": "aa",
				"b": "bb",
			})
			So(err, ShouldBeNil)
			So(actual, ShouldResemble, T{
				"a": "aa",
				"b": "b",
			})

			// Cannot update when level is not enough
			_, err = T{
				"a": "a",
				"b": "b",
			}.Update(accesscontrol.T{
				accesscontrol.Subject("/a"): map[accesscontrol.Role]accesscontrol.Level{
					config.RoleEndUser: config.AccessControlLevelReadonly,
				},
			}, config.RoleEndUser, []string{"/a"}, T{
				"a": "aa",
			})
			So(err, ShouldBeError, "/a being updated by end_user with level 2")

			// Cannot delete when level is not enough
			_, err = T{
				"a": "a",
				"b": "b",
			}.Update(accesscontrol.T{
				accesscontrol.Subject("/a"): map[accesscontrol.Role]accesscontrol.Level{
					config.RoleEndUser: config.AccessControlLevelReadonly,
				},
			}, config.RoleEndUser, []string{"/a"}, T{})
			So(err, ShouldBeError, "/a being deleted by end_user with level 2")
		})

		Convey("ReadWithAccessControl", func() {
			accessControl := accesscontrol.T{
				accesscontrol.Subject("/a"): map[accesscontrol.Role]accesscontrol.Level{
					config.RoleEndUser:  config.AccessControlLevelHidden,
					config.RoleBearer:   config.AccessControlLevelReadwrite,
					config.RolePortalUI: config.AccessControlLevelReadwrite,
				},
				accesscontrol.Subject("/b"): map[accesscontrol.Role]accesscontrol.Level{
					config.RoleEndUser:  config.AccessControlLevelReadwrite,
					config.RoleBearer:   config.AccessControlLevelReadwrite,
					config.RolePortalUI: config.AccessControlLevelReadwrite,
				},
			}
			stdAttrs := T{
				"a": "a",
				"b": "b",
			}

			So(stdAttrs.ReadWithAccessControl(accessControl, config.RoleEndUser), ShouldResemble, T{
				"b": "b",
			})

			So(stdAttrs.ReadWithAccessControl(accessControl, config.RolePortalUI), ShouldResemble, T{
				"a": "a",
				"b": "b",
			})
		})
	})
}
