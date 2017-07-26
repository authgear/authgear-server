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

package builder

import (
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skydb/mock_skydb"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

func TestPredicateSqlizerFactory(t *testing.T) {
	Convey("Expression", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		db := mock_skydb.NewMockDatabase(ctrl)
		db.EXPECT().RemoteColumnTypes(gomock.Eq("note")).
			Return(
				skydb.RecordSchema{
					"title": skydb.FieldType{Type: skydb.TypeString},
				}, nil,
			).AnyTimes()

		f := NewPredicateSqlizerFactory(db, "note").(*predicateSqlizerFactory)

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
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		db := mock_skydb.NewMockDatabase(ctrl)
		db.EXPECT().RemoteColumnTypes(gomock.Eq("note")).
			Return(
				skydb.RecordSchema{
					"title":   skydb.FieldType{Type: skydb.TypeString},
					"content": skydb.FieldType{Type: skydb.TypeString},
				}, nil,
			).AnyTimes()

		f := NewPredicateSqlizerFactory(db, "note").(*predicateSqlizerFactory)

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

	Convey("Distance Predicate", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		db := mock_skydb.NewMockDatabase(ctrl)

		f := NewPredicateSqlizerFactory(db, "note").(*predicateSqlizerFactory)

		expr1 := skydb.Expression{
			skydb.Function,
			skydb.DistanceFunc{
				"latlng",
				skydb.NewLocation(22.25, 114.1667),
			},
		}
		expr2 := skydb.Expression{skydb.Literal, 500.0}

		Convey("within distance", func() {
			for _, predicate := range []skydb.Predicate{
				{skydb.LessThan, []interface{}{expr1, expr2}},
				{skydb.GreaterThan, []interface{}{expr2, expr1}},
			} {
				sqlizer, err := f.newComparisonPredicateSqlizer(predicate)
				So(err, ShouldBeNil)
				So(sqlizer, ShouldResemble, &distancePredicateSqlizer{
					"note",
					"latlng",
					skydb.NewLocation(22.25, 114.1667),
					newExpressionSqlizer("note", skydb.FieldType{Type: skydb.TypeNumber}, skydb.Expression{skydb.Literal, 500.0}),
				})
			}
		})

		Convey("beyond distance", func() {
			sqlizer, err := f.newComparisonPredicateSqlizer(skydb.Predicate{
				skydb.GreaterThan, []interface{}{expr1, expr2},
			})
			So(err, ShouldBeNil)
			So(sqlizer, ShouldResemble, &comparisonPredicateSqlizer{
				[]expressionSqlizer{
					newExpressionSqlizer("note", skydb.FieldType{Type: skydb.TypeNumber}, expr1),
					newExpressionSqlizer("note", skydb.FieldType{Type: skydb.TypeNumber}, expr2),
				},
				skydb.GreaterThan,
			})
		})
	})
}

func TestNotSqlizer(t *testing.T) {
	Convey("NotSqlizer", t, func() {
		Convey("should generate not predicate", func() {
			expSqlizer := newExpressionSqlizer("table", skydb.FieldType{Type: skydb.TypeBoolean}, skydb.Expression{skydb.Literal, false})
			sqlizer := &NotSqlizer{&expSqlizer}
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
			authinfo := skydb.AuthInfo{
				ID: "userid",
			}
			sqlizer := &accessPredicateSqlizer{
				"note",
				&authinfo,
				skydb.ReadLevel,
			}
			sql, args, err := sqlizer.ToSql()
			So(err, ShouldBeNil)
			So(sql, ShouldEqual,
				`("note"."_access" @> '[{"user_id": "userid"}]' OR `+
					`"note"."_owner_id" = ? OR `+
					`"note"."_access" @> '[{"public": true}]' OR `+
					`"note"."_access" IS NULL)`)
			So(args, ShouldResemble, []interface{}{"userid"})
		})

		Convey("serialized for nil user and read", func() {
			sqlizer := &accessPredicateSqlizer{
				"",
				nil,
				skydb.ReadLevel,
			}
			sql, args, err := sqlizer.ToSql()
			So(err, ShouldBeNil)
			So(sql, ShouldEqual,
				`("_access" @> '[{"public": true}]' OR `+
					`"_access" IS NULL)`)
			So(args, ShouldResemble, []interface{}{})
		})

		Convey("serialized for nil user and write", func() {
			sqlizer := &accessPredicateSqlizer{
				"",
				nil,
				skydb.WriteLevel,
			}
			sql, args, err := sqlizer.ToSql()
			So(err, ShouldBeNil)
			So(sql, ShouldEqual,
				`("_access" @> '[{"public": true, "level": "write"}]' OR `+
					`"_access" IS NULL)`)
			So(args, ShouldResemble, []interface{}{})
		})

		Convey("serialized for role based ACE", func() {
			authinfo := skydb.AuthInfo{
				ID:    "userid",
				Roles: []string{"admin", "writer"},
			}
			sqlizer := &accessPredicateSqlizer{
				"",
				&authinfo,
				skydb.ReadLevel,
			}
			sql, args, err := sqlizer.ToSql()
			So(err, ShouldBeNil)
			So(sql, ShouldEqual,
				`("_access" @> '[{"role": "admin"}]' OR `+
					`"_access" @> '[{"role": "writer"}]' OR `+
					`"_access" @> '[{"user_id": "userid"}]' OR `+
					`"_owner_id" = ? OR `+
					`"_access" @> '[{"public": true}]' OR `+
					`"_access" IS NULL)`)
			So(args, ShouldResemble, []interface{}{"userid"})
		})
	})
}

func TestDistancePredicateSqlizer(t *testing.T) {
	Convey("distance predicate", t, func() {
		Convey("serialized", func() {
			sqlizer := &distancePredicateSqlizer{
				"note",
				"latlng",
				skydb.NewLocation(22.25, 114.1667),
				newExpressionSqlizer("note", skydb.FieldType{Type: skydb.TypeNumber}, skydb.Expression{skydb.Literal, 500.0}),
			}
			sql, args, err := sqlizer.ToSql()
			So(err, ShouldBeNil)
			So(sql, ShouldEqual,
				`ST_DWithin("note"."latlng"::geography, ST_MakePoint(?, ?)::geography, ?)`)
			So(args, ShouldResemble, []interface{}{22.25, 114.1667, 500.0})
		})
	})
}
