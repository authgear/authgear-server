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

func TestUserCRUD(t *testing.T) {
	var c *conn

	Convey("Conn", t, func() {
		c = getTestConn(t)

		userinfo := oddb.UserInfo{
			ID:             "userid",
			Email:          "john.doe@example.com",
			HashedPassword: []byte("$2a$10$RbmNb3Rw.PONA2QTcpjBg.1E00zdSI6dWTUwZi.XC0wZm9OhOEvKO"),
			Auth: oddb.AuthInfo{
				"authproto": map[string]interface{}{
					"string": "string",
					"bool":   true,
					"number": float64(1),
				},
			},
		}

		Convey("creates user", func() {
			err := c.CreateUser(&userinfo)
			So(err, ShouldBeNil)

			email := ""
			password := []byte{}
			auth := authInfoValue{}
			err = c.DBMap.Db.QueryRow("SELECT email, password, auth FROM app_com_oursky_ourd._user WHERE id = 'userid'").
				Scan(&email, &password, &auth)
			So(err, ShouldBeNil)

			So(email, ShouldEqual, "john.doe@example.com")
			So(password, ShouldResemble, []byte("$2a$10$RbmNb3Rw.PONA2QTcpjBg.1E00zdSI6dWTUwZi.XC0wZm9OhOEvKO"))
			So(auth, ShouldResemble, authInfoValue{
				"authproto": map[string]interface{}{
					"string": "string",
					"bool":   true,
					"number": float64(1),
				},
			})
		})

		Convey("return ErrUserDuplicated when user to create already exists", func() {
			err := c.CreateUser(&userinfo)
			So(err, ShouldBeNil)

			err = c.CreateUser(&userinfo)
			So(err, ShouldEqual, oddb.ErrUserDuplicated)
		})

		Reset(func() {
			_, err := c.DBMap.Db.Exec("TRUNCATE app_com_oursky_ourd._user")
			So(err, ShouldBeNil)
		})
	})

	cleanupDB(t, c)
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
