// Copyright 2015-present Oursky Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pq

import (
	"fmt"
	"testing"

	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSqlizer(t *testing.T) {
	Convey("Expression Sqlizer", t, func() {
		Convey("literal null expression", func() {
			expr := &expressionSqlizer{"table", skydb.Expression{skydb.Literal, nil}}
			sql, args, err := expr.ToSql()
			So(sql, ShouldEqual, "NULL")
			So(args, ShouldResemble, []interface{}{})
			So(err, ShouldBeNil)
		})
	})
}

func TestPredicateSqlizerFactory(t *testing.T) {
	Convey("Expression", t, func() {
		c := getTestConn(t)
		defer cleanupConn(t, c)

		db := c.PublicDB().(*database)
		_, err := db.Extend("note", skydb.RecordSchema{
			"title": skydb.FieldType{Type: skydb.TypeString},
		})
		So(err, ShouldBeNil)

		f := newPredicateSqlizerFactory(db, "note")

		Convey("existing keypath", func() {
			sqlizer, err := f.newExpressionSqlizer(
				skydb.Expression{skydb.KeyPath, "title"},
			)
			So(err, ShouldBeNil)
			sql, args, err := sqlizer.ToSql()
			So(sql, ShouldEqual, `"note"."title"`)
			So(args, ShouldResemble, []interface{}{})
			So(err, ShouldBeNil)
		})

		Convey("non-existing keypath", func() {
			_, err := f.newExpressionSqlizer(
				skydb.Expression{skydb.KeyPath, "wrong_title"},
			)
			builderError, ok := err.(skyerr.Error)
			So(ok, ShouldBeTrue)
			So(builderError.Code(), ShouldEqual, skyerr.RecordQueryInvalid)
		})
	})

	Convey("Comparison Predicate", t, func() {
		c := getTestConn(t)
		defer cleanupConn(t, c)

		db := c.PublicDB().(*database)
		_, err := db.Extend("note", skydb.RecordSchema{
			"title":   skydb.FieldType{Type: skydb.TypeString},
			"content": skydb.FieldType{Type: skydb.TypeString},
		})
		So(err, ShouldBeNil)

		f := newPredicateSqlizerFactory(db, "note")

		Convey("keypath equal null", func() {
			sqlizer, err := f.newComparisonPredicateSqlizer(skydb.Predicate{
				skydb.Equal,
				[]interface{}{
					skydb.Expression{skydb.KeyPath, "content"},
					skydb.Expression{skydb.Literal, nil},
				},
			})
			So(err, ShouldBeNil)
			sql, args, err := sqlizer.ToSql()
			So(sql, ShouldEqual, "\"note\".\"content\" IS NULL")
			So(args, ShouldResemble, []interface{}{})
			So(err, ShouldBeNil)
		})

		Convey("null equal keypath", func() {
			sqlizer, err := f.newComparisonPredicateSqlizer(skydb.Predicate{
				skydb.Equal,
				[]interface{}{
					skydb.Expression{skydb.Literal, nil},
					skydb.Expression{skydb.KeyPath, "content"},
				},
			})
			So(err, ShouldBeNil)
			sql, args, err := sqlizer.ToSql()
			So(sql, ShouldEqual, "\"note\".\"content\" IS NULL")
			So(args, ShouldResemble, []interface{}{})
			So(err, ShouldBeNil)
		})

		Convey("keypath not equal null", func() {
			sqlizer, err := f.newComparisonPredicateSqlizer(skydb.Predicate{
				skydb.NotEqual,
				[]interface{}{
					skydb.Expression{skydb.KeyPath, "content"},
					skydb.Expression{skydb.Literal, nil},
				},
			})
			So(err, ShouldBeNil)
			sql, args, err := sqlizer.ToSql()
			So(sql, ShouldEqual, "\"note\".\"content\" IS NOT NULL")
			So(args, ShouldResemble, []interface{}{})
			So(err, ShouldBeNil)
		})

		Convey("keypath is in array of values", func() {
			sqlizer, err := f.newComparisonPredicateSqlizer(skydb.Predicate{
				skydb.In,
				[]interface{}{
					skydb.Expression{skydb.KeyPath, "content"},
					skydb.Expression{skydb.Literal, []string{"hello", "world"}},
				},
			})
			So(err, ShouldBeNil)
			sql, args, err := sqlizer.ToSql()
			So(sql, ShouldEqual, "\"note\".\"content\" IN ?")
			So(args, ShouldResemble, []interface{}{[]string{"hello", "world"}})
			So(err, ShouldBeNil)
		})

		Convey("non-existent keypath for equality", func() {
			_, err := f.newComparisonPredicateSqlizer(skydb.Predicate{
				skydb.Equal,
				[]interface{}{
					skydb.Expression{skydb.KeyPath, "wrong_title"},
					skydb.Expression{skydb.Literal, "hello world"},
				},
			})
			builderError, ok := err.(skyerr.Error)
			So(ok, ShouldBeTrue)
			So(builderError.Code(), ShouldEqual, skyerr.RecordQueryInvalid)
		})

		Convey("non-existent keypath is in array of values", func() {
			_, err := f.newComparisonPredicateSqlizer(skydb.Predicate{
				skydb.In,
				[]interface{}{
					skydb.Expression{skydb.KeyPath, "wrong_title"},
					skydb.Expression{skydb.Literal, []string{"hello", "world"}},
				},
			})
			builderError, ok := err.(skyerr.Error)
			So(ok, ShouldBeTrue)
			So(builderError.Code(), ShouldEqual, skyerr.RecordQueryInvalid)
		})
	})

	Convey("Functional Predicate", t, func() {
		c := getTestConn(t)
		defer cleanupConn(t, c)

		db := c.PublicDB().(*database)
		_, err := db.Extend("note", skydb.RecordSchema{
			"title": skydb.FieldType{Type: skydb.TypeString},
		})
		So(err, ShouldBeNil)

		f := newPredicateSqlizerFactory(db, "user")

		Convey("user discover must be used with user record", func() {
			noteSqlizerFactory := newPredicateSqlizerFactory(db, "note")
			userDiscover := skydb.UserDiscoverFunc{
				Emails: []string{},
			}
			_, err := noteSqlizerFactory.newUserDiscoverFunctionalPredicateSqlizer(userDiscover)
			So(err, ShouldNotBeNil)
		})

		Convey("should generate false for empty user discover args", func() {
			userDiscover := skydb.UserDiscoverFunc{}
			sqlizer, err := f.newUserDiscoverFunctionalPredicateSqlizer(userDiscover)
			So(err, ShouldBeNil)
			sql, args, err := sqlizer.ToSql()
			So(sql, ShouldEqual, "FALSE")
			So(args, ShouldResemble, []interface{}{})
			So(err, ShouldBeNil)
		})

		Convey("should generate sql for email args", func() {
			userDiscover := skydb.UserDiscoverFunc{
				Emails: []string{"jane.doe@example.com"},
			}
			sqlizer, err := f.newUserDiscoverFunctionalPredicateSqlizer(userDiscover)
			So(err, ShouldBeNil)
			alias := f.createLeftJoin("_user", "_id", "id")
			sql, args, err := sqlizer.ToSql()
			So(sql, ShouldEqual, fmt.Sprintf(`("%s"."email" IN (?))`, alias))
			So(args, ShouldResemble, []interface{}{"jane.doe@example.com"})
			So(err, ShouldBeNil)
		})

		Convey("should generate sql for both email and username args", func() {
			userDiscover := skydb.UserDiscoverFunc{
				Usernames: []string{"jane.doe"},
				Emails:    []string{"jane.doe@example.com"},
			}
			sqlizer, err := f.newUserDiscoverFunctionalPredicateSqlizer(userDiscover)
			So(err, ShouldBeNil)
			alias := f.createLeftJoin("_user", "_id", "id")
			sql, args, err := sqlizer.ToSql()
			So(sql, ShouldEqual, fmt.Sprintf(`("%s"."username" IN (?) OR "%s"."email" IN (?))`, alias, alias))
			So(args, ShouldResemble, []interface{}{"jane.doe", "jane.doe@example.com"})
			So(err, ShouldBeNil)
		})
	})
}

func TestNotSqlizer(t *testing.T) {
	Convey("NotSqlizer", t, func() {
		Convey("should generate not predicate", func() {
			sqlizer := &NotSqlizer{
				&expressionSqlizer{"table", skydb.Expression{skydb.Literal, false}},
			}
			sql, args, err := sqlizer.ToSql()
			So(sql, ShouldEqual, "NOT (?)")
			So(args, ShouldResemble, []interface{}{false})
			So(err, ShouldBeNil)
		})
	})
}

func TestFalseSqlizer(t *testing.T) {
	Convey("FalseSqlizer", t, func() {
		Convey("should generate predicate that evaluates to false", func() {
			sqlizer := &FalseSqlizer{}
			sql, args, err := sqlizer.ToSql()
			So(sql, ShouldEqual, "FALSE")
			So(args, ShouldResemble, []interface{}{})
			So(err, ShouldBeNil)
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
				&userinfo,
				skydb.ReadLevel,
			}
			sql, args, err := sqlizer.ToSql()
			So(err, ShouldBeNil)
			So(sql, ShouldEqual,
				`(_access @> '[{"user_id": "userid"}]' OR `+
					`_owner_id = ? OR `+
					`_access @> '[{"public": true}]' OR `+
					`_access IS NULL)`)
			So(args, ShouldResemble, []interface{}{"userid"})
		})

		Convey("serialized for nil user and read", func() {
			sqlizer := &accessPredicateSqlizer{
				nil,
				skydb.ReadLevel,
			}
			sql, args, err := sqlizer.ToSql()
			So(err, ShouldBeNil)
			So(sql, ShouldEqual,
				`(_access @> '[{"public": true}]' OR `+
					`_access IS NULL)`)
			So(args, ShouldResemble, []interface{}{})
		})

		Convey("serialized for nil user and write", func() {
			sqlizer := &accessPredicateSqlizer{
				nil,
				skydb.WriteLevel,
			}
			sql, args, err := sqlizer.ToSql()
			So(err, ShouldBeNil)
			So(sql, ShouldEqual,
				`(_access @> '[{"public": true, "level": "write"}]' OR `+
					`_access IS NULL)`)
			So(args, ShouldResemble, []interface{}{})
		})

		Convey("serialized for role based ACE", func() {
			userinfo := skydb.UserInfo{
				ID:    "userid",
				Roles: []string{"admin", "writer"},
			}
			sqlizer := &accessPredicateSqlizer{
				&userinfo,
				skydb.ReadLevel,
			}
			sql, args, err := sqlizer.ToSql()
			So(err, ShouldBeNil)
			So(sql, ShouldEqual,
				`(_access @> '[{"role": "admin"}]' OR `+
					`_access @> '[{"role": "writer"}]' OR `+
					`_access @> '[{"user_id": "userid"}]' OR `+
					`_owner_id = ? OR `+
					`_access @> '[{"public": true}]' OR `+
					`_access IS NULL)`)
			So(args, ShouldResemble, []interface{}{"userid"})
		})

	})
}
