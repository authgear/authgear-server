package pq

import (
	"testing"

	"github.com/oursky/skygear/skydb"
	. "github.com/smartystreets/goconvey/convey"
)

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
			So(err.Error(), ShouldEqual, "conflicting schema {TypeString  {0 <nil>}} => {TypeNumber  {0 <nil>}}")
		})
	})

	Convey("RenameSchema", t, func() {
		c := getTestConn(t)
		defer cleanupConn(t, c)

		db := c.PublicDB()

		Convey("rename column normally", func() {
			err := db.Extend("note", skydb.RecordSchema{
				"content":   skydb.FieldType{Type: skydb.TypeString},
				"noteOrder": skydb.FieldType{Type: skydb.TypeNumber},
				"createdAt": skydb.FieldType{Type: skydb.TypeDateTime},
			})
			So(err, ShouldBeNil)

			err = db.RenameSchema("note", "content", "content2")
			So(err, ShouldBeNil)

			// verify with an insert
			_, err = c.Exec(
				`INSERT INTO app_com_oursky_skygear."note" ` +
					`(_id, _database_id, _owner_id, _created_at, _created_by, _updated_at, _updated_by, "content", "noteOrder", "createdAt") ` +
					`VALUES (1, 1, 1, '1988-02-06', 'creator', '1988-02-06', 'updater', 'some content', 2, '1988-02-06')`)
			So(err, ShouldNotBeNil)

			result, err := c.Exec(
				`INSERT INTO app_com_oursky_skygear."note" ` +
					`(_id, _database_id, _owner_id, _created_at, _created_by, _updated_at, _updated_by, "content2", "noteOrder", "createdAt") ` +
					`VALUES (1, 1, 1, '1988-02-06', 'creator', '1988-02-06', 'updater', 'some content', 2, '1988-02-06')`)
			So(err, ShouldBeNil)

			i, err := result.RowsAffected()
			So(err, ShouldBeNil)
			So(i, ShouldEqual, 1)
		})

		Convey("rename column with reserved name", func() {
			err := db.Extend("note", skydb.RecordSchema{
				"some":      skydb.FieldType{Type: skydb.TypeString},
				"noteOrder": skydb.FieldType{Type: skydb.TypeNumber},
				"createdAt": skydb.FieldType{Type: skydb.TypeDateTime},
			})
			So(err, ShouldBeNil)

			// "some" is reserved by psql
			err = db.RenameSchema("note", "some", "content")
			So(err, ShouldBeNil)
		})

		Convey("rename unexisting column", func() {
			err := db.Extend("note", skydb.RecordSchema{
				"content":   skydb.FieldType{Type: skydb.TypeString},
				"noteOrder": skydb.FieldType{Type: skydb.TypeNumber},
				"createdAt": skydb.FieldType{Type: skydb.TypeDateTime},
			})
			So(err, ShouldBeNil)

			err = db.RenameSchema("note", "notExist", "content2")
			So(err, ShouldNotBeNil)

			// schema should remain unchanged
			result, err := c.Exec(
				`INSERT INTO app_com_oursky_skygear."note" ` +
					`(_id, _database_id, _owner_id, _created_at, _created_by, _updated_at, _updated_by, "content", "noteOrder", "createdAt") ` +
					`VALUES (1, 1, 1, '1988-02-06', 'creator', '1988-02-06', 'updater', 'some content', 2, '1988-02-06')`)
			So(err, ShouldBeNil)

			i, err := result.RowsAffected()
			So(err, ShouldBeNil)
			So(i, ShouldEqual, 1)
		})

		Convey("rename to an existing column", func() {
			err := db.Extend("note", skydb.RecordSchema{
				"content":   skydb.FieldType{Type: skydb.TypeString},
				"content2":  skydb.FieldType{Type: skydb.TypeString},
				"noteOrder": skydb.FieldType{Type: skydb.TypeNumber},
				"createdAt": skydb.FieldType{Type: skydb.TypeDateTime},
			})
			So(err, ShouldBeNil)

			err = db.RenameSchema("note", "content", "content2")
			So(err, ShouldNotBeNil)

			// schema should remain unchanged
			result, err := c.Exec(
				`INSERT INTO app_com_oursky_skygear."note" ` +
					`(_id, _database_id, _owner_id, _created_at, _created_by, _updated_at, _updated_by, "content", "noteOrder", "createdAt") ` +
					`VALUES (1, 1, 1, '1988-02-06', 'creator', '1988-02-06', 'updater', 'some content', 2, '1988-02-06')`)
			So(err, ShouldBeNil)

			i, err := result.RowsAffected()
			So(err, ShouldBeNil)
			So(i, ShouldEqual, 1)
		})

		Convey("rename unexisting table", func() {
			err := db.RenameSchema("notExist", "content", "content2")
			So(err, ShouldNotBeNil)
		})
	})

	Convey("DeleteSchema", t, func() {
		c := getTestConn(t)
		defer cleanupConn(t, c)

		db := c.PublicDB()

		Convey("delete column normally", func() {
			err := db.Extend("note", skydb.RecordSchema{
				"content":   skydb.FieldType{Type: skydb.TypeString},
				"noteOrder": skydb.FieldType{Type: skydb.TypeNumber},
				"createdAt": skydb.FieldType{Type: skydb.TypeDateTime},
			})
			So(err, ShouldBeNil)

			err = db.DeleteSchema("note", "content")
			So(err, ShouldBeNil)

			// verify with an insert
			_, err = c.Exec(
				`INSERT INTO app_com_oursky_skygear."note" ` +
					`(_id, _database_id, _owner_id, _created_at, _created_by, _updated_at, _updated_by, "content", "noteOrder", "createdAt") ` +
					`VALUES (1, 1, 1, '1988-02-06', 'creator', '1988-02-06', 'updater', 'some content', 2, '1988-02-06')`)
			So(err, ShouldNotBeNil)

			result, err := c.Exec(
				`INSERT INTO app_com_oursky_skygear."note" ` +
					`(_id, _database_id, _owner_id, _created_at, _created_by, _updated_at, _updated_by, "noteOrder", "createdAt") ` +
					`VALUES (1, 1, 1, '1988-02-06', 'creator', '1988-02-06', 'updater', 2, '1988-02-06')`)
			So(err, ShouldBeNil)

			i, err := result.RowsAffected()
			So(err, ShouldBeNil)
			So(i, ShouldEqual, 1)
		})

		Convey("delete column with reserved name", func() {
			err := db.Extend("note", skydb.RecordSchema{
				"some":      skydb.FieldType{Type: skydb.TypeString},
				"noteOrder": skydb.FieldType{Type: skydb.TypeNumber},
				"createdAt": skydb.FieldType{Type: skydb.TypeDateTime},
			})
			So(err, ShouldBeNil)

			// "some" is reserved by psql
			err = db.DeleteSchema("note", "some")
			So(err, ShouldBeNil)
		})

		Convey("delete unexisting column", func() {
			err := db.Extend("note", skydb.RecordSchema{
				"content":   skydb.FieldType{Type: skydb.TypeString},
				"noteOrder": skydb.FieldType{Type: skydb.TypeNumber},
				"createdAt": skydb.FieldType{Type: skydb.TypeDateTime},
			})
			So(err, ShouldBeNil)

			err = db.DeleteSchema("note", "notExist")
			So(err, ShouldNotBeNil)

			// schema should remain unchanged
			result, err := c.Exec(
				`INSERT INTO app_com_oursky_skygear."note" ` +
					`(_id, _database_id, _owner_id, _created_at, _created_by, _updated_at, _updated_by, "content", "noteOrder", "createdAt") ` +
					`VALUES (1, 1, 1, '1988-02-06', 'creator', '1988-02-06', 'updater', 'some content', 2, '1988-02-06')`)
			So(err, ShouldBeNil)

			i, err := result.RowsAffected()
			So(err, ShouldBeNil)
			So(i, ShouldEqual, 1)
		})

		Convey("delete unexisting table", func() {
			err := db.DeleteSchema("notExist", "content")
			So(err, ShouldNotBeNil)
		})
	})
}
