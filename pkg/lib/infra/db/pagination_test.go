package db_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

func TestApplyPageArgs(t *testing.T) {
	Convey("ApplyPageArgs", t, func() {
		builder := db.NewSQLBuilderApp("", "my-app").
			Select("id", "value").
			From("values")

		query := func(after, before, first, last *uint64) (string, uint64) {
			var a, b *db.PageKey
			if after != nil {
				a = &db.PageKey{Offset: *after}
			}
			if before != nil {
				b = &db.PageKey{Offset: *before}
			}
			aftera, err := a.ToPageCursor()
			if err != nil {
				panic(err)
			}
			beforea, err := b.ToPageCursor()
			if err != nil {
				panic(err)
			}

			query, offset, err := db.ApplyPageArgs(builder, graphqlutil.PageArgs{
				First:  first,
				Last:   last,
				After:  graphqlutil.Cursor(aftera),
				Before: graphqlutil.Cursor(beforea),
			})
			if err != nil {
				panic(err)
			}
			sql, _, err := query.ToSql()
			if err != nil {
				panic(err)
			}
			return sql, offset
		}
		newInt := func(v uint64) *uint64 { return &v }

		var sql string
		var offset uint64

		sql, offset = query(nil, nil, nil, nil)
		So(sql, ShouldEqual, `SELECT id, value FROM values WHERE app_id = $1`)
		So(offset, ShouldEqual, 0)

		sql, offset = query(nil, nil, newInt(0), nil)
		So(sql, ShouldEqual, `SELECT id, value FROM values WHERE app_id = $1 LIMIT 0`)
		So(offset, ShouldEqual, 0)

		sql, offset = query(nil, nil, newInt(10), nil)
		So(sql, ShouldEqual, `SELECT id, value FROM values WHERE app_id = $1 LIMIT 10`)
		So(offset, ShouldEqual, 0)

		sql, offset = query(newInt(9), nil, newInt(10), nil)
		So(sql, ShouldEqual, `SELECT id, value FROM values WHERE app_id = $1 LIMIT 10 OFFSET 10`)
		So(offset, ShouldEqual, 10)

		sql, offset = query(newInt(9), newInt(20), nil, nil)
		So(sql, ShouldEqual, `SELECT id, value FROM values WHERE app_id = $1 LIMIT 10 OFFSET 10`)
		So(offset, ShouldEqual, 10)
	})
}

func TestPageKey(t *testing.T) {
	Convey("PageKey", t, func() {
		Convey("should round-trip correctly", func() {
			pageKey := db.PageKey{Offset: 10}
			cursor, err := pageKey.ToPageCursor()
			So(err, ShouldBeNil)
			So(cursor, ShouldEqual, "b2Zmc2V0OjEw")

			key, err := db.NewFromPageCursor(cursor)
			So(err, ShouldBeNil)
			So(key, ShouldResemble, &db.PageKey{Offset: 10})
		})
		Convey("should return nil DB key if empty", func() {
			key, err := db.NewFromPageCursor("")
			So(err, ShouldBeNil)
			So(key, ShouldBeNil)
		})
	})
}
