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

	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/server/skydb"
)

func TestSqlizer(t *testing.T) {
	Convey("Expression Sqlizer", t, func() {
		Convey("literal null expression", func() {
			expr := newExpressionSqlizer("table", skydb.FieldType{}, skydb.Expression{skydb.Literal, nil})
			sql, args, err := expr.ToSql()
			So(sql, ShouldEqual, "NULL")
			So(args, ShouldResemble, []interface{}{})
			So(err, ShouldBeNil)
		})
	})
}

func TestExpressionSqlizerWithGeometry(t *testing.T) {
	Convey("expression sqlizer with geometry", t, func() {
		Convey("serialized", func() {
			sqlizer := newExpressionSqlizer("note", skydb.FieldType{Type: skydb.TypeGeometry}, skydb.Expression{skydb.KeyPath, "geometry"})
			sqlizer.requireCast = true
			sql, args, err := sqlizer.ToSql()
			So(err, ShouldBeNil)
			So(sql, ShouldEqual,
				`ST_AsGeoJSON("note"."geometry")`)
			So(args, ShouldResemble, []interface{}{})
		})
	})
}
