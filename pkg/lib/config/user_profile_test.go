package config

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestStandardAttributesConfig(t *testing.T) {
	Convey("StandardAttributesConfig", t, func() {
		c := &StandardAttributesConfig{}
		c.SetDefaults()

		accessControl := c.GetAccessControl()

		So(accessControl.GetLevel("/name", RoleEndUser, 0), ShouldEqual, AccessControlLevelHidden)
		So(accessControl.GetLevel("/family_name", RoleEndUser, 0), ShouldEqual, AccessControlLevelReadwrite)

		So(accessControl.GetLevel("/name", RoleBearer, 0), ShouldEqual, AccessControlLevelHidden)
		So(accessControl.GetLevel("/family_name", RoleBearer, 0), ShouldEqual, AccessControlLevelReadwrite)

		So(accessControl.GetLevel("/name", RoleBearer, 0), ShouldEqual, AccessControlLevelHidden)
		So(accessControl.GetLevel("/family_name", RoleBearer, 0), ShouldEqual, AccessControlLevelReadwrite)
	})
}
