package pq

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/oursky/skygear/skydb"
	. "github.com/oursky/skygear/skytest"
	. "github.com/smartystreets/goconvey/convey"
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
	defaultTo("PGDATABASE", "skygear_test")
	defaultTo("PGSSLMODE", "disable")
	c, err := Open("com.oursky.skygear", skydb.RoleBasedAccess, "")
	if err != nil {
		t.Fatal(err)
	}
	return c.(*conn)
}

func cleanupConn(t *testing.T, c *conn) {
	_, err := c.db.Exec("DROP SCHEMA app_com_oursky_skygear CASCADE")
	if err != nil && !isInvalidSchemaName(err) {
		t.Fatal(err)
	}
}

func addUser(t *testing.T, c *conn, userid string) {
	_, err := c.Exec("INSERT INTO app_com_oursky_skygear._user (id, password) VALUES ($1, 'somepassword')", userid)
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

func exhaustRows(rows *skydb.Rows, errin error) (records []skydb.Record, err error) {
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

func TestExtend(t *testing.T) {
	Convey("remoteColumnTypes", t, func() {
		c := getTestConn(t)
		defer cleanupConn(t, c)
		db := c.PublicDB().(*database)

		Convey("return Resemble RecordSchema on second call", func() {
			err := db.Extend("note", skydb.RecordSchema{
				"content": skydb.FieldType{Type: skydb.TypeString},
			})
			So(err, ShouldBeNil)
			schema, _ := db.remoteColumnTypes("note")
			schema2, _ := db.remoteColumnTypes("note")
			So(schema, ShouldResemble, schema2)
		})

		Convey("return cached RecordSchema instance on second call", func() {
			cachedSchema := skydb.RecordSchema{
				"cached": skydb.FieldType{Type: skydb.TypeString},
			}
			c.RecordSchema["note"] = cachedSchema
			schema, _ := db.remoteColumnTypes("note")
			So(schema, ShouldResemble, cachedSchema)
		})

		Convey("clean the cache of RecordSchema on extend recordType", func() {
			err := db.Extend("note", skydb.RecordSchema{
				"content": skydb.FieldType{Type: skydb.TypeString},
			})
			So(err, ShouldBeNil)
			schema, _ := db.remoteColumnTypes("note")
			err = db.Extend("note", skydb.RecordSchema{
				"description": skydb.FieldType{Type: skydb.TypeString},
			})
			schema2, _ := db.remoteColumnTypes("note")
			So(schema, ShouldNotResemble, schema2)
		})
	})

	Convey("Extend", t, func() {
		c := getTestConn(t)
		defer cleanupConn(t, c)

		db := c.PublicDB()

		Convey("creates table if not exist", func() {
			err := db.Extend("note", skydb.RecordSchema{
				"content":   skydb.FieldType{Type: skydb.TypeString},
				"noteOrder": skydb.FieldType{Type: skydb.TypeNumber},
				"createdAt": skydb.FieldType{Type: skydb.TypeDateTime},
			})
			So(err, ShouldBeNil)

			// verify with an insert
			result, err := c.Exec(
				`INSERT INTO app_com_oursky_skygear."note" ` +
					`(_id, _database_id, _owner_id, _created_at, _created_by, _updated_at, _updated_by, "content", "noteOrder", "createdAt") ` +
					`VALUES (1, 1, 1, '1988-02-06', 'creator', '1988-02-06', 'updater', 'some content', 2, '1988-02-06')`)
			So(err, ShouldBeNil)

			i, err := result.RowsAffected()
			So(err, ShouldBeNil)
			So(i, ShouldEqual, 1)
		})

		Convey("REGRESSION #277: creates table with `:`", func() {
			err := db.Extend("table:name", nil)
			So(err, ShouldBeNil)
		})

		Convey("creates table with JSON field", func() {
			err := db.Extend("note", skydb.RecordSchema{
				"tags": skydb.FieldType{Type: skydb.TypeJSON},
			})
			So(err, ShouldBeNil)

			result, err := c.Exec(
				`INSERT INTO app_com_oursky_skygear."note" ` +
					`(_id, _database_id, _owner_id, _created_at, _created_by, _updated_at, _updated_by, "tags") ` +
					`VALUES (1, 1, 1, '1988-02-06', 'creator', '1988-02-06', 'updater', '["tag0", "tag1"]')`)
			So(err, ShouldBeNil)

			i, err := result.RowsAffected()
			So(err, ShouldBeNil)
			So(i, ShouldEqual, 1)
		})

		Convey("creates table with asset", func() {
			err := db.Extend("note", skydb.RecordSchema{
				"image": skydb.FieldType{Type: skydb.TypeAsset},
			})
			So(err, ShouldBeNil)
		})

		Convey("creates table with multiple assets", func() {
			err := db.Extend("note", skydb.RecordSchema{
				"image0": skydb.FieldType{Type: skydb.TypeAsset},
			})
			So(err, ShouldBeNil)
			err = db.Extend("note", skydb.RecordSchema{
				"image1": skydb.FieldType{Type: skydb.TypeAsset},
			})
			So(err, ShouldBeNil)
		})

		Convey("creates table with reference", func() {
			err := db.Extend("collection", skydb.RecordSchema{
				"name": skydb.FieldType{Type: skydb.TypeString},
			})
			So(err, ShouldBeNil)
			err = db.Extend("note", skydb.RecordSchema{
				"content": skydb.FieldType{Type: skydb.TypeString},
				"collection": skydb.FieldType{
					Type:          skydb.TypeReference,
					ReferenceType: "collection",
				},
			})
			So(err, ShouldBeNil)
		})

		Convey("REGRESSION #318: creates table with `:` with reference", func() {
			err := db.Extend("colon:fever", skydb.RecordSchema{
				"name": skydb.FieldType{Type: skydb.TypeString},
			})
			So(err, ShouldBeNil)
			err = db.Extend("note", skydb.RecordSchema{
				"content": skydb.FieldType{Type: skydb.TypeString},
				"colon:fever": skydb.FieldType{
					Type:          skydb.TypeReference,
					ReferenceType: "colon:fever",
				},
			})
			So(err, ShouldBeNil)
		})

		Convey("creates table with location", func() {
			err := db.Extend("photo", skydb.RecordSchema{
				"location": skydb.FieldType{Type: skydb.TypeLocation},
			})
			So(err, ShouldBeNil)
		})

		Convey("creates table with sequence", func() {
			err := db.Extend("note", skydb.RecordSchema{
				"order": skydb.FieldType{Type: skydb.TypeSequence},
			})
			So(err, ShouldBeNil)
		})

		Convey("extend sequence twice", func() {
			err := db.Extend("note", skydb.RecordSchema{
				"order": skydb.FieldType{Type: skydb.TypeSequence},
			})
			So(err, ShouldBeNil)

			err = db.Extend("note", skydb.RecordSchema{
				"order": skydb.FieldType{Type: skydb.TypeSequence},
			})
			So(err, ShouldBeNil)
		})

		Convey("error if creates table with reference not exist", func() {
			err := db.Extend("note", skydb.RecordSchema{
				"content": skydb.FieldType{Type: skydb.TypeString},
				"tag": skydb.FieldType{
					Type:          skydb.TypeReference,
					ReferenceType: "tag",
				},
			})
			So(err, ShouldNotBeNil)
		})

		Convey("adds new column if table already exist", func() {
			err := db.Extend("note", skydb.RecordSchema{
				"content":   skydb.FieldType{Type: skydb.TypeString},
				"noteOrder": skydb.FieldType{Type: skydb.TypeNumber},
				"createdAt": skydb.FieldType{Type: skydb.TypeDateTime},
			})
			So(err, ShouldBeNil)

			err = db.Extend("note", skydb.RecordSchema{
				"createdAt": skydb.FieldType{Type: skydb.TypeDateTime},
				"dirty":     skydb.FieldType{Type: skydb.TypeBoolean},
			})
			So(err, ShouldBeNil)

			// verify with an insert
			result, err := c.Exec(
				`INSERT INTO app_com_oursky_skygear."note" ` +
					`(_id, _database_id, _owner_id, _created_at, _created_by, _updated_at, _updated_by, "content", "noteOrder", "createdAt", "dirty") ` +
					`VALUES (1, 1, 1, '1988-02-06', 'creator', '1988-02-06', 'updater', 'some content', 2, '1988-02-06', TRUE)`)
			So(err, ShouldBeNil)

			i, err := result.RowsAffected()
			So(err, ShouldBeNil)
			So(i, ShouldEqual, 1)
		})

		Convey("errors if conflict with existing column type", func() {
			err := db.Extend("note", skydb.RecordSchema{
				"content":   skydb.FieldType{Type: skydb.TypeString},
				"noteOrder": skydb.FieldType{Type: skydb.TypeNumber},
				"createdAt": skydb.FieldType{Type: skydb.TypeDateTime},
			})
			So(err, ShouldBeNil)

			err = db.Extend("note", skydb.RecordSchema{
				"content":   skydb.FieldType{Type: skydb.TypeNumber},
				"createdAt": skydb.FieldType{Type: skydb.TypeDateTime},
				"dirty":     skydb.FieldType{Type: skydb.TypeNumber},
			})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "conflicting schema {TypeString  {%!s(skydb.ExpressionType=0) <nil>}} => {TypeNumber  {%!s(skydb.ExpressionType=0) <nil>}}")
		})
	})
}

func TestGet(t *testing.T) {
	Convey("Database", t, func() {
		c := getTestConn(t)
		defer cleanupConn(t, c)

		db := c.PrivateDB("getuser")
		So(db.Extend("record", skydb.RecordSchema{
			"string":   skydb.FieldType{Type: skydb.TypeString},
			"number":   skydb.FieldType{Type: skydb.TypeNumber},
			"datetime": skydb.FieldType{Type: skydb.TypeDateTime},
			"boolean":  skydb.FieldType{Type: skydb.TypeBoolean},
		}), ShouldBeNil)

		insertRow(t, c.Db(), `INSERT INTO app_com_oursky_skygear."record" `+
			`(_database_id, _id, _owner_id, _created_at, _created_by, _updated_at, _updated_by, "string", "number", "datetime", "boolean") `+
			`VALUES ('getuser', 'id0', 'getuser', '1988-02-06', 'getuser', '1988-02-06', 'getuser', 'string', 1, '1988-02-06', TRUE)`)
		insertRow(t, c.Db(), `INSERT INTO app_com_oursky_skygear."record" `+
			`(_database_id, _id, _owner_id, _created_at, _created_by, _updated_at, _updated_by, "string", "number", "datetime", "boolean") `+
			`VALUES ('getuser', 'id1', 'getuser', '1988-02-06', 'getuser', '1988-02-06', 'getuser', 'string', 1, '1988-02-06', TRUE)`)

		Convey("gets an existing record from database", func() {
			record := skydb.Record{}
			err := db.Get(skydb.NewRecordID("record", "id1"), &record)
			So(err, ShouldBeNil)

			So(record.ID, ShouldResemble, skydb.NewRecordID("record", "id1"))
			So(record.DatabaseID, ShouldResemble, "getuser")
			So(record.OwnerID, ShouldResemble, "getuser")
			So(record.CreatorID, ShouldResemble, "getuser")
			So(record.UpdaterID, ShouldResemble, "getuser")
			So(record.Data["string"], ShouldEqual, "string")
			So(record.Data["number"], ShouldEqual, 1)
			So(record.Data["boolean"], ShouldEqual, true)

			So(record.CreatedAt, ShouldResemble, time.Date(1988, 2, 6, 0, 0, 0, 0, time.UTC))
			So(record.UpdatedAt, ShouldResemble, time.Date(1988, 2, 6, 0, 0, 0, 0, time.UTC))
			So(record.Data["datetime"].(time.Time), ShouldResemble, time.Date(1988, 2, 6, 0, 0, 0, 0, time.UTC))
		})

		Convey("errors if gets a non-existing record", func() {
			record := skydb.Record{}
			err := db.Get(skydb.NewRecordID("record", "notexistid"), &record)
			So(err, ShouldEqual, skydb.ErrRecordNotFound)
		})
	})
}

func TestSave(t *testing.T) {
	var c *conn
	Convey("Database", t, func() {
		c = getTestConn(t)
		defer cleanupConn(t, c)

		db := c.PublicDB()
		So(db.Extend("note", skydb.RecordSchema{
			"content":   skydb.FieldType{Type: skydb.TypeString},
			"number":    skydb.FieldType{Type: skydb.TypeNumber},
			"timestamp": skydb.FieldType{Type: skydb.TypeDateTime},
		}), ShouldBeNil)

		record := skydb.Record{
			ID:        skydb.NewRecordID("note", "someid"),
			OwnerID:   "user_id",
			CreatedAt: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
			CreatorID: "creator",
			UpdatedAt: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
			UpdaterID: "updater",
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
			err = c.QueryRowx(
				"SELECT content, number, timestamp, _owner_id "+
					"FROM app_com_oursky_skygear.note WHERE _id = 'someid' and _database_id = ''").
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
			err = c.QueryRowx("SELECT content FROM app_com_oursky_skygear.note WHERE _id = 'someid' and _database_id = ''").
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
			// FIXME: Wrap me with skydb.ErrXXX
			So(err, ShouldNotBeNil)
		})

		Convey("ignore Record.DatabaseID when saving", func() {
			record.DatabaseID = "someuserid"
			err := db.Save(&record)
			So(err, ShouldBeNil)
			So(record.DatabaseID, ShouldEqual, "")

			var count int
			err = c.QueryRowx("SELECT count(*) FROM app_com_oursky_skygear.note WHERE _id = 'someid' and _database_id = 'someuserid'").
				Scan(&count)
			So(err, ShouldBeNil)
			So(count, ShouldEqual, 0)
		})

		Convey("REGRESSION: update record with attribute having capital letters", func() {
			So(db.Extend("note", skydb.RecordSchema{
				"noteOrder": skydb.FieldType{Type: skydb.TypeNumber},
			}), ShouldBeNil)

			record = skydb.Record{
				ID:      skydb.NewRecordID("note", "1"),
				OwnerID: "user_id",
				Data: map[string]interface{}{
					"noteOrder": 1,
				},
			}

			ShouldBeNil(db.Save(&record))

			record.Data["noteOrder"] = 2
			ShouldBeNil(db.Save(&record))

			var noteOrder int
			err := c.QueryRowx(`SELECT "noteOrder" FROM app_com_oursky_skygear.note WHERE _id = '1' and _database_id = ''`).
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
			err = c.QueryRowx(`SELECT "_owner_id" FROM app_com_oursky_skygear.note WHERE _id = 'someid' and _database_id = ''`).
				Scan(&ownerID)
			So(ownerID, ShouldEqual, "user_id")
		})
	})
}

func TestACL(t *testing.T) {
	Convey("Database", t, func() {
		c := getTestConn(t)
		defer cleanupConn(t, c)

		db := c.PublicDB()
		So(db.Extend("note", nil), ShouldBeNil)

		record := skydb.Record{
			ID:      skydb.NewRecordID("note", "1"),
			OwnerID: "someuserid",
			ACL:     nil,
		}

		Convey("saves public access correctly", func() {
			err := db.Save(&record)

			So(err, ShouldBeNil)

			var b []byte
			err = c.QueryRowx(`SELECT _access FROM app_com_oursky_skygear.note WHERE _id = '1'`).
				Scan(&b)
			So(err, ShouldBeNil)
			So(b, ShouldResemble, []byte(nil))
		})
	})
}

func TestJSON(t *testing.T) {
	Convey("Database", t, func() {
		c := getTestConn(t)
		defer cleanupConn(t, c)

		db := c.PublicDB()
		So(db.Extend("note", skydb.RecordSchema{
			"jsonfield": skydb.FieldType{Type: skydb.TypeJSON},
		}), ShouldBeNil)

		Convey("fetch record with json field", func() {
			So(db.Extend("record", skydb.RecordSchema{
				"array":      skydb.FieldType{Type: skydb.TypeJSON},
				"dictionary": skydb.FieldType{Type: skydb.TypeJSON},
			}), ShouldBeNil)

			insertRow(t, c.Db(), `INSERT INTO app_com_oursky_skygear."record" `+
				`(_database_id, _id, _owner_id, _created_at, _created_by, _updated_at, _updated_by, "array", "dictionary") `+
				`VALUES ('', 'id', '', '0001-01-01 00:00:00', '', '0001-01-01 00:00:00', '', '[1, "string", true]', '{"number": 0, "string": "value", "bool": false}')`)

			var record skydb.Record
			err := db.Get(skydb.NewRecordID("record", "id"), &record)
			So(err, ShouldBeNil)

			So(record, ShouldResemble, skydb.Record{
				ID: skydb.NewRecordID("record", "id"),
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
			record := skydb.Record{
				ID:      skydb.NewRecordID("note", "1"),
				OwnerID: "user_id",
				Data: map[string]interface{}{
					"jsonfield": []interface{}{0.0, "string", true},
				},
			}

			So(db.Save(&record), ShouldBeNil)

			var jsonBytes []byte
			err := c.QueryRowx(`SELECT jsonfield FROM app_com_oursky_skygear.note WHERE _id = '1' and _database_id = ''`).
				Scan(&jsonBytes)
			So(err, ShouldBeNil)
			So(jsonBytes, ShouldEqualJSON, `[0, "string", true]`)
		})

		Convey("saves record field with dictionary", func() {
			record := skydb.Record{
				ID:      skydb.NewRecordID("note", "1"),
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
			err := c.QueryRowx(`SELECT jsonfield FROM app_com_oursky_skygear.note WHERE _id = '1' and _database_id = ''`).
				Scan(&jsonBytes)
			So(err, ShouldBeNil)
			So(jsonBytes, ShouldEqualJSON, `{"number": 1, "string": "", "bool": false}`)
		})
	})
}

func TestRecordAssetField(t *testing.T) {
	Convey("Record Asset", t, func() {
		c := getTestConn(t)
		defer cleanupConn(t, c)

		So(c.SaveAsset(&skydb.Asset{
			Name:        "picture.png",
			ContentType: "image/png",
			Size:        1,
		}), ShouldBeNil)

		db := c.PublicDB()
		So(db.Extend("note", skydb.RecordSchema{
			"image": skydb.FieldType{Type: skydb.TypeAsset},
		}), ShouldBeNil)

		Convey("can be associated", func() {
			err := db.Save(&skydb.Record{
				ID: skydb.NewRecordID("note", "id"),
				Data: map[string]interface{}{
					"image": &skydb.Asset{Name: "picture.png"},
				},
				OwnerID: "user_id",
			})
			So(err, ShouldBeNil)
		})

		Convey("errors when associated with non-existing asset", func() {
			err := db.Save(&skydb.Record{
				ID: skydb.NewRecordID("note", "id"),
				Data: map[string]interface{}{
					"image": &skydb.Asset{Name: "notexist.png"},
				},
				OwnerID: "user_id",
			})
			So(err, ShouldNotBeNil)
		})

		Convey("REGRESSION #229: can be fetched", func() {
			So(db.Save(&skydb.Record{
				ID: skydb.NewRecordID("note", "id"),
				Data: map[string]interface{}{
					"image": &skydb.Asset{Name: "picture.png"},
				},
				OwnerID: "user_id",
			}), ShouldBeNil)

			var record skydb.Record
			err := db.Get(skydb.NewRecordID("note", "id"), &record)
			So(err, ShouldBeNil)
			So(record, ShouldResemble, skydb.Record{
				ID: skydb.NewRecordID("note", "id"),
				Data: map[string]interface{}{
					"image": &skydb.Asset{Name: "picture.png"},
				},
				OwnerID: "user_id",
			})
		})
	})
}

func TestRecordLocationField(t *testing.T) {
	Convey("Database", t, func() {
		c := getTestConn(t)
		defer cleanupConn(t, c)

		db := c.PublicDB()
		So(db.Extend("photo", skydb.RecordSchema{
			"location": skydb.FieldType{Type: skydb.TypeLocation},
		}), ShouldBeNil)

		Convey("saves & load location field", func() {
			err := db.Save(&skydb.Record{
				ID: skydb.NewRecordID("photo", "1"),
				Data: map[string]interface{}{
					"location": skydb.NewLocation(1, 2),
				},
				OwnerID: "userid",
			})

			So(err, ShouldBeNil)

			record := skydb.Record{}
			err = db.Get(skydb.NewRecordID("photo", "1"), &record)
			So(err, ShouldBeNil)
			So(record, ShouldResemble, skydb.Record{
				ID: skydb.NewRecordID("photo", "1"),
				Data: map[string]interface{}{
					"location": skydb.NewLocation(1, 2),
				},
				OwnerID: "userid",
			})
		})
	})
}

func TestRecordSequenceField(t *testing.T) {
	Convey("Database", t, func() {
		c := getTestConn(t)
		defer cleanupConn(t, c)

		db := c.PublicDB()
		So(db.Extend("note", skydb.RecordSchema{
			"seq": skydb.FieldType{Type: skydb.TypeSequence},
		}), ShouldBeNil)

		Convey("saves & load sequence field", func() {
			record := skydb.Record{
				ID:      skydb.NewRecordID("note", "1"),
				OwnerID: "userid",
			}

			err := db.Save(&record)
			So(err, ShouldBeNil)
			So(record, ShouldResemble, skydb.Record{
				ID: skydb.NewRecordID("note", "1"),
				Data: map[string]interface{}{
					"seq": int64(1),
				},
				OwnerID: "userid",
			})

			record = skydb.Record{
				ID:      skydb.NewRecordID("note", "2"),
				OwnerID: "userid",
			}

			err = db.Save(&record)
			So(err, ShouldBeNil)
			So(record, ShouldResemble, skydb.Record{
				ID: skydb.NewRecordID("note", "2"),
				Data: map[string]interface{}{
					"seq": int64(2),
				},
				OwnerID: "userid",
			})
		})

		Convey("updates sequence field manually", func() {
			record := skydb.Record{
				ID:      skydb.NewRecordID("note", "1"),
				OwnerID: "userid",
			}

			So(db.Save(&record), ShouldBeNil)
			So(record.Data["seq"], ShouldEqual, 1)

			record.Data["seq"] = 10
			So(db.Save(&record), ShouldBeNil)

			So(record, ShouldResemble, skydb.Record{
				ID: skydb.NewRecordID("note", "1"),
				Data: map[string]interface{}{
					"seq": int64(10),
				},
				OwnerID: "userid",
			})

			// next record should's seq value should be 11
			record = skydb.Record{
				ID:      skydb.NewRecordID("note", "2"),
				OwnerID: "userid",
			}
			So(db.Save(&record), ShouldBeNil)
			So(record, ShouldResemble, skydb.Record{
				ID: skydb.NewRecordID("note", "2"),
				Data: map[string]interface{}{
					"seq": int64(11),
				},
				OwnerID: "userid",
			})
		})
	})
}

func TestDelete(t *testing.T) {
	var c *conn
	Convey("Database", t, func() {
		c = getTestConn(t)
		defer cleanupConn(t, c)

		db := c.PrivateDB("userid")

		So(db.Extend("note", skydb.RecordSchema{
			"content": skydb.FieldType{Type: skydb.TypeString},
		}), ShouldBeNil)

		record := skydb.Record{
			ID:      skydb.NewRecordID("note", "someid"),
			OwnerID: "user_id",
			Data: map[string]interface{}{
				"content": "some content",
			},
		}

		Convey("deletes existing record", func() {
			err := db.Save(&record)
			So(err, ShouldBeNil)

			err = db.Delete(skydb.NewRecordID("note", "someid"))
			So(err, ShouldBeNil)

			err = db.(*database).c.QueryRowx("SELECT * FROM app_com_oursky_skygear.note WHERE _id = 'someid' AND _database_id = 'userid'").Scan((*string)(nil))
			So(err, ShouldEqual, sql.ErrNoRows)
		})

		Convey("returns ErrRecordNotFound when record to delete doesn't exist", func() {
			err := db.Delete(skydb.NewRecordID("note", "notexistid"))
			So(err, ShouldEqual, skydb.ErrRecordNotFound)
		})

		Convey("return ErrRecordNotFound when deleting other user record", func() {
			err := db.Save(&record)
			So(err, ShouldBeNil)
			otherDB := c.PrivateDB("otheruserid")
			err = otherDB.Delete(skydb.NewRecordID("note", "someid"))
			So(err, ShouldEqual, skydb.ErrRecordNotFound)
		})
	})
}

func TestQuery(t *testing.T) {
	Convey("Database", t, func() {
		c := getTestConn(t)
		defer cleanupConn(t, c)

		// fixture
		record1 := skydb.Record{
			ID:      skydb.NewRecordID("note", "id1"),
			OwnerID: "user_id",
			Data: map[string]interface{}{
				"noteOrder": float64(1),
				"content":   "Hello World",
			},
		}
		record2 := skydb.Record{
			ID:      skydb.NewRecordID("note", "id2"),
			OwnerID: "user_id",
			Data: map[string]interface{}{
				"noteOrder": float64(2),
				"content":   "Bye World",
			},
		}
		record3 := skydb.Record{
			ID:      skydb.NewRecordID("note", "id3"),
			OwnerID: "user_id",
			Data: map[string]interface{}{
				"noteOrder": float64(3),
				"content":   "Good Hello",
			},
		}

		db := c.PrivateDB("userid")
		So(db.Extend("note", skydb.RecordSchema{
			"noteOrder": skydb.FieldType{Type: skydb.TypeNumber},
			"content":   skydb.FieldType{Type: skydb.TypeString},
		}), ShouldBeNil)

		err := db.Save(&record2)
		So(err, ShouldBeNil)
		err = db.Save(&record1)
		So(err, ShouldBeNil)
		err = db.Save(&record3)
		So(err, ShouldBeNil)

		Convey("queries records", func() {
			query := skydb.Query{
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
			query := skydb.Query{
				Type: "note",
				Sorts: []skydb.Sort{
					skydb.Sort{
						KeyPath: "noteOrder",
						Order:   skydb.Ascending,
					},
				},
			}
			records, err := exhaustRows(db.Query(&query))

			So(err, ShouldBeNil)
			So(records, ShouldResemble, []skydb.Record{
				record1,
				record2,
				record3,
			})
		})

		Convey("sorts queried records descendingly", func() {
			query := skydb.Query{
				Type: "note",
				Sorts: []skydb.Sort{
					skydb.Sort{
						KeyPath: "noteOrder",
						Order:   skydb.Descending,
					},
				},
			}
			records, err := exhaustRows(db.Query(&query))

			So(err, ShouldBeNil)
			So(records, ShouldResemble, []skydb.Record{
				record3,
				record2,
				record1,
			})
		})

		Convey("query records by note order", func() {
			query := skydb.Query{
				Type: "note",
				Predicate: skydb.Predicate{
					Operator: skydb.Equal,
					Children: []interface{}{
						skydb.Expression{
							Type:  skydb.KeyPath,
							Value: "noteOrder",
						},
						skydb.Expression{
							Type:  skydb.Literal,
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

		Convey("query records by content matching", func() {
			query := skydb.Query{
				Type: "note",
				Predicate: skydb.Predicate{
					Operator: skydb.Like,
					Children: []interface{}{
						skydb.Expression{
							Type:  skydb.KeyPath,
							Value: "content",
						},
						skydb.Expression{
							Type:  skydb.Literal,
							Value: "Hello%",
						},
					},
				},
			}
			records, err := exhaustRows(db.Query(&query))

			So(err, ShouldBeNil)
			So(records[0], ShouldResemble, record1)
			So(len(records), ShouldEqual, 1)
		})

		Convey("query records by case insensitive content matching", func() {
			query := skydb.Query{
				Type: "note",
				Predicate: skydb.Predicate{
					Operator: skydb.ILike,
					Children: []interface{}{
						skydb.Expression{
							Type:  skydb.KeyPath,
							Value: "content",
						},
						skydb.Expression{
							Type:  skydb.Literal,
							Value: "hello%",
						},
					},
				},
			}
			records, err := exhaustRows(db.Query(&query))

			So(err, ShouldBeNil)
			So(records[0], ShouldResemble, record1)
			So(len(records), ShouldEqual, 1)
		})

		Convey("query records by check array members", func() {
			query := skydb.Query{
				Type: "note",
				Predicate: skydb.Predicate{
					Operator: skydb.In,
					Children: []interface{}{
						skydb.Expression{
							Type:  skydb.KeyPath,
							Value: "content",
						},
						skydb.Expression{
							Type:  skydb.Literal,
							Value: []interface{}{"Bye World", "Good Hello", "Anything"},
						},
					},
				},
				Sorts: []skydb.Sort{
					skydb.Sort{
						KeyPath: "noteOrder",
						Order:   skydb.Descending,
					},
				},
			}
			records, err := exhaustRows(db.Query(&query))

			So(err, ShouldBeNil)
			So(records[0], ShouldResemble, record3)
			So(records[1], ShouldResemble, record2)
			So(len(records), ShouldEqual, 2)
		})

		Convey("query records by checking empty array", func() {
			query := skydb.Query{
				Type: "note",
				Predicate: skydb.Predicate{
					Operator: skydb.In,
					Children: []interface{}{
						skydb.Expression{
							Type:  skydb.KeyPath,
							Value: "content",
						},
						skydb.Expression{
							Type:  skydb.Literal,
							Value: []interface{}{},
						},
					},
				},
			}
			records, err := exhaustRows(db.Query(&query))

			So(err, ShouldBeNil)
			So(len(records), ShouldEqual, 0)
		})

		Convey("query records by note order using or predicate", func() {
			keyPathExpr := skydb.Expression{
				Type:  skydb.KeyPath,
				Value: "noteOrder",
			}
			value1 := skydb.Expression{
				Type:  skydb.Literal,
				Value: 2,
			}
			value2 := skydb.Expression{
				Type:  skydb.Literal,
				Value: 3,
			}
			query := skydb.Query{
				Type: "note",
				Predicate: skydb.Predicate{
					Operator: skydb.Or,
					Children: []interface{}{
						skydb.Predicate{
							Operator: skydb.Equal,
							Children: []interface{}{keyPathExpr, value1},
						},
						skydb.Predicate{
							Operator: skydb.Equal,
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

		Convey("query records by offset and paging", func() {
			query := skydb.Query{
				Type:   "note",
				Limit:  new(uint64),
				Offset: 1,
				Sorts: []skydb.Sort{
					skydb.Sort{
						KeyPath: "noteOrder",
						Order:   skydb.Descending,
					},
				},
			}
			*query.Limit = 2
			records, err := exhaustRows(db.Query(&query))

			So(err, ShouldBeNil)
			So(records[0], ShouldResemble, record2)
			So(records[1], ShouldResemble, record1)
			So(len(records), ShouldEqual, 2)
		})
	})

	Convey("Database with reference", t, func() {
		c := getTestConn(t)
		defer cleanupConn(t, c)

		// fixture
		record1 := skydb.Record{
			ID:      skydb.NewRecordID("note", "id1"),
			OwnerID: "user_id",
			Data: map[string]interface{}{
				"noteOrder": float64(1),
			},
		}
		record2 := skydb.Record{
			ID:      skydb.NewRecordID("note", "id2"),
			OwnerID: "user_id",
			Data: map[string]interface{}{
				"noteOrder": float64(2),
				"category":  skydb.NewReference("category", "important"),
			},
		}
		record3 := skydb.Record{
			ID:      skydb.NewRecordID("note", "id3"),
			OwnerID: "user_id",
			Data: map[string]interface{}{
				"noteOrder": float64(3),
				"category":  skydb.NewReference("category", "funny"),
			},
		}
		category1 := skydb.Record{
			ID:      skydb.NewRecordID("category", "important"),
			OwnerID: "user_id",
			Data:    map[string]interface{}{},
		}
		category2 := skydb.Record{
			ID:      skydb.NewRecordID("category", "funny"),
			OwnerID: "user_id",
			Data:    map[string]interface{}{},
		}

		db := c.PrivateDB("userid")
		So(db.Extend("category", skydb.RecordSchema{}), ShouldBeNil)
		So(db.Extend("note", skydb.RecordSchema{
			"noteOrder": skydb.FieldType{Type: skydb.TypeNumber},
			"category": skydb.FieldType{
				Type:          skydb.TypeReference,
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
			query := skydb.Query{
				Type: "note",
				Predicate: skydb.Predicate{
					Operator: skydb.Equal,
					Children: []interface{}{
						skydb.Expression{
							Type:  skydb.KeyPath,
							Value: "category",
						},
						skydb.Expression{
							Type:  skydb.Literal,
							Value: skydb.NewReference("category", "important"),
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

	Convey("Database with location", t, func() {
		c := getTestConn(t)
		defer cleanupConn(t, c)

		record0 := skydb.Record{
			ID:      skydb.NewRecordID("restaurant", "0"),
			OwnerID: "someuserid",
			Data: map[string]interface{}{
				"location": skydb.NewLocation(0, 0),
			},
		}
		record1 := skydb.Record{
			ID:      skydb.NewRecordID("restaurant", "1"),
			OwnerID: "someuserid",
			Data: map[string]interface{}{
				"location": skydb.NewLocation(1, 0),
			},
		}
		record2 := skydb.Record{
			ID:      skydb.NewRecordID("restaurant", "2"),
			OwnerID: "someuserid",
			Data: map[string]interface{}{
				"location": skydb.NewLocation(0, 1),
			},
		}

		db := c.PublicDB()
		So(db.Extend("restaurant", skydb.RecordSchema{
			"location": skydb.FieldType{Type: skydb.TypeLocation},
		}), ShouldBeNil)
		So(db.Save(&record0), ShouldBeNil)
		So(db.Save(&record1), ShouldBeNil)
		So(db.Save(&record2), ShouldBeNil)

		Convey("query within distance", func() {
			query := skydb.Query{
				Type: "restaurant",
				Predicate: skydb.Predicate{
					Operator: skydb.LessThanOrEqual,
					Children: []interface{}{
						skydb.Expression{
							Type: skydb.Function,
							Value: skydb.DistanceFunc{
								Field:    "location",
								Location: skydb.NewLocation(1, 1),
							},
						},
						skydb.Expression{
							Type:  skydb.Literal,
							Value: 157260,
						},
					},
				},
			}
			records, err := exhaustRows(db.Query(&query))

			So(err, ShouldBeNil)
			So(records, ShouldResemble, []skydb.Record{record0, record1, record2})
		})

		Convey("query within distance with func on R.H.S.", func() {
			query := skydb.Query{
				Type: "restaurant",
				Predicate: skydb.Predicate{
					Operator: skydb.GreaterThan,
					Children: []interface{}{
						skydb.Expression{
							Type:  skydb.Literal,
							Value: 157260,
						},
						skydb.Expression{
							Type: skydb.Function,
							Value: skydb.DistanceFunc{
								Field:    "location",
								Location: skydb.NewLocation(1, 1),
							},
						},
					},
				},
			}
			records, err := exhaustRows(db.Query(&query))

			So(err, ShouldBeNil)
			So(records, ShouldResemble, []skydb.Record{record0, record1, record2})
		})

		Convey("query with computed distance", func() {
			query := skydb.Query{
				Type: "restaurant",
				ComputedKeys: map[string]skydb.Expression{
					"distance": skydb.Expression{
						Type: skydb.Function,
						Value: skydb.DistanceFunc{
							Field:    "location",
							Location: skydb.NewLocation(1, 1),
						},
					},
				},
			}
			records, err := exhaustRows(db.Query(&query))

			So(err, ShouldBeNil)
			So(len(records), ShouldEqual, 3)
			So(records[0].Transient["distance"], ShouldAlmostEqual, 157249, 1)
		})

		Convey("query records ordered by distance", func() {
			query := skydb.Query{
				Type: "restaurant",
				Sorts: []skydb.Sort{
					{
						Func: skydb.DistanceFunc{
							Field:    "location",
							Location: skydb.NewLocation(0, 0),
						},
						Order: skydb.Desc,
					},
				},
			}

			records, err := exhaustRows(db.Query(&query))
			So(err, ShouldBeNil)
			So(records, ShouldResemble, []skydb.Record{record1, record2, record0})
		})
	})

	Convey("Database with multiple fields", t, func() {
		c := getTestConn(t)
		defer cleanupConn(t, c)

		record0 := skydb.Record{
			ID:      skydb.NewRecordID("restaurant", "0"),
			OwnerID: "someuserid",
			Data: map[string]interface{}{
				"cuisine": "american",
				"title":   "American Restaurant",
			},
		}
		record1 := skydb.Record{
			ID:      skydb.NewRecordID("restaurant", "1"),
			OwnerID: "someuserid",
			Data: map[string]interface{}{
				"cuisine": "chinese",
				"title":   "Chinese Restaurant",
			},
		}
		record2 := skydb.Record{
			ID:      skydb.NewRecordID("restaurant", "2"),
			OwnerID: "someuserid",
			Data: map[string]interface{}{
				"cuisine": "italian",
				"title":   "Italian Restaurant",
			},
		}

		recordsInDB := []skydb.Record{record0, record1, record2}

		db := c.PublicDB()
		So(db.Extend("restaurant", skydb.RecordSchema{
			"title":   skydb.FieldType{Type: skydb.TypeString},
			"cuisine": skydb.FieldType{Type: skydb.TypeString},
		}), ShouldBeNil)
		So(db.Save(&record0), ShouldBeNil)
		So(db.Save(&record1), ShouldBeNil)
		So(db.Save(&record2), ShouldBeNil)

		Convey("query with desired keys", func() {
			query := skydb.Query{
				Type:        "restaurant",
				DesiredKeys: []string{"cuisine"},
			}
			records, err := exhaustRows(db.Query(&query))

			So(err, ShouldBeNil)
			So(len(records), ShouldEqual, 3)
			for i, record := range records {
				So(record.Data["title"], ShouldBeNil)
				So(record.Data["cuisine"], ShouldEqual, recordsInDB[i].Data["cuisine"])
			}
		})

		Convey("query with empty desired keys", func() {
			query := skydb.Query{
				Type:        "restaurant",
				DesiredKeys: []string{},
			}
			records, err := exhaustRows(db.Query(&query))

			So(err, ShouldBeNil)
			So(len(records), ShouldEqual, 3)
			for _, record := range records {
				So(record.Data["title"], ShouldBeNil)
				So(record.Data["cuisine"], ShouldBeNil)
			}
		})

		Convey("query with nil desired keys", func() {
			query := skydb.Query{
				Type:        "restaurant",
				DesiredKeys: nil,
			}
			records, err := exhaustRows(db.Query(&query))

			So(err, ShouldBeNil)
			So(len(records), ShouldEqual, 3)
			for i, record := range records {
				So(record.Data["title"], ShouldEqual, recordsInDB[i].Data["title"])
				So(record.Data["cuisine"], ShouldEqual, recordsInDB[i].Data["cuisine"])
			}
		})

		Convey("query with non-recognized desired keys", func() {
			query := skydb.Query{
				Type:        "restaurant",
				DesiredKeys: []string{"pricing"},
			}
			_, err := exhaustRows(db.Query(&query))

			So(err, ShouldNotBeNil)
		})
	})

	Convey("Database with JSON", t, func() {
		c := getTestConn(t)
		defer cleanupConn(t, c)

		// fixture
		record1 := skydb.Record{
			ID:      skydb.NewRecordID("note", "id1"),
			OwnerID: "user_id",
			Data: map[string]interface{}{
				"primaryTag": "red",
				"tags":       []interface{}{"red", "green"},
			},
		}
		record2 := skydb.Record{
			ID:      skydb.NewRecordID("note", "id2"),
			OwnerID: "user_id",
			Data: map[string]interface{}{
				"primaryTag": "yellow",
				"tags":       []interface{}{"red", "green"},
			},
		}
		record3 := skydb.Record{
			ID:      skydb.NewRecordID("note", "id3"),
			OwnerID: "user_id",
			Data: map[string]interface{}{
				"primaryTag": "green",
				"tags":       []interface{}{"red", "yellow"},
			},
		}

		db := c.PrivateDB("userid")
		So(db.Extend("note", skydb.RecordSchema{
			"primaryTag": skydb.FieldType{Type: skydb.TypeString},
			"tags":       skydb.FieldType{Type: skydb.TypeJSON},
		}), ShouldBeNil)

		err := db.Save(&record2)
		So(err, ShouldBeNil)
		err = db.Save(&record1)
		So(err, ShouldBeNil)
		err = db.Save(&record3)
		So(err, ShouldBeNil)

		Convey("query records by literal string in JSON", func() {
			query := skydb.Query{
				Type: "note",
				Predicate: skydb.Predicate{
					Operator: skydb.In,
					Children: []interface{}{
						skydb.Expression{
							Type:  skydb.Literal,
							Value: "yellow",
						},
						skydb.Expression{
							Type:  skydb.KeyPath,
							Value: "tags",
						},
					},
				},
			}
			records, err := exhaustRows(db.Query(&query))

			So(err, ShouldBeNil)
			So(records, ShouldResemble, []skydb.Record{record3})
		})
	})

	Convey("Empty Conn", t, func() {
		c := getTestConn(t)
		defer cleanupConn(t, c)

		Convey("gets no users", func() {
			userinfo := skydb.UserInfo{}
			err := c.GetUser("notexistuserid", &userinfo)
			So(err, ShouldEqual, skydb.ErrUserNotFound)
		})

		Convey("gets no users with principal", func() {
			userinfo := skydb.UserInfo{}
			err := c.GetUserByPrincipalID("com.example:johndoe", &userinfo)
			So(err, ShouldEqual, skydb.ErrUserNotFound)
		})

		Convey("query no users", func() {
			emails := []string{"user@example.com"}
			result, err := c.QueryUser(emails)
			So(err, ShouldBeNil)
			So(len(result), ShouldEqual, 0)
		})

		Convey("updates no users", func() {
			userinfo := skydb.UserInfo{
				ID: "notexistuserid",
			}
			err := c.UpdateUser(&userinfo)
			So(err, ShouldEqual, skydb.ErrUserNotFound)
		})

		Convey("deletes no users", func() {
			err := c.DeleteUser("notexistuserid")
			So(err, ShouldEqual, skydb.ErrUserNotFound)
		})

		Convey("gets no devices", func() {
			device := skydb.Device{}
			err := c.GetDevice("notexistdeviceid", &device)
			So(err, ShouldEqual, skydb.ErrDeviceNotFound)
		})

		Convey("deletes no devices", func() {
			err := c.DeleteDevice("notexistdeviceid")
			So(err, ShouldEqual, skydb.ErrDeviceNotFound)
		})

		Convey("Empty Database", func() {
			db := c.PublicDB()

			Convey("gets nothing", func() {
				record := skydb.Record{}

				err := db.Get(skydb.NewRecordID("type", "notexistid"), &record)

				So(err, ShouldEqual, skydb.ErrRecordNotFound)
			})

			Convey("deletes nothing", func() {
				err := db.Delete(skydb.NewRecordID("type", "notexistid"))
				So(err, ShouldEqual, skydb.ErrRecordNotFound)
			})

			Convey("queries nothing", func() {
				query := skydb.Query{
					Type: "notexisttype",
				}

				records, err := exhaustRows(db.Query(&query))

				So(err, ShouldBeNil)
				So(records, ShouldBeEmpty)
			})
		})
	})
}

func TestQueryCount(t *testing.T) {
	Convey("Database", t, func() {
		c := getTestConn(t)
		defer cleanupConn(t, c)

		// fixture
		record1 := skydb.Record{
			ID:      skydb.NewRecordID("note", "id1"),
			OwnerID: "user_id",
			Data: map[string]interface{}{
				"noteOrder": float64(1),
				"content":   "Hello World",
			},
		}
		record2 := skydb.Record{
			ID:      skydb.NewRecordID("note", "id2"),
			OwnerID: "user_id",
			Data: map[string]interface{}{
				"noteOrder": float64(2),
				"content":   "Bye World",
			},
		}
		record3 := skydb.Record{
			ID:      skydb.NewRecordID("note", "id3"),
			OwnerID: "user_id",
			Data: map[string]interface{}{
				"noteOrder": float64(3),
				"content":   "Good Hello",
			},
		}

		db := c.PrivateDB("userid")
		So(db.Extend("note", skydb.RecordSchema{
			"noteOrder": skydb.FieldType{Type: skydb.TypeNumber},
			"content":   skydb.FieldType{Type: skydb.TypeString},
		}), ShouldBeNil)

		err := db.Save(&record2)
		So(err, ShouldBeNil)
		err = db.Save(&record1)
		So(err, ShouldBeNil)
		err = db.Save(&record3)
		So(err, ShouldBeNil)

		Convey("count records", func() {
			query := skydb.Query{
				Type: "note",
			}
			count, err := db.QueryCount(&query)

			So(err, ShouldBeNil)
			So(count, ShouldEqual, 3)
		})

		Convey("count records by content matching", func() {
			query := skydb.Query{
				Type: "note",
				Predicate: skydb.Predicate{
					Operator: skydb.Like,
					Children: []interface{}{
						skydb.Expression{
							Type:  skydb.KeyPath,
							Value: "content",
						},
						skydb.Expression{
							Type:  skydb.Literal,
							Value: "Hello%",
						},
					},
				},
			}
			count, err := db.QueryCount(&query)

			So(err, ShouldBeNil)
			So(count, ShouldEqual, 1)
		})

		Convey("count records by content with none matching", func() {
			query := skydb.Query{
				Type: "note",
				Predicate: skydb.Predicate{
					Operator: skydb.Like,
					Children: []interface{}{
						skydb.Expression{
							Type:  skydb.KeyPath,
							Value: "content",
						},
						skydb.Expression{
							Type:  skydb.Literal,
							Value: "Not Exist",
						},
					},
				},
			}
			count, err := db.QueryCount(&query)

			So(err, ShouldBeNil)
			So(count, ShouldEqual, 0)
		})
	})
}

func TestAggregateQuery(t *testing.T) {
	Convey("Database", t, func() {
		c := getTestConn(t)
		defer cleanupConn(t, c)

		// fixture
		db := c.PrivateDB("userid")
		So(db.Extend("note", skydb.RecordSchema{
			"category": skydb.FieldType{Type: skydb.TypeString},
		}), ShouldBeNil)

		categories := []string{"funny", "funny", "serious"}
		dbRecords := []skydb.Record{}

		for i, category := range categories {
			record := skydb.Record{
				ID:      skydb.NewRecordID("note", fmt.Sprintf("id%d", i)),
				OwnerID: "user_id",
				Data: map[string]interface{}{
					"category": category,
				},
			}
			err := db.Save(&record)
			dbRecords = append(dbRecords, record)
			So(err, ShouldBeNil)
		}

		equalCategoryPredicate := func(category string) skydb.Predicate {
			return skydb.Predicate{
				Operator: skydb.Equal,
				Children: []interface{}{
					skydb.Expression{
						Type:  skydb.KeyPath,
						Value: "category",
					},
					skydb.Expression{
						Type:  skydb.Literal,
						Value: category,
					},
				},
			}
		}

		Convey("queries records", func() {
			query := skydb.Query{
				Type:      "note",
				Predicate: equalCategoryPredicate("funny"),
				GetCount:  true,
			}
			rows, err := db.Query(&query)
			records, err := exhaustRows(rows, err)

			So(err, ShouldBeNil)
			So(len(records), ShouldEqual, 2)
			So(records[0], ShouldResemble, dbRecords[0])
			So(records[1], ShouldResemble, dbRecords[1])

			recordCount := rows.OverallRecordCount()
			So(recordCount, ShouldNotBeNil)
			So(*recordCount, ShouldEqual, 2)
		})

		Convey("queries no records", func() {
			query := skydb.Query{
				Type:      "note",
				Predicate: equalCategoryPredicate("interesting"),
				GetCount:  true,
			}
			rows, err := db.Query(&query)
			records, err := exhaustRows(rows, err)

			So(err, ShouldBeNil)
			So(len(records), ShouldEqual, 0)

			recordCount := rows.OverallRecordCount()
			So(recordCount, ShouldBeNil)
		})

		Convey("queries records with limit", func() {
			query := skydb.Query{
				Type:      "note",
				Predicate: equalCategoryPredicate("funny"),
				GetCount:  true,
				Limit:     new(uint64),
			}
			*query.Limit = 1
			rows, err := db.Query(&query)
			records, err := exhaustRows(rows, err)

			So(err, ShouldBeNil)
			So(records[0], ShouldResemble, dbRecords[0])
			So(len(records), ShouldEqual, 1)

			recordCount := rows.OverallRecordCount()
			So(recordCount, ShouldNotBeNil)
			So(*recordCount, ShouldEqual, 2)
		})
	})
}

func TestMetaDataQuery(t *testing.T) {
	Convey("Database", t, func() {
		c := getTestConn(t)
		defer cleanupConn(t, c)

		record0 := skydb.Record{
			ID:        skydb.NewRecordID("record", "0"),
			OwnerID:   "ownerID0",
			CreatedAt: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
			CreatorID: "creatorID0",
			UpdatedAt: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
			UpdaterID: "updaterID0",
			Data:      skydb.Data{},
		}
		record1 := skydb.Record{
			ID:        skydb.NewRecordID("record", "1"),
			OwnerID:   "ownerID1",
			CreatedAt: time.Date(2006, 1, 2, 15, 4, 6, 0, time.UTC),
			CreatorID: "creatorID1",
			UpdatedAt: time.Date(2006, 1, 2, 15, 4, 6, 0, time.UTC),
			UpdaterID: "updaterID1",
			Data:      skydb.Data{},
		}

		db := c.PublicDB()
		So(db.Extend("record", nil), ShouldBeNil)
		So(db.Save(&record0), ShouldBeNil)
		So(db.Save(&record1), ShouldBeNil)

		Convey("queries by record id", func() {
			query := skydb.Query{
				Type: "record",
				Predicate: skydb.Predicate{
					Operator: skydb.Equal,
					Children: []interface{}{
						skydb.Expression{
							Type:  skydb.KeyPath,
							Value: "_id",
						},
						skydb.Expression{
							Type:  skydb.Literal,
							Value: skydb.NewReference("record", "0"),
						},
					},
				},
			}
			records, err := exhaustRows(db.Query(&query))

			So(err, ShouldBeNil)
			So(records, ShouldResemble, []skydb.Record{record0})
		})

		Convey("queries by owner id", func() {
			query := skydb.Query{
				Type: "record",
				Predicate: skydb.Predicate{
					Operator: skydb.Equal,
					Children: []interface{}{
						skydb.Expression{
							Type:  skydb.KeyPath,
							Value: "_owner_id",
						},
						skydb.Expression{
							Type:  skydb.Literal,
							Value: skydb.NewReference("_user", "ownerID1"),
						},
					},
				},
			}
			records, err := exhaustRows(db.Query(&query))

			So(err, ShouldBeNil)
			So(records, ShouldResemble, []skydb.Record{record1})
		})

		Convey("queries by created at", func() {
			query := skydb.Query{
				Type: "record",
				Predicate: skydb.Predicate{
					Operator: skydb.LessThan,
					Children: []interface{}{
						skydb.Expression{
							Type:  skydb.KeyPath,
							Value: "_created_at",
						},
						skydb.Expression{
							Type:  skydb.Literal,
							Value: time.Date(2006, 1, 2, 15, 4, 6, 0, time.UTC),
						},
					},
				},
			}
			records, err := exhaustRows(db.Query(&query))

			So(err, ShouldBeNil)
			So(records, ShouldResemble, []skydb.Record{record0})
		})

		Convey("queries by created by", func() {
			query := skydb.Query{
				Type: "record",
				Predicate: skydb.Predicate{
					Operator: skydb.Equal,
					Children: []interface{}{
						skydb.Expression{
							Type:  skydb.KeyPath,
							Value: "_created_by",
						},
						skydb.Expression{
							Type:  skydb.Literal,
							Value: skydb.NewReference("_user", "creatorID0"),
						},
					},
				},
			}
			records, err := exhaustRows(db.Query(&query))

			So(err, ShouldBeNil)
			So(records, ShouldResemble, []skydb.Record{record0})
		})

		Convey("queries by updated at", func() {
			query := skydb.Query{
				Type: "record",
				Predicate: skydb.Predicate{
					Operator: skydb.GreaterThan,
					Children: []interface{}{
						skydb.Expression{
							Type:  skydb.KeyPath,
							Value: "_updated_at",
						},
						skydb.Expression{
							Type:  skydb.Literal,
							Value: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
						},
					},
				},
			}
			records, err := exhaustRows(db.Query(&query))

			So(err, ShouldBeNil)
			So(records, ShouldResemble, []skydb.Record{record1})
		})

		Convey("queries by updated by", func() {
			query := skydb.Query{
				Type: "record",
				Predicate: skydb.Predicate{
					Operator: skydb.Equal,
					Children: []interface{}{
						skydb.Expression{
							Type:  skydb.KeyPath,
							Value: "_updated_by",
						},
						skydb.Expression{
							Type:  skydb.Literal,
							Value: skydb.NewReference("_user", "updaterID1"),
						},
					},
				},
			}
			records, err := exhaustRows(db.Query(&query))

			So(err, ShouldBeNil)
			So(records, ShouldResemble, []skydb.Record{record1})
		})
	})
}

func TestUnsupportedQuery(t *testing.T) {
	Convey("Database", t, func() {
		c := getTestConn(t)
		defer cleanupConn(t, c)

		record0 := skydb.Record{
			ID:      skydb.NewRecordID("record", "0"),
			OwnerID: "ownerID0",
			Data:    skydb.Data{},
		}
		record1 := skydb.Record{
			ID:      skydb.NewRecordID("record", "1"),
			OwnerID: "ownerID1",
			Data:    skydb.Data{},
		}

		db := c.PublicDB()
		So(db.Extend("record", nil), ShouldBeNil)
		So(db.Save(&record0), ShouldBeNil)
		So(db.Save(&record1), ShouldBeNil)

		Convey("both side of IN is keypath", func() {
			query := skydb.Query{
				Type: "record",
				Predicate: skydb.Predicate{
					Operator: skydb.In,
					Children: []interface{}{
						skydb.Expression{
							Type:  skydb.KeyPath,
							Value: "categories",
						},
						skydb.Expression{
							Type:  skydb.KeyPath,
							Value: "favoriteCategory",
						},
					},
				},
			}
			So(func() { db.Query(&query) }, ShouldPanicWith, "malformed query")
		})
	})
}
