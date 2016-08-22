package pq

import (
	"testing"

	"github.com/skygeario/skygear-server/pkg/server/skydb"
	. "github.com/smartystreets/goconvey/convey"
)

func TestRecordCreationAccess(t *testing.T) {
	var c *conn

	Convey("RecordCreationAccess", t, func() {
		c = getTestConn(t)
		defer cleanupConn(t, c)

		// prepare some initial data
		c.ensureRole([]string{"Developer", "Tester", "ProjectManager"})
		c.insertRecordCreationAccess("ProgressUpdate", []string{"ProjectManager"})
		c.insertRecordCreationAccess("SourceCode", []string{"Developer"})

		Convey("get record creation access", func() {
			access, err := c.GetRecordAccess("ProgressUpdate")

			So(err, ShouldBeNil)
			So(access, ShouldHaveLength, 1)
			So(access[0].Role, ShouldEqual, "ProjectManager")
		})

		Convey("set creation access", func() {
			err := c.SetRecordAccess("SourceCode", skydb.NewRecordACL(
				[]skydb.RecordACLEntry{
					skydb.NewRecordACLEntryRole("Developer", skydb.CreateLevel),
					skydb.NewRecordACLEntryRole("Tester", skydb.CreateLevel),
				},
			))

			So(err, ShouldBeNil)

			access, err := c.GetRecordAccess("SourceCode")

			So(err, ShouldBeNil)
			So(access, ShouldHaveLength, 2)

			roles := []string{}
			for _, ace := range access {
				if ace.Role != "" {
					roles = append(roles, ace.Role)
				}
			}

			So(roles, ShouldContain, "Developer")
			So(roles, ShouldContain, "Tester")
		})

		Convey("remove not necessary creation access", func() {
			err := c.SetRecordAccess("ProgressUpdate", skydb.NewRecordACL(
				[]skydb.RecordACLEntry{
					skydb.NewRecordACLEntryRole("Developer", skydb.CreateLevel),
					skydb.NewRecordACLEntryRole("Tester", skydb.CreateLevel),
				},
			))

			So(err, ShouldBeNil)

			access, err := c.GetRecordAccess("ProgressUpdate")

			So(err, ShouldBeNil)
			So(access, ShouldHaveLength, 2)

			roles := []string{}
			for _, ace := range access {
				if ace.Role != "" {
					roles = append(roles, ace.Role)
				}
			}

			So(roles, ShouldContain, "Developer")
			So(roles, ShouldContain, "Tester")
			So(roles, ShouldNotContain, "ProjectManager")
		})
	})
}
