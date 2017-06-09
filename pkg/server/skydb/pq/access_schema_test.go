package pq

import (
	"sort"
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

func TestRecordFieldAccess(t *testing.T) {
	var c *conn

	Convey("RecordFieldAccess", t, func() {
		c = getTestConn(t)
		defer cleanupConn(t, c)

		tableName := c.tableName("_record_field_access")
		tableColumns := []string{
			"record_type",
			"record_field",
			"user_role",
			"writable",
			"readable",
			"comparable",
			"discoverable",
		}

		insertEntries := func(list skydb.FieldACLEntryList) error {
			builder := psql.
				Insert(tableName).
				Columns(tableColumns...)
			for _, entry := range list {
				builder = builder.Values(
					entry.RecordType,
					entry.RecordField,
					entry.UserRole,
					entry.Writable,
					entry.Readable,
					entry.Comparable,
					entry.Discoverable,
				)
			}
			_, err := c.ExecWith(builder)
			return err
		}

		Convey("should return with default entry for empty table", func() {
			acl, err := c.GetRecordFieldAccess()
			So(err, ShouldBeNil)
			entry := acl.FindDefaultEntry()
			So(entry, ShouldNotBeNil)
			So(entry.Writable, ShouldBeTrue)
			So(entry.Readable, ShouldBeTrue)
			So(entry.Comparable, ShouldBeTrue)
			So(entry.Discoverable, ShouldBeTrue)
		})

		Convey("should return with default entry if not exists in table", func() {
			So(insertEntries(skydb.FieldACLEntryList{
				{"note", "*", "_public", false, false, false, false},
			}), ShouldBeNil)
			acl, err := c.GetRecordFieldAccess()
			So(err, ShouldBeNil)
			entry := acl.FindDefaultEntry()
			So(entry, ShouldNotBeNil)
			So(entry.Writable, ShouldBeTrue)
			So(entry.Readable, ShouldBeTrue)
			So(entry.Comparable, ShouldBeTrue)
			So(entry.Discoverable, ShouldBeTrue)
		})

		Convey("should return all entries", func() {
			fixture := skydb.FieldACLEntryList{
				{"*", "*", "_public", false, false, false, false},
				{"*", "content", "_any_user", false, false, true, true},
				{"note", "*", "_public", true, true, false, false},
			}
			So(insertEntries(fixture), ShouldBeNil)
			acl, err := c.GetRecordFieldAccess()
			So(err, ShouldBeNil)

			defaultEntry := acl.FindDefaultEntry()
			So(defaultEntry, ShouldNotBeNil)
			So(defaultEntry.UserRole, ShouldEqual, "_public")
			So(defaultEntry.Writable, ShouldBeFalse)
			So(defaultEntry.Readable, ShouldBeFalse)
			So(defaultEntry.Comparable, ShouldBeFalse)
			So(defaultEntry.Discoverable, ShouldBeFalse)

			entries := acl.AllEntries()
			sort.Stable(entries)
			So(entries, ShouldResemble, fixture)
		})

		Convey("should insert all entries", func() {
			fixture := skydb.FieldACLEntryList{
				{"*", "*", "_public", false, false, false, false},
				{"*", "content", "_any_user", false, false, true, true},
				{"note", "*", "_public", true, true, false, false},
			}
			acl := skydb.NewFieldACL(fixture)
			err := c.SetRecordFieldAccess(acl)
			So(err, ShouldBeNil)
			acl, err = c.GetRecordFieldAccess()
			So(acl, ShouldNotBeNil)
			So(err, ShouldBeNil)
			So(acl.AllEntries(), ShouldResemble, fixture)
		})

		Convey("should remove entry before insert", func() {
			So(insertEntries(skydb.FieldACLEntryList{
				{"*", "*", "_public", false, false, false, false},
				{"photo", "*", "_public", true, true, false, false},
			}), ShouldBeNil)
			fixture := skydb.FieldACLEntryList{
				{"*", "*", "_public", true, false, false, true},
				{"*", "content", "_any_user", false, false, true, true},
				{"note", "*", "_public", true, true, false, false},
			}
			acl := skydb.NewFieldACL(fixture)
			err := c.SetRecordFieldAccess(acl)
			So(err, ShouldBeNil)
			acl, err = c.GetRecordFieldAccess()
			So(acl, ShouldNotBeNil)
			So(err, ShouldBeNil)
			So(acl.AllEntries(), ShouldResemble, fixture)
		})
	})
}
