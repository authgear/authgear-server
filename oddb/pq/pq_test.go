package pq

import (
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"database/sql"
	"time"

	"github.com/lib/pq"
	"github.com/oursky/ourd/oddb"
	. "github.com/oursky/ourd/ourtest"
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
	defaultTo := func(envvar string, value string) {
		if os.Getenv(envvar) == "" {
			os.Setenv(envvar, value)
		}
	}
	defaultTo("PGDATABASE", "ourd_test")
	defaultTo("PGSSLMODE", "disable")
	c, err := Open("com.oursky.ourd", "")
	if err != nil {
		t.Fatal(err)
	}
	return c.(*conn)
}

func cleanupDB(t *testing.T, execori execor) {
	_, err := execori.Exec("DROP SCHEMA app_com_oursky_ourd CASCADE")
	if err != nil && !isInvalidSchemaName(err) {
		t.Fatal(err)
	}
}

func addUser(t *testing.T, c *conn, userid string) {
	_, err := c.Db.Exec("INSERT INTO app_com_oursky_ourd._user (id, password) VALUES ($1, 'somepassword')", userid)
	if err != nil {
		t.Fatal(err)
	}
}

type execor interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

func insertRow(t *testing.T, db execor, query string, args ...interface{}) {
	result, err := db.Exec(query, args...)
	if err != nil {
		t.Fatal(err)
	}

	n, err := result.RowsAffected()
	if err != nil {
		t.Fatal(err)
	}

	if n != 1 {
		t.Fatalf("got rows affected = %v, want 1", n)
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
		defer cleanupDB(t, c.Db)

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

		Convey("query for empty", func() {
			userinfo.Email = ""
			err := c.CreateUser(&userinfo)
			So(err, ShouldBeNil)

			emails := []string{""}
			results, err := c.QueryUser(emails)
			So(err, ShouldBeNil)
			So(len(results), ShouldEqual, 0)
		})

		Convey("query for users", func() {
			err := c.CreateUser(&userinfo)
			So(err, ShouldBeNil)

			userinfo.Email = "jane.doe@example.com"
			userinfo.ID = "userid2"
			So(c.CreateUser(&userinfo), ShouldBeNil)

			emails := []string{"john.doe@example.com", "jane.doe@example.com"}
			results, err := c.QueryUser(emails)
			So(err, ShouldBeNil)

			userids := []string{}
			for _, userinfo := range results {
				userids = append(userids, userinfo.ID)
			}
			So(userids, ShouldContain, "userid")
			So(userids, ShouldContain, "userid2")
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
	})
}

func TestRelation(t *testing.T) {
	Convey("Conn", t, func() {
		c := getTestConn(t)
		defer cleanupDB(t, c.Db)

		addUser(t, c, "userid")
		addUser(t, c, "friendid")

		Convey("add relation", func() {
			err := c.AddRelation("userid", "friend", "friendid")
			So(err, ShouldBeNil)
		})

		Convey("add a user not exist relation", func() {
			err := c.AddRelation("userid", "friend", "non-exist")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "userID not exist")
		})

		Convey("remove non-exist relation", func() {
			err := c.RemoveRelation("userid", "friend", "friendid")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual,
				"friend relation not exist {userid} => {friendid}")
		})

		Convey("remove relation", func() {
			err := c.AddRelation("userid", "friend", "friendid")
			So(err, ShouldBeNil)
			err = c.RemoveRelation("userid", "friend", "friendid")
			So(err, ShouldBeNil)
		})
	})
}

func TestDevice(t *testing.T) {
	Convey("Conn", t, func() {
		c := getTestConn(t)
		defer cleanupDB(t, c.Db)

		addUser(t, c, "userid")

		Convey("gets an existing Device", func() {
			device := oddb.Device{
				ID:         "deviceid",
				Type:       "ios",
				Token:      "devicetoken",
				UserInfoID: "userid",
			}
			So(c.SaveDevice(&device), ShouldBeNil)

			device = oddb.Device{}
			err := c.GetDevice("deviceid", &device)
			So(err, ShouldBeNil)
			So(device, ShouldResemble, oddb.Device{
				ID:         "deviceid",
				Type:       "ios",
				Token:      "devicetoken",
				UserInfoID: "userid",
			})
		})

		Convey("creates a new Device", func() {
			device := oddb.Device{
				ID:         "deviceid",
				Type:       "ios",
				Token:      "devicetoken",
				UserInfoID: "userid",
			}

			err := c.SaveDevice(&device)
			So(err, ShouldBeNil)

			var deviceType, token, userInfoID string
			err = c.Db.QueryRow("SELECT type, token, user_id FROM app_com_oursky_ourd._device WHERE id = 'deviceid'").
				Scan(&deviceType, &token, &userInfoID)
			So(err, ShouldBeNil)
			So(deviceType, ShouldEqual, "ios")
			So(token, ShouldEqual, "devicetoken")
			So(userInfoID, ShouldEqual, "userid")
		})

		Convey("updates an existing Device", func() {
			device := oddb.Device{
				ID:         "deviceid",
				Type:       "ios",
				Token:      "devicetoken",
				UserInfoID: "userid",
			}

			err := c.SaveDevice(&device)
			So(err, ShouldBeNil)

			device.Token = "anotherdevicetoken"
			err = c.SaveDevice(&device)

			So(err, ShouldBeNil)
			var deviceType, token, userInfoID string
			err = c.Db.QueryRow("SELECT type, token, user_id FROM app_com_oursky_ourd._device WHERE id = 'deviceid'").
				Scan(&deviceType, &token, &userInfoID)
			So(err, ShouldBeNil)
			So(deviceType, ShouldEqual, "ios")
			So(token, ShouldEqual, "anotherdevicetoken")
			So(userInfoID, ShouldEqual, "userid")
		})

		Convey("cannot save Device without id", func() {
			device := oddb.Device{
				Type:       "ios",
				Token:      "devicetoken",
				UserInfoID: "userid",
			}

			err := c.SaveDevice(&device)
			So(err, ShouldNotBeNil)
		})

		Convey("cannot save Device without type", func() {
			device := oddb.Device{
				ID:         "deviceid",
				Token:      "devicetoken",
				UserInfoID: "userid",
			}

			err := c.SaveDevice(&device)
			So(err, ShouldNotBeNil)
		})

		Convey("cannot save Device without token", func() {
			device := oddb.Device{
				ID:         "deviceid",
				Type:       "ios",
				UserInfoID: "userid",
			}

			err := c.SaveDevice(&device)
			So(err, ShouldNotBeNil)
		})

		Convey("cannot save Device without user id", func() {
			device := oddb.Device{
				ID:    "deviceid",
				Type:  "ios",
				Token: "devicetoken",
			}

			err := c.SaveDevice(&device)
			So(err, ShouldNotBeNil)
		})

		Convey("deletes an existing record", func() {
			device := oddb.Device{
				ID:         "deviceid",
				Type:       "ios",
				Token:      "devicetoken",
				UserInfoID: "userid",
			}
			So(c.SaveDevice(&device), ShouldBeNil)

			err := c.DeleteDevice("deviceid")
			So(err, ShouldBeNil)

			var count int
			err = c.Db.QueryRow("SELECT COUNT(*) FROM app_com_oursky_ourd._device WHERE id = 'deviceid'").Scan(&count)
			So(err, ShouldBeNil)
			So(count, ShouldEqual, 0)
		})
	})
}

func TestExtend(t *testing.T) {
	Convey("Extend", t, func() {
		c := getTestConn(t)
		defer cleanupDB(t, c.Db)

		db := c.PublicDB()

		Convey("creates table if not exist", func() {
			err := db.Extend("note", oddb.RecordSchema{
				"content":   oddb.FieldType{Type: oddb.TypeString},
				"noteOrder": oddb.FieldType{Type: oddb.TypeNumber},
				"createdAt": oddb.FieldType{Type: oddb.TypeDateTime},
			})
			So(err, ShouldBeNil)

			// verify with an insert
			result, err := c.Db.Exec(
				`INSERT INTO app_com_oursky_ourd."note" ` +
					`(_id, _database_id, _owner_id, "content", "noteOrder", "createdAt") ` +
					`VALUES (1, 1, 1, 'some content', 2, '1988-02-06')`)
			So(err, ShouldBeNil)

			i, err := result.RowsAffected()
			So(err, ShouldBeNil)
			So(i, ShouldEqual, 1)
		})

		Convey("creates table with JSON field", func() {
			err := db.Extend("note", oddb.RecordSchema{
				"tags": oddb.FieldType{Type: oddb.TypeJSON},
			})
			So(err, ShouldBeNil)

			result, err := c.Db.Exec(
				`INSERT INTO app_com_oursky_ourd."note" ` +
					`(_id, _database_id, _owner_id, "tags") ` +
					`VALUES (1, 1, 1, '["tag0", "tag1"]')`)
			So(err, ShouldBeNil)

			i, err := result.RowsAffected()
			So(err, ShouldBeNil)
			So(i, ShouldEqual, 1)
		})

		Convey("creates table with asset", func() {
			err := db.Extend("note", oddb.RecordSchema{
				"image": oddb.FieldType{Type: oddb.TypeAsset},
			})
			So(err, ShouldBeNil)
		})

		Convey("creates table with multiple assets", func() {
			err := db.Extend("note", oddb.RecordSchema{
				"image0": oddb.FieldType{Type: oddb.TypeAsset},
			})
			So(err, ShouldBeNil)
			err = db.Extend("note", oddb.RecordSchema{
				"image1": oddb.FieldType{Type: oddb.TypeAsset},
			})
			So(err, ShouldBeNil)
		})

		Convey("creates table with reference", func() {
			err := db.Extend("collection", oddb.RecordSchema{
				"name": oddb.FieldType{Type: oddb.TypeString},
			})
			So(err, ShouldBeNil)
			err = db.Extend("note", oddb.RecordSchema{
				"content": oddb.FieldType{Type: oddb.TypeString},
				"collection": oddb.FieldType{
					Type:          oddb.TypeReference,
					ReferenceType: "collection",
				},
			})
			So(err, ShouldBeNil)
		})

		Convey("error if creates table with reference not exist", func() {
			err := db.Extend("note", oddb.RecordSchema{
				"content": oddb.FieldType{Type: oddb.TypeString},
				"tag": oddb.FieldType{
					Type:          oddb.TypeReference,
					ReferenceType: "tag",
				},
			})
			So(err, ShouldNotBeNil)
		})

		Convey("adds new column if table already exist", func() {
			err := db.Extend("note", oddb.RecordSchema{
				"content":   oddb.FieldType{Type: oddb.TypeString},
				"noteOrder": oddb.FieldType{Type: oddb.TypeNumber},
				"createdAt": oddb.FieldType{Type: oddb.TypeDateTime},
			})
			So(err, ShouldBeNil)

			err = db.Extend("note", oddb.RecordSchema{
				"createdAt": oddb.FieldType{Type: oddb.TypeDateTime},
				"dirty":     oddb.FieldType{Type: oddb.TypeBoolean},
			})
			So(err, ShouldBeNil)

			// verify with an insert
			result, err := c.Db.Exec(
				`INSERT INTO app_com_oursky_ourd."note" ` +
					`(_id, _database_id, _owner_id, "content", "noteOrder", "createdAt", "dirty") ` +
					`VALUES (1, 1, 1, 'some content', 2, '1988-02-06', TRUE)`)
			So(err, ShouldBeNil)

			i, err := result.RowsAffected()
			So(err, ShouldBeNil)
			So(i, ShouldEqual, 1)
		})

		Convey("errors if conflict with existing column type", func() {
			err := db.Extend("note", oddb.RecordSchema{
				"content":   oddb.FieldType{Type: oddb.TypeString},
				"noteOrder": oddb.FieldType{Type: oddb.TypeNumber},
				"createdAt": oddb.FieldType{Type: oddb.TypeDateTime},
			})
			So(err, ShouldBeNil)

			err = db.Extend("note", oddb.RecordSchema{
				"content":   oddb.FieldType{Type: oddb.TypeNumber},
				"createdAt": oddb.FieldType{Type: oddb.TypeDateTime},
				"dirty":     oddb.FieldType{Type: oddb.TypeNumber},
			})
			So(err.Error(), ShouldEqual, "conflicting schema {TypeString } => {TypeNumber }")
		})
	})
}

func TestGet(t *testing.T) {
	Convey("Database", t, func() {
		c := getTestConn(t)
		defer cleanupDB(t, c.Db)

		db := c.PrivateDB("getuser")
		So(db.Extend("record", oddb.RecordSchema{
			"string":   oddb.FieldType{Type: oddb.TypeString},
			"number":   oddb.FieldType{Type: oddb.TypeNumber},
			"datetime": oddb.FieldType{Type: oddb.TypeDateTime},
			"boolean":  oddb.FieldType{Type: oddb.TypeBoolean},
		}), ShouldBeNil)

		insertRow(t, c.Db, `INSERT INTO app_com_oursky_ourd."record" `+
			`(_database_id, _id, _owner_id, "string", "number", "datetime", "boolean") `+
			`VALUES ('getuser', 'id0', 'getuser', 'string', 1, '1988-02-06', TRUE)`)
		insertRow(t, c.Db, `INSERT INTO app_com_oursky_ourd."record" `+
			`(_database_id, _id, _owner_id, "string", "number", "datetime", "boolean") `+
			`VALUES ('getuser', 'id1', 'getuser', 'string', 1, '1988-02-06', TRUE)`)

		Convey("gets an existing record from database", func() {
			record := oddb.Record{}
			err := db.Get(oddb.NewRecordID("record", "id1"), &record)
			So(err, ShouldBeNil)

			So(record.ID, ShouldResemble, oddb.NewRecordID("record", "id1"))
			So(record.Data["string"], ShouldEqual, "string")
			So(record.Data["number"], ShouldEqual, 1)
			So(record.Data["boolean"], ShouldEqual, true)

			dt, _ := record.Data["datetime"].(time.Time)
			So(dt.Unix(), ShouldEqual, time.Date(1988, 2, 6, 0, 0, 0, 0, time.UTC).Unix())

			So(record.DatabaseID, ShouldEqual, "getuser")
		})

		Convey("errors if gets a non-existing record", func() {
			record := oddb.Record{}
			err := db.Get(oddb.NewRecordID("record", "notexistid"), &record)
			So(err, ShouldEqual, oddb.ErrRecordNotFound)
		})
	})
}

func TestSave(t *testing.T) {
	var c *conn
	Convey("Database", t, func() {
		c = getTestConn(t)
		defer cleanupDB(t, c.Db)

		db := c.PublicDB()
		So(db.Extend("note", oddb.RecordSchema{
			"content":   oddb.FieldType{Type: oddb.TypeString},
			"number":    oddb.FieldType{Type: oddb.TypeNumber},
			"timestamp": oddb.FieldType{Type: oddb.TypeDateTime},
		}), ShouldBeNil)

		record := oddb.Record{
			ID:      oddb.NewRecordID("note", "someid"),
			OwnerID: "user_id",
			Data: map[string]interface{}{
				"content":   "some content",
				"number":    float64(1),
				"timestamp": time.Date(1988, 2, 6, 1, 1, 1, 1, time.UTC),
			},
		}

		Convey("creates record if it doesn't exist", func() {
			err := db.Save(&record)
			So(err, ShouldBeNil)
			So(record.DatabaseID, ShouldEqual, "")

			var (
				content   string
				number    float64
				timestamp time.Time
				ownerID   string
			)
			err = c.Db.QueryRow(
				"SELECT content, number, timestamp, _owner_id "+
					"FROM app_com_oursky_ourd.note WHERE _id = 'someid' and _database_id = ''").
				Scan(&content, &number, &timestamp, &ownerID)
			So(err, ShouldBeNil)
			So(content, ShouldEqual, "some content")
			So(number, ShouldEqual, float64(1))
			So(timestamp.In(time.UTC), ShouldResemble, time.Date(1988, 2, 6, 1, 1, 1, 0, time.UTC))
			So(ownerID, ShouldEqual, "user_id")
		})

		Convey("updates record if it already exists", func() {
			err := db.Save(&record)
			So(err, ShouldBeNil)
			So(record.DatabaseID, ShouldEqual, "")

			record.Set("content", "more content")
			err = db.Save(&record)
			So(err, ShouldBeNil)

			var content string
			err = c.Db.QueryRow("SELECT content FROM app_com_oursky_ourd.note WHERE _id = 'someid' and _database_id = ''").
				Scan(&content)
			So(err, ShouldBeNil)
			So(content, ShouldEqual, "more content")
		})

		Convey("error if saving with recordid already taken by other user", func() {
			ownerDB := c.PrivateDB("ownerid")
			err := ownerDB.Save(&record)
			So(err, ShouldBeNil)
			otherDB := c.PrivateDB("otheruserid")
			err = otherDB.Save(&record)
			// FIXME: Wrap me with oddb.ErrXXX
			So(err, ShouldNotBeNil)
		})

		Convey("ignore Record.DatabaseID when saving", func() {
			record.DatabaseID = "someuserid"
			err := db.Save(&record)
			So(err, ShouldBeNil)
			So(record.DatabaseID, ShouldEqual, "")

			var count int
			err = c.Db.QueryRowx("SELECT count(*) FROM app_com_oursky_ourd.note WHERE _id = 'someid' and _database_id = 'someuserid'").
				Scan(&count)
			So(err, ShouldBeNil)
			So(count, ShouldEqual, 0)
		})

		Convey("REGRESSION: update record with attribute having capital letters", func() {
			So(db.Extend("note", oddb.RecordSchema{
				"noteOrder": oddb.FieldType{Type: oddb.TypeNumber},
			}), ShouldBeNil)

			record = oddb.Record{
				ID:      oddb.NewRecordID("note", "1"),
				OwnerID: "user_id",
				Data: map[string]interface{}{
					"noteOrder": 1,
				},
			}

			ShouldBeNil(db.Save(&record))

			record.Data["noteOrder"] = 2
			ShouldBeNil(db.Save(&record))

			var noteOrder int
			err := c.Db.QueryRow(`SELECT "noteOrder" FROM app_com_oursky_ourd.note WHERE _id = '1' and _database_id = ''`).
				Scan(&noteOrder)
			So(err, ShouldBeNil)
			So(noteOrder, ShouldEqual, 2)
		})

		Convey("errors if OwnerID not set", func() {
			record.OwnerID = ""
			err := db.Save(&record)
			So(err.Error(), ShouldEndWith, "got empty OwnerID")
		})

		Convey("ignore OwnerID when update", func() {
			err := db.Save(&record)
			So(err, ShouldBeNil)

			record.OwnerID = "user_id2"
			So(err, ShouldBeNil)

			var ownerID string
			err = c.Db.QueryRow(`SELECT "_owner_id" FROM app_com_oursky_ourd.note WHERE _id = 'someid' and _database_id = ''`).
				Scan(&ownerID)
			So(ownerID, ShouldEqual, "user_id")
		})
	})
}

func TestJSON(t *testing.T) {
	Convey("Database", t, func() {
		c := getTestConn(t)
		defer cleanupDB(t, c.Db)

		db := c.PublicDB()
		So(db.Extend("note", oddb.RecordSchema{
			"jsonfield": oddb.FieldType{Type: oddb.TypeJSON},
		}), ShouldBeNil)

		Convey("fetch record with json field", func() {
			So(db.Extend("record", oddb.RecordSchema{
				"array":      oddb.FieldType{Type: oddb.TypeJSON},
				"dictionary": oddb.FieldType{Type: oddb.TypeJSON},
			}), ShouldBeNil)

			insertRow(t, c.Db, `INSERT INTO app_com_oursky_ourd."record" `+
				`(_database_id, _id, _owner_id, "array", "dictionary") `+
				`VALUES ('', 'id', 'owner_id', '[1, "string", true]', '{"number": 0, "string": "value", "bool": false}')`)

			var record oddb.Record
			err := db.Get(oddb.NewRecordID("record", "id"), &record)
			So(err, ShouldBeNil)

			So(record, ShouldResemble, oddb.Record{
				ID:      oddb.NewRecordID("record", "id"),
				OwnerID: "owner_id",
				Data: map[string]interface{}{
					"array": []interface{}{float64(1), "string", true},
					"dictionary": map[string]interface{}{
						"number": float64(0),
						"string": "value",
						"bool":   false,
					},
				},
			})
		})

		Convey("saves record field with array", func() {
			record := oddb.Record{
				ID:      oddb.NewRecordID("note", "1"),
				OwnerID: "user_id",
				Data: map[string]interface{}{
					"jsonfield": []interface{}{0.0, "string", true},
				},
			}

			So(db.Save(&record), ShouldBeNil)

			var jsonBytes []byte
			err := c.Db.QueryRow(`SELECT jsonfield FROM app_com_oursky_ourd.note WHERE _id = '1' and _database_id = ''`).
				Scan(&jsonBytes)
			So(err, ShouldBeNil)
			So(jsonBytes, ShouldEqualJSON, `[0, "string", true]`)
		})

		Convey("saves record field with dictionary", func() {
			record := oddb.Record{
				ID:      oddb.NewRecordID("note", "1"),
				OwnerID: "user_id",
				Data: map[string]interface{}{
					"jsonfield": map[string]interface{}{
						"number": float64(1),
						"string": "",
						"bool":   false,
					},
				},
			}

			So(db.Save(&record), ShouldBeNil)

			var jsonBytes []byte
			err := c.Db.QueryRow(`SELECT jsonfield FROM app_com_oursky_ourd.note WHERE _id = '1' and _database_id = ''`).
				Scan(&jsonBytes)
			So(err, ShouldBeNil)
			So(jsonBytes, ShouldEqualJSON, `{"number": 1, "string": "", "bool": false}`)
		})
	})
}

func TestRecordAssetField(t *testing.T) {
	Convey("Record Asset", t, func() {
		c := getTestConn(t)
		defer cleanupDB(t, c.Db)

		So(c.SaveAsset(&oddb.Asset{
			Name:        "picture.png",
			ContentType: "image/png",
			Size:        1,
		}), ShouldBeNil)

		db := c.PublicDB()
		So(db.Extend("note", oddb.RecordSchema{
			"image": oddb.FieldType{Type: oddb.TypeAsset},
		}), ShouldBeNil)

		Convey("can be associated", func() {
			err := db.Save(&oddb.Record{
				ID: oddb.NewRecordID("note", "id"),
				Data: map[string]interface{}{
					"image": oddb.Asset{Name: "picture.png"},
				},
				OwnerID: "user_id",
			})
			So(err, ShouldBeNil)
		})

		Convey("errors when associated with non-existing asset", func() {
			err := db.Save(&oddb.Record{
				ID: oddb.NewRecordID("note", "id"),
				Data: map[string]interface{}{
					"image": oddb.Asset{Name: "notexist.png"},
				},
				OwnerID: "user_id",
			})
			So(err, ShouldNotBeNil)
		})
	})
}

func TestDelete(t *testing.T) {
	var c *conn
	Convey("Database", t, func() {
		c = getTestConn(t)
		defer cleanupDB(t, c.Db)

		db := c.PrivateDB("userid")

		So(db.Extend("note", oddb.RecordSchema{
			"content": oddb.FieldType{Type: oddb.TypeString},
		}), ShouldBeNil)

		record := oddb.Record{
			ID:      oddb.NewRecordID("note", "someid"),
			OwnerID: "user_id",
			Data: map[string]interface{}{
				"content": "some content",
			},
		}

		Convey("deletes existing record", func() {
			err := db.Save(&record)
			So(err, ShouldBeNil)

			err = db.Delete(oddb.NewRecordID("note", "someid"))
			So(err, ShouldBeNil)

			err = db.(*database).Db.QueryRow("SELECT * FROM app_com_oursky_ourd.note WHERE _id = 'someid' AND _database_id = 'userid'").Scan((*string)(nil))
			So(err, ShouldEqual, sql.ErrNoRows)
		})

		Convey("returns ErrRecordNotFound when record to delete doesn't exist", func() {
			err := db.Delete(oddb.NewRecordID("note", "notexistid"))
			So(err, ShouldEqual, oddb.ErrRecordNotFound)
		})

		Convey("return ErrRecordNotFound when deleting other user record", func() {
			err := db.Save(&record)
			So(err, ShouldBeNil)
			otherDB := c.PrivateDB("otheruserid")
			err = otherDB.Delete(oddb.NewRecordID("note", "someid"))
			So(err, ShouldEqual, oddb.ErrRecordNotFound)
		})
	})
}

func TestQuery(t *testing.T) {
	Convey("Database", t, func() {
		c := getTestConn(t)
		defer cleanupDB(t, c.Db)

		// fixture
		record1 := oddb.Record{
			ID:      oddb.NewRecordID("note", "id1"),
			OwnerID: "user_id",
			Data: map[string]interface{}{
				"noteOrder": float64(1),
			},
		}
		record2 := oddb.Record{
			ID:      oddb.NewRecordID("note", "id2"),
			OwnerID: "user_id",
			Data: map[string]interface{}{
				"noteOrder": float64(2),
			},
		}
		record3 := oddb.Record{
			ID:      oddb.NewRecordID("note", "id3"),
			OwnerID: "user_id",
			Data: map[string]interface{}{
				"noteOrder": float64(3),
			},
		}

		db := c.PrivateDB("userid")
		So(db.Extend("note", oddb.RecordSchema{
			"noteOrder": oddb.FieldType{Type: oddb.TypeNumber},
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

		Convey("query records by note order", func() {
			query := oddb.Query{
				Type: "note",
				Predicate: &oddb.Predicate{
					Operator: oddb.Equal,
					Children: []interface{}{
						oddb.Expression{
							Type:  oddb.KeyPath,
							Value: "noteOrder",
						},
						oddb.Expression{
							Type:  oddb.Literal,
							Value: 1,
						},
					},
				},
			}
			records, err := exhaustRows(db.Query(&query))

			So(err, ShouldBeNil)
			So(records[0], ShouldResemble, record1)
			So(len(records), ShouldEqual, 1)
		})

		Convey("query records by note order using or predicate", func() {
			keyPathExpr := oddb.Expression{
				Type:  oddb.KeyPath,
				Value: "noteOrder",
			}
			value1 := oddb.Expression{
				Type:  oddb.Literal,
				Value: 2,
			}
			value2 := oddb.Expression{
				Type:  oddb.Literal,
				Value: 3,
			}
			query := oddb.Query{
				Type: "note",
				Predicate: &oddb.Predicate{
					Operator: oddb.Or,
					Children: []interface{}{
						oddb.Predicate{
							Operator: oddb.Equal,
							Children: []interface{}{keyPathExpr, value1},
						},
						oddb.Predicate{
							Operator: oddb.Equal,
							Children: []interface{}{keyPathExpr, value2},
						},
					},
				},
			}
			records, err := exhaustRows(db.Query(&query))

			So(err, ShouldBeNil)
			So(records[0], ShouldResemble, record2)
			So(records[1], ShouldResemble, record3)
			So(len(records), ShouldEqual, 2)
		})
	})

	Convey("Database with reference", t, func() {
		c := getTestConn(t)
		defer cleanupDB(t, c.Db)

		// fixture
		record1 := oddb.Record{
			ID:      oddb.NewRecordID("note", "id1"),
			OwnerID: "user_id",
			Data: map[string]interface{}{
				"noteOrder": float64(1),
			},
		}
		record2 := oddb.Record{
			ID:      oddb.NewRecordID("note", "id2"),
			OwnerID: "user_id",
			Data: map[string]interface{}{
				"noteOrder": float64(2),
				"category":  oddb.NewReference("category", "important"),
			},
		}
		record3 := oddb.Record{
			ID:      oddb.NewRecordID("note", "id3"),
			OwnerID: "user_id",
			Data: map[string]interface{}{
				"noteOrder": float64(3),
				"category":  oddb.NewReference("category", "funny"),
			},
		}
		category1 := oddb.Record{
			ID:      oddb.NewRecordID("category", "important"),
			OwnerID: "user_id",
			Data:    map[string]interface{}{},
		}
		category2 := oddb.Record{
			ID:      oddb.NewRecordID("category", "funny"),
			OwnerID: "user_id",
			Data:    map[string]interface{}{},
		}

		db := c.PrivateDB("userid")
		So(db.Extend("category", oddb.RecordSchema{}), ShouldBeNil)
		So(db.Extend("note", oddb.RecordSchema{
			"noteOrder": oddb.FieldType{Type: oddb.TypeNumber},
			"category": oddb.FieldType{
				Type:          oddb.TypeReference,
				ReferenceType: "category",
			},
		}), ShouldBeNil)

		err := db.Save(&category1)
		So(err, ShouldBeNil)
		err = db.Save(&category2)
		So(err, ShouldBeNil)
		err = db.Save(&record2)
		So(err, ShouldBeNil)
		err = db.Save(&record1)
		So(err, ShouldBeNil)
		err = db.Save(&record3)
		So(err, ShouldBeNil)

		Convey("query records by reference", func() {
			query := oddb.Query{
				Type: "note",
				Predicate: &oddb.Predicate{
					Operator: oddb.Equal,
					Children: []interface{}{
						oddb.Expression{
							Type:  oddb.KeyPath,
							Value: "category",
						},
						oddb.Expression{
							Type:  oddb.Literal,
							Value: oddb.NewReference("category", "important"),
						},
					},
				},
			}
			records, err := exhaustRows(db.Query(&query))

			So(err, ShouldBeNil)
			So(records[0], ShouldResemble, record2)
			So(len(records), ShouldEqual, 1)
		})
	})

	Convey("Empty Conn", t, func() {
		c := getTestConn(t)
		defer cleanupDB(t, c.Db)

		Convey("gets no users", func() {
			userinfo := oddb.UserInfo{}
			err := c.GetUser("notexistuserid", &userinfo)
			So(err, ShouldEqual, oddb.ErrUserNotFound)
		})

		Convey("query no users", func() {
			emails := []string{"user@example.com"}
			result, err := c.QueryUser(emails)
			So(err, ShouldBeNil)
			So(len(result), ShouldEqual, 0)
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

		Convey("gets no devices", func() {
			device := oddb.Device{}
			err := c.GetDevice("notexistdeviceid", &device)
			So(err, ShouldEqual, oddb.ErrDeviceNotFound)
		})

		Convey("deletes no devices", func() {
			err := c.DeleteDevice("notexistdeviceid")
			So(err, ShouldEqual, oddb.ErrDeviceNotFound)
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
		})
	})
}
