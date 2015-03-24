package pq

import (
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/oursky/ourd/oddb"
	. "github.com/smartystreets/goconvey/convey"

	"testing"
)

func getTestConn(t *testing.T) *conn {
	c, err := Open("com.oursky.ourd", "dbname=ourd_test sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	return c.(*conn)
}

func cleanupDB(t *testing.T, c *conn) {
	_, err := c.DBMap.Db.Exec("DROP SCHEMA app_com_oursky_ourd CASCADE")
	if err != nil {
		t.Fatal(err)
	}
}

func TestInsert(t *testing.T) {
	var c *conn
	Convey("Database", t, func() {
		c = getTestConn(t)
		db := c.PublicDB()

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

	cleanupDB(t, c)
}

func TestDelete(t *testing.T) {
	var c *conn
	Convey("Database", t, func() {
		c = getTestConn(t)
		db := c.PublicDB()

		record := oddb.Record{
			Key:  "someid",
			Type: "note",
			Data: map[string]interface{}{
				"content": "some content",
			},
		}

		Convey("deletes existing record", func() {
			err := db.Save(&record)
			So(err, ShouldBeNil)

			err = db.Delete("someid")
			So(err, ShouldBeNil)

			err = db.(*database).DBMap.Db.QueryRow("SELECT * FROM app_com_oursky_ourd.note WHERE _id = 'someid'").Scan((*string)(nil))
			So(err, ShouldEqual, sql.ErrNoRows)
		})

		Convey("returns ErrRecordNotFound when record to delete doesn't exist", func() {
			err := db.Delete("notexistid")
			So(err, ShouldEqual, oddb.ErrRecordNotFound)
		})
	})

	cleanupDB(t, c)
}
