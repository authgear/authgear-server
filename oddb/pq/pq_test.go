package pq

import (
	_ "github.com/lib/pq"
	"github.com/oursky/ourd/oddb"
	. "github.com/smartystreets/goconvey/convey"

	"testing"
)

func getTestConn(t *testing.T) oddb.Conn {
	conn, err := Open("com.oursky.ourd", "dbname=ourd_test sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	return conn
}

func TestInsert(t *testing.T) {
	Convey("Database", t, func() {
		conn := getTestConn(t)
		db := conn.PublicDB()

		record := oddb.Record{
			Key:  "someid",
			Type: "note",
			Data: map[string]interface{}{
				"content": "some content",
			},
		}

		Convey("creates record if it doesn't exist", func() {
			err := db.Save(&record)
			So(err, ShouldBeNil)

			var content string
			err = db.(*database).DBMap.Db.QueryRow("SELECT content FROM app_com_oursky_ourd.note WHERE _id = 'someid'").Scan(&content)
			So(err, ShouldBeNil)
			So(content, ShouldEqual, "some content")
		})

		Convey("updates record if it already exists", func() {
			err := db.Save(&record)
			So(err, ShouldBeNil)

			record.Set("content", "more content")
			err = db.Save(&record)
			So(err, ShouldBeNil)

			var content string
			err = db.(*database).DBMap.Db.QueryRow("SELECT content FROM app_com_oursky_ourd.note WHERE _id = 'someid'").Scan(&content)
			So(err, ShouldBeNil)
			So(content, ShouldEqual, "more content")
		})

		Reset(func() {
			_, err := db.(*database).DBMap.Exec("TRUNCATE app_com_oursky_ourd.note")
			So(err, ShouldBeNil)
		})
	})
}
