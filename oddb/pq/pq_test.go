package pq

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"

	"database/sql"
	log "github.com/Sirupsen/logrus"
	"time"

	"github.com/lib/pq"
	"github.com/oursky/ourd/oddb"
)

// NOTE(limouren): postgresql uses this error to signify a non-exist
// schema
func isInvalidSchemaName(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "3F000" {
		return true
	}

	return false
}

func getTestConn(t *testing.T) *conn {
	c, err := Open("com.oursky.ourd", "host=127.0.0.1 dbname=ourd_test sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	return c.(*conn)
}

func cleanupDB(t *testing.T, c *conn) {
	_, err := c.Db.Exec("DROP SCHEMA app_com_oursky_ourd CASCADE")
	if err != nil && !isInvalidSchemaName(err) {
		t.Fatal(err)
	}
}

type execor interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

func insertRow(db execor, query string, args ...interface{}) {
	result, err := db.Exec(query, args...)
	if err != nil {
		log.WithFields(log.Fields{
			"query": query,
			"args":  args,
			"err":   err,
		}).Panicln("failed to execute insert")
	}

	n, err := result.RowsAffected()
	if err != nil {
		log.WithFields(log.Fields{
			"query": query,
			"args":  args,
			"err":   err,
		}).Panicln("failed to fetch RowsAffected")
	}

	if n != 1 {
		log.WithFields(log.Fields{
			"query": query,
			"args":  args,
			"n":     n,
		}).Panicln("rows affected != 1")
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

func TestExtend(t *testing.T) {
	Convey("Extend", t, func() {
		c := getTestConn(t)
		db := c.PublicDB()

		Convey("creates table if not exist", func() {
			err := db.Extend("note", oddb.RecordSchema{
				"content":   oddb.TypeString,
				"noteOrder": oddb.TypeNumber,
				"createdAt": oddb.TypeDateTime,
			})
			So(err, ShouldBeNil)

			// verify with an insert
			result, err := c.Db.Exec(
				`INSERT INTO app_com_oursky_ourd."note" ` +
					`(_id, _user_id, "content", "noteOrder", "createdAt") ` +
					`VALUES (1, 1, 'some content', 2, '1988-02-06')`)
			So(err, ShouldBeNil)

			i, err := result.RowsAffected()
			So(err, ShouldBeNil)
			So(i, ShouldEqual, 1)
		})

		Convey("adds new column if table already exist", func() {
			err := db.Extend("note", oddb.RecordSchema{
				"content":   oddb.TypeString,
				"noteOrder": oddb.TypeNumber,
				"createdAt": oddb.TypeDateTime,
			})
			So(err, ShouldBeNil)

			err = db.Extend("note", oddb.RecordSchema{
				"createdAt": oddb.TypeDateTime,
				"dirty":     oddb.TypeBoolean,
			})
			So(err, ShouldBeNil)

			// verify with an insert
			result, err := c.Db.Exec(
				`INSERT INTO app_com_oursky_ourd."note" ` +
					`(_id, _user_id, "content", "noteOrder", "createdAt", "dirty") ` +
					`VALUES (1, 1, 'some content', 2, '1988-02-06', TRUE)`)
			So(err, ShouldBeNil)

			i, err := result.RowsAffected()
			So(err, ShouldBeNil)
			So(i, ShouldEqual, 1)
		})

		Convey("errors if conflict with existing column type", func() {
			err := db.Extend("note", oddb.RecordSchema{
				"content":   oddb.TypeString,
				"noteOrder": oddb.TypeNumber,
				"createdAt": oddb.TypeDateTime,
			})
			So(err, ShouldBeNil)

			err = db.Extend("note", oddb.RecordSchema{
				"content":   oddb.TypeNumber,
				"createdAt": oddb.TypeDateTime,
				"dirty":     oddb.TypeNumber,
			})
			So(err.Error(), ShouldEqual, "conflicting dataType TypeString => TypeNumber")
		})

		Reset(func() {
			cleanupDB(t, c)
		})
	})
}

func TestGet(t *testing.T) {
	SkipConvey("Database", t, func() {
		c := getTestConn(t)
		defer cleanupDB(t, c)
		db := c.PrivateDB("getuser")
		So(db.Extend("record", oddb.RecordSchema{
			"string":   oddb.TypeString,
			"number":   oddb.TypeNumber,
			"datetime": oddb.TypeDateTime,
			"boolean":  oddb.TypeBoolean,
		}), ShouldBeNil)

		insertRow(c.Db, `INSERT INTO app_com_oursky_ourd."record" `+
			`(_user_id, _id, "string", "number", "datetime", "boolean") `+
			`VALUES ('getuser', 'id', 'string', 1, '1988-02-06', TRUE)`)

		Convey("gets an existing record from database", func() {
			record := oddb.Record{}
			err := db.Get(oddb.NewRecordID("record", "id"), &record)
			So(err, ShouldBeNil)

			So(record, ShouldResemble, oddb.Record{
				ID: oddb.NewRecordID("record", "id"),
				Data: map[string]interface{}{
					"string":   "string",
					"number":   float64(1),
					"datetime": time.Date(1988, 2, 6, 0, 0, 0, 0, time.UTC),
				},
			})
		})

		Convey("errors if gets a non-existing record", func() {
			record := oddb.Record{}
			err := db.Get(oddb.NewRecordID("record", "notexistid"), &record)
			So(err, ShouldBeNil, oddb.ErrRecordNotFound)
		})
	})
}

func TestSave(t *testing.T) {
	var c *conn
	Convey("Database", t, func() {
		c = getTestConn(t)
		defer cleanupDB(t, c)

		db := c.PublicDB()
		So(db.Extend("note", oddb.RecordSchema{
			"content":   oddb.TypeString,
			"number":    oddb.TypeNumber,
			"timestamp": oddb.TypeDateTime,
		}), ShouldBeNil)

		record := oddb.Record{
			ID: oddb.NewRecordID("note", "someid"),
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
			err = c.Db.QueryRow("SELECT content, number, timestamp FROM app_com_oursky_ourd.note WHERE _id = 'someid' and _user_id = ''").
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
			err = c.Db.QueryRow("SELECT content FROM app_com_oursky_ourd.note WHERE _id = 'someid' and _user_id = ''").
				Scan(&content)
			So(err, ShouldBeNil)
			So(content, ShouldEqual, "more content")
		})

		Convey("REGRESSION: update record with attribute having capital letters", func() {
			So(db.Extend("note", oddb.RecordSchema{
				"noteOrder": oddb.TypeNumber,
			}), ShouldBeNil)

			record = oddb.Record{
				ID: oddb.NewRecordID("note", "1"),
				Data: map[string]interface{}{
					"noteOrder": 1,
				},
			}

			ShouldBeNil(db.Save(&record))

			record.Data["noteOrder"] = 2
			ShouldBeNil(db.Save(&record))

			var noteOrder int
			err := c.Db.QueryRow(`SELECT "noteOrder" FROM app_com_oursky_ourd.note WHERE _id = '1' and _user_id = ''`).
				Scan(&noteOrder)
			So(err, ShouldBeNil)
			So(noteOrder, ShouldEqual, 2)
		})
	})
}

func TestDelete(t *testing.T) {
	var c *conn
	Convey("Database", t, func() {
		c = getTestConn(t)
		db := c.PrivateDB("userid")

		So(db.Extend("note", oddb.RecordSchema{
			"content": oddb.TypeString,
		}), ShouldBeNil)

		record := oddb.Record{
			ID: oddb.NewRecordID("note", "someid"),
			Data: map[string]interface{}{
				"content": "some content",
			},
		}

		Convey("deletes existing record", func() {
			err := db.Save(&record)
			So(err, ShouldBeNil)

			err = db.Delete(oddb.NewRecordID("note", "someid"))
			So(err, ShouldBeNil)

			err = db.(*database).Db.QueryRow("SELECT * FROM app_com_oursky_ourd.note WHERE _id = 'someid' AND _user_id = 'userid'").Scan((*string)(nil))
			So(err, ShouldEqual, sql.ErrNoRows)
		})

		Convey("returns ErrRecordNotFound when record to delete doesn't exist", func() {
			err := db.Delete(oddb.NewRecordID("note", "notexistid"))
			So(err, ShouldEqual, oddb.ErrRecordNotFound)
		})

		Convey("deletes only record of the current user", func() {
			err := db.Save(&record)
			So(err, ShouldBeNil)

			otherDB := c.PrivateDB("otheruserid")
			err = otherDB.Save(&record)
			So(err, ShouldBeNil)

			err = db.Delete(oddb.NewRecordID("note", "someid"))
			So(err, ShouldBeNil)

			count := 0
			err = c.Db.Get(&count, "SELECT COUNT(*) FROM app_com_oursky_ourd.note WHERE _id = 'someid' AND _user_id = 'otheruserid'")
			So(err, ShouldBeNil)
			So(count, ShouldEqual, 1)
		})

		cleanupDB(t, c)
	})
}

func TestQuery(t *testing.T) {
	Convey("Database", t, func() {
		c := getTestConn(t)

		// fixture
		record1 := oddb.Record{
			ID: oddb.NewRecordID("note", "id1"),
			Data: map[string]interface{}{
				"noteOrder": float64(1),
			},
		}
		record2 := oddb.Record{
			ID: oddb.NewRecordID("note", "id2"),
			Data: map[string]interface{}{
				"noteOrder": float64(2),
			},
		}
		record3 := oddb.Record{
			ID: oddb.NewRecordID("note", "id3"),
			Data: map[string]interface{}{
				"noteOrder": float64(3),
			},
		}

		db := c.PrivateDB("userid")
		So(db.Extend("note", oddb.RecordSchema{
			"noteOrder": oddb.TypeNumber,
		}), ShouldBeNil)

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

	Convey("Empty Conn", t, func() {
		c := getTestConn(t)

		Convey("gets no users", func() {
			userinfo := oddb.UserInfo{}
			err := c.GetUser("notexistuserid", &userinfo)
			So(err, ShouldEqual, oddb.ErrUserNotFound)
		})

		Convey("updates no users", func() {
			userinfo := oddb.UserInfo{
				ID: "notexistuserid",
			}
			err := c.UpdateUser(&userinfo)
			So(err, ShouldEqual, oddb.ErrUserNotFound)
		})

		Convey("deletes no users", func() {
			err := c.DeleteUser("notexistuserid")
			So(err, ShouldEqual, oddb.ErrUserNotFound)
		})

		Convey("Empty Database", func() {
			db := c.PublicDB()

			Convey("gets nothing", func() {
				record := oddb.Record{}

				err := db.Get(oddb.NewRecordID("type", "notexistid"), &record)

				So(err, ShouldEqual, oddb.ErrRecordNotFound)
			})

			Convey("deletes nothing", func() {
				err := db.Delete(oddb.NewRecordID("type", "notexistid"))
				So(err, ShouldEqual, oddb.ErrRecordNotFound)
			})

			Convey("queries nothing", func() {
				query := oddb.Query{
					Type: "notexisttype",
				}

				records, err := exhaustRows(db.Query(&query))

				So(err, ShouldBeNil)
				So(records, ShouldBeEmpty)
			})

			cleanupDB(t, c)
		})

		cleanupDB(t, c)
	})
}
