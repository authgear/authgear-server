package customattrs

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
)

func TestT(t *testing.T) {
	Convey("T", t, func() {
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
