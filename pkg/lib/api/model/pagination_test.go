package model_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/api/model"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
)

func TestPageCursor(t *testing.T) {
	Convey("PageCursor", t, func() {
		Convey("should round-trip correctly", func() {
			cursor, err := model.NewCursor("query-key", "id")
			So(err, ShouldBeNil)
			So(cursor, ShouldEqual, "eyJrZXkiOiJxdWVyeS1rZXkiLCJpZCI6ImlkIn0")

			key, err := cursor.AsDBKey()
			So(err, ShouldBeNil)
			So(key, ShouldResemble, &db.PageKey{
				Key: "query-key",
				ID:  "id",
			})
		})
		Convey("should return nil DB key if empty", func() {
			key, err := model.PageCursor("").AsDBKey()
			So(err, ShouldBeNil)
			So(key, ShouldBeNil)
		})
	})
}
