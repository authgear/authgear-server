package db_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/infra/db"
)

func TestQueryPage(t *testing.T) {
	Convey("QueryPage", t, func() {
		builder := db.NewSQLBuilder("", "", "my-app").Tenant().
			Select("id", "value").
			From("values")

		doQuery := db.QueryPage(db.QueryPageConfig{
			KeyColumn: "value",
			IDColumn:  "id",
		})
		query := func(after, before *db.PageKey, first, last *uint64) (string, []interface{}) {
			query, err := doQuery(builder, after, before, first, last)
			if err != nil {
				panic(err)
			}
			sql, args, err := query.ToSql()
			if err != nil {
				panic(err)
			}
			return sql, args
		}
		newInt := func(v uint64) *uint64 { return &v }

		var sql string
		var args []interface{}

		sql, args = query(nil, nil, nil, nil)
		So(sql, ShouldEqual, `SELECT id, value FROM values WHERE app_id = $1 ORDER BY value, id`)
		So(args, ShouldResemble, []interface{}{"my-app"})

		sql, args = query(nil, nil, newInt(0), nil)
		So(sql, ShouldEqual, `SELECT id, value FROM values WHERE app_id = $1 ORDER BY value, id LIMIT 0`)
		So(args, ShouldResemble, []interface{}{"my-app"})

		sql, args = query(nil, nil, newInt(10), nil)
		So(sql, ShouldEqual, `SELECT id, value FROM values WHERE app_id = $1 ORDER BY value, id LIMIT 10`)
		So(args, ShouldResemble, []interface{}{"my-app"})

		sql, args = query(&db.PageKey{Key: "1", ID: "2"}, nil, newInt(10), nil)
		So(sql, ShouldEqual, `SELECT id, value FROM values WHERE app_id = $1 AND value > $2 OR (value = $3 AND id > $4) ORDER BY value, id LIMIT 10`)
		So(args, ShouldResemble, []interface{}{"my-app", "1", "1", "2"})

		sql, args = query(nil, nil, nil, newInt(10))
		So(sql, ShouldEqual, `SELECT page.* FROM (SELECT id, value FROM values WHERE app_id = $1 ORDER BY value DESC, id DESC LIMIT 10) AS page ORDER BY value, id`)
		So(args, ShouldResemble, []interface{}{"my-app"})

		sql, args = query(nil, &db.PageKey{Key: "3", ID: "4"}, nil, newInt(10))
		So(sql, ShouldEqual, `SELECT page.* FROM (SELECT id, value FROM values WHERE app_id = $1 AND value < $2 OR (value = $3 AND id < $4) ORDER BY value DESC, id DESC LIMIT 10) AS page ORDER BY value, id`)
		So(args, ShouldResemble, []interface{}{"my-app", "3", "3", "4"})

		sql, args = query(&db.PageKey{Key: "1", ID: "2"}, &db.PageKey{Key: "9", ID: "10"}, nil, nil)
		So(sql, ShouldEqual, `SELECT id, value FROM values WHERE app_id = $1 AND value > $2 OR (value = $3 AND id > $4) AND value < $5 OR (value = $6 AND id < $7) ORDER BY value, id`)
		So(args, ShouldResemble, []interface{}{"my-app", "1", "1", "2", "9", "9", "10"})
	})
}
