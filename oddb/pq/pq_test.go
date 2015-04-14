package pq

import (
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/oursky/ourd/oddb"
	. "github.com/smartystreets/goconvey/convey"
	"time"

	"testing"
)

func getTestConn(t *testing.T) *conn {
	c, err := Open("com.oursky.ourd", "host=127.0.0.1 dbname=ourd_test sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	return c.(*conn)
}

func cleanupDB(t *testing.T, c *conn) {
	_, err := c.Db.Exec("DROP SCHEMA app_com_oursky_ourd CASCADE")
	if err != nil {
		t.Fatal(err)
	}
}

func exhaustRows(rows *oddb.Rows, errin error) (records []oddb.Record, err error) {
	if errin != nil {
		err = errin
		return
	}

	for rows.Scan() {
		records = append(records, rows.Record())
	}

	err = rows.Err()
	return
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
			err = c.Db.QueryRow("SELECT email, password, auth FROM app_com_oursky_ourd._user WHERE id = 'userid'").
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

		Convey("returns ErrUserDuplicated when user to create already exists", func() {
			err := c.CreateUser(&userinfo)
			So(err, ShouldBeNil)

			err = c.CreateUser(&userinfo)
			So(err, ShouldEqual, oddb.ErrUserDuplicated)
		})

		Convey("gets an existing User", func() {
			err := c.CreateUser(&userinfo)
			So(err, ShouldBeNil)

			fetcheduserinfo := oddb.UserInfo{}
			err = c.GetUser("userid", &fetcheduserinfo)
			So(err, ShouldBeNil)

			So(fetcheduserinfo, ShouldResemble, userinfo)
		})

		Convey("returns ErrUserNotFound when the user does not exist", func() {
			err := c.GetUser("userid", (*oddb.UserInfo)(nil))
			So(err, ShouldEqual, oddb.ErrUserNotFound)
		})

		Convey("updates a user", func() {
			err := c.CreateUser(&userinfo)
			So(err, ShouldBeNil)

			userinfo.Email = "jane.doe@example.com"

			err = c.UpdateUser(&userinfo)
			So(err, ShouldBeNil)

			updateduserinfo := userInfo{}
			err = c.Db.Get(&updateduserinfo, "SELECT id, email, password, auth FROM app_com_oursky_ourd._user WHERE id = $1", "userid")
			So(err, ShouldBeNil)
			So(updateduserinfo, ShouldResemble, userInfo{
				ID:             "userid",
				Email:          "jane.doe@example.com",
				HashedPassword: []byte("$2a$10$RbmNb3Rw.PONA2QTcpjBg.1E00zdSI6dWTUwZi.XC0wZm9OhOEvKO"),
				Auth: authInfoValue{
					"authproto": map[string]interface{}{
						"string": "string",
						"bool":   true,
						"number": float64(1),
					},
				},
			})
		})

		Convey("returns ErrUserNotFound when the user to update does not exist", func() {
			err := c.UpdateUser(&userinfo)
			So(err, ShouldEqual, oddb.ErrUserNotFound)
		})

		Convey("deletes an existing user", func() {
			err := c.CreateUser(&userinfo)
			So(err, ShouldBeNil)

			err = c.DeleteUser("userid")
			So(err, ShouldBeNil)

			placeholder := []byte{}
			err = c.Db.QueryRow("SELECT false FROM app_com_oursky_ourd._user WHERE id = $1", "userid").Scan(&placeholder)
			So(err, ShouldEqual, sql.ErrNoRows)
			So(placeholder, ShouldBeEmpty)
		})

		Convey("returns ErrUserNotFound when the user to delete does not exist", func() {
			err := c.DeleteUser("notexistid")
			So(err, ShouldEqual, oddb.ErrUserNotFound)
		})

		Convey("deletes only the desired user", func() {
			userinfo.ID = "1"
			err := c.CreateUser(&userinfo)
			So(err, ShouldBeNil)

			userinfo.ID = "2"
			err = c.CreateUser(&userinfo)
			So(err, ShouldBeNil)

			count := 0
			c.Db.QueryRow("SELECT COUNT(*) FROM app_com_oursky_ourd._user").Scan(&count)
			So(count, ShouldEqual, 2)

			err = c.DeleteUser("2")
			So(err, ShouldBeNil)

			c.Db.QueryRow("SELECT COUNT(*) FROM app_com_oursky_ourd._user").Scan(&count)
			So(count, ShouldEqual, 1)
		})

		Reset(func() {
			_, err := c.Db.Exec("TRUNCATE app_com_oursky_ourd._user")
			So(err, ShouldBeNil)
		})
	})

	cleanupDB(t, c)
}

func TestSave(t *testing.T) {
	var c *conn
	Convey("Database", t, func() {
		c = getTestConn(t)
		db := c.PublicDB()

		record := oddb.Record{
			Key:  "someid",
			Type: "note",
			Data: map[string]interface{}{
				"content":   "some content",
				"number":    float64(1),
				"timestamp": time.Date(1988, 2, 6, 1, 1, 1, 1, time.UTC),
			},
		}

		Convey("creates record if it doesn't exist", func() {
			err := db.Save(&record)
			So(err, ShouldBeNil)

			var (
				content   string
				number    float64
				timestamp time.Time
			)
			err = db.(*database).Db.
				QueryRow("SELECT content, number, timestamp FROM app_com_oursky_ourd.note WHERE _id = 'someid' and _user_id = ''").
				Scan(&content, &number, &timestamp)
			So(err, ShouldBeNil)
			So(content, ShouldEqual, "some content")
			So(number, ShouldEqual, float64(1))
			So(timestamp.In(time.UTC), ShouldResemble, time.Date(1988, 2, 6, 1, 1, 1, 0, time.UTC))
		})

		Convey("updates record if it already exists", func() {
			err := db.Save(&record)
			So(err, ShouldBeNil)

			record.Set("content", "more content")
			err = db.Save(&record)
			So(err, ShouldBeNil)

			var content string
			err = db.(*database).Db.QueryRow("SELECT content FROM app_com_oursky_ourd.note WHERE _id = 'someid' and _user_id = ''").Scan(&content)
			So(err, ShouldBeNil)
			So(content, ShouldEqual, "more content")
		})

		Reset(func() {
			_, err := db.(*database).Db.Exec("TRUNCATE app_com_oursky_ourd.note")
			So(err, ShouldBeNil)
		})
	})

	cleanupDB(t, c)
}

func TestDelete(t *testing.T) {
	var c *conn
	Convey("Database", t, func() {
		c = getTestConn(t)
		db := c.PrivateDB("userid")

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

			err = db.(*database).Db.QueryRow("SELECT * FROM app_com_oursky_ourd.note WHERE _id = 'someid' AND _user_id = 'userid'").Scan((*string)(nil))
			So(err, ShouldEqual, sql.ErrNoRows)
		})

		Convey("returns ErrRecordNotFound when record to delete doesn't exist", func() {
			err := db.Delete("notexistid")
			So(err, ShouldEqual, oddb.ErrRecordNotFound)
		})

		Convey("deletes only record of the current user", func() {
			err := db.Save(&record)
			So(err, ShouldBeNil)

			otherDB := c.PrivateDB("otheruserid")
			err = otherDB.Save(&record)
			So(err, ShouldBeNil)

			err = db.Delete("someid")
			So(err, ShouldBeNil)

			count := 0
			err = c.Db.Get(&count, "SELECT COUNT(*) FROM app_com_oursky_ourd.note WHERE _id = 'someid' AND _user_id = 'otheruserid'")
			So(err, ShouldBeNil)
			So(count, ShouldEqual, 1)
		})
	})

	cleanupDB(t, c)
}

func TestQuery(t *testing.T) {
	Convey("Database", t, func() {
		c := getTestConn(t)

		// fixture
		record1 := oddb.Record{
			Key:  "id1",
			Type: "note",
			Data: map[string]interface{}{
				"noteOrder": float64(1),
			},
		}
		record2 := oddb.Record{
			Key:  "id2",
			Type: "note",
			Data: map[string]interface{}{
				"noteOrder": float64(2),
			},
		}
		record3 := oddb.Record{
			Key:  "id3",
			Type: "note",
			Data: map[string]interface{}{
				"noteOrder": float64(3),
			},
		}

		db := c.PrivateDB("userid")
		err := db.Save(&record2)
		So(err, ShouldBeNil)
		err = db.Save(&record1)
		So(err, ShouldBeNil)
		err = db.Save(&record3)
		So(err, ShouldBeNil)

		Convey("queries records", func() {
			query := oddb.Query{
				Type: "note",
			}
			records, err := exhaustRows(db.Query(&query))

			So(err, ShouldBeNil)
			So(records[0], ShouldResemble, record2)
			So(records[1], ShouldResemble, record1)
			So(records[2], ShouldResemble, record3)
			So(len(records), ShouldEqual, 3)
		})

		Convey("sorts queried records ascendingly", func() {
			query := oddb.Query{
				Type: "note",
				Sorts: []oddb.Sort{
					oddb.Sort{
						KeyPath: "noteOrder",
						Order:   oddb.Ascending,
					},
				},
			}
			records, err := exhaustRows(db.Query(&query))

			So(err, ShouldBeNil)
			So(records, ShouldResemble, []oddb.Record{
				record1,
				record2,
				record3,
			})
		})

		Convey("sorts queried records descendingly", func() {
			query := oddb.Query{
				Type: "note",
				Sorts: []oddb.Sort{
					oddb.Sort{
						KeyPath: "noteOrder",
						Order:   oddb.Descending,
					},
				},
			}
			records, err := exhaustRows(db.Query(&query))

			So(err, ShouldBeNil)
			So(records, ShouldResemble, []oddb.Record{
				record3,
				record2,
				record1,
			})
		})

		cleanupDB(t, c)
	})
}
