package pq

import (
	"testing"

	"github.com/oursky/skygear/skydb"
	. "github.com/smartystreets/goconvey/convey"
)

func TestPredicateSqlizerFactory(t *testing.T) {
	Convey("Functional Predicate", t, func() {
		Convey("user discover must be used with user record", func() {
			f := newPredicateSqlizerFactory(nil, "note")
			userDiscover := skydb.UserDiscoverFunc{
				Emails: []string{},
			}
			_, err := f.newUserDiscoverFunctionalPredicateSqlizer(userDiscover)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestAccessPredicateSqlizer(t *testing.T) {
	Convey("access Predicate", t, func() {
		Convey("serialized for direct ACE", func() {
			userinfo := skydb.UserInfo{
				ID: "userid",
			}
			sqlizer := &accessPredicateSqlizer{
				userinfo,
				"read",
			}
			sql, args, err := sqlizer.ToSql()
			So(err, ShouldBeNil)
			So(sql, ShouldEqual, `(_access @> '[{"user_id":"userid"}]' OR`+
				` _access IS NULL OR _owner_id = ?)`)
			So(args, ShouldResemble, []interface{}{"userid"})
		})

		Convey("serialized for role based ACE", func() {
			userinfo := skydb.UserInfo{
				ID:    "userid",
				Roles: []string{"admin", "writer"},
			}
			sqlizer := &accessPredicateSqlizer{
				userinfo,
				"read",
			}
			sql, args, err := sqlizer.ToSql()
			So(err, ShouldBeNil)
			So(sql, ShouldEqual, `(_access @> '[{"role":"admin"}]' OR `+
				`_access @> '[{"role":"writer"}]' OR `+
				`_access @> '[{"user_id":"userid"}]' OR `+
				`_access IS NULL OR _owner_id = ?)`)
			So(args, ShouldResemble, []interface{}{"userid"})
		})

	})
}
