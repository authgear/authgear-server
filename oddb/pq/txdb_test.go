package pq

import (
	"database/sql"
	"testing"

	"github.com/oursky/ourd/oddb"
	. "github.com/smartystreets/goconvey/convey"
)

func TestTxDB(t *testing.T) {
	Convey("TxDatabase", t, func() {
		c := getTestConn(t)
		defer cleanupDB(t, c.Db)

		dbx := c.Db
		db := c.PublicDB().(*database)
		So(db.Extend("record", oddb.RecordSchema{
			"content": oddb.FieldType{Type: oddb.TypeString},
		}), ShouldBeNil)

		insertRow(t, dbx, `INSERT INTO app_com_oursky_ourd."record" `+
			`(_database_id, _id, _owner_id, _created_at, _created_by, _updated_at, _updated_by, "content") `+
			`VALUES ('', '1', '', '0001-01-01 00:00:00', '', '0001-01-01 00:00:00', '', 'original1')`)
		insertRow(t, dbx, `INSERT INTO app_com_oursky_ourd."record" `+
			`(_database_id, _id, _owner_id, _created_at, _created_by, _updated_at, _updated_by, "content") `+
			`VALUES ('', '2', '', '0001-01-01 00:00:00', '', '0001-01-01 00:00:00', '', 'original2')`)
		insertRow(t, dbx, `INSERT INTO app_com_oursky_ourd."record" `+
			`(_database_id, _id, _owner_id, _created_at, _created_by, _updated_at, _updated_by, "content") `+
			`VALUES ('', '3', '', '0001-01-01 00:00:00', '', '0001-01-01 00:00:00', '', 'original3')`)

		Convey("with modification after Begin", func() {
			err := db.Begin()
			So(err, ShouldBeNil)

			// create
			So(db.Save(&oddb.Record{
				ID:      oddb.NewRecordID("record", "0"),
				Data:    map[string]interface{}{"content": "new0"},
				OwnerID: "ownerID",
			}), ShouldBeNil)

			// update
			So(db.Save(&oddb.Record{
				ID:      oddb.NewRecordID("record", "1"),
				Data:    map[string]interface{}{"content": "new1"},
				OwnerID: "ownerID",
			}), ShouldBeNil)

			// delete
			So(db.Delete(oddb.NewRecordID("record", "2")), ShouldBeNil)

			Convey("Commit saves all the changes", func() {
				err = db.Commit()
				So(err, ShouldBeNil)

				var content string
				err = dbx.QueryRowx(`SELECT content FROM app_com_oursky_ourd."record" WHERE _id = '0'`).
					Scan(&content)
				So(err, ShouldBeNil)
				So(content, ShouldEqual, "new0")

				err = dbx.QueryRowx(`SELECT content FROM app_com_oursky_ourd."record" WHERE _id = '1'`).
					Scan(&content)
				So(err, ShouldBeNil)
				So(content, ShouldEqual, "new1")

				err = dbx.QueryRowx(`SELECT content FROM app_com_oursky_ourd."record" WHERE _id = '2'`).
					Scan(&content)
				So(err, ShouldEqual, sql.ErrNoRows)
			})

			Convey("Rollback undo all the changes", func() {
				err = db.Rollback()
				So(err, ShouldBeNil)

				var content string
				err = dbx.QueryRowx(`SELECT content FROM app_com_oursky_ourd."record" WHERE _id = '0'`).
					Scan(&content)
				So(err, ShouldEqual, sql.ErrNoRows)

				err = dbx.QueryRowx(`SELECT content FROM app_com_oursky_ourd."record" WHERE _id = '1'`).
					Scan(&content)
				So(err, ShouldBeNil)
				So(content, ShouldEqual, "original1")

				err = dbx.QueryRowx(`SELECT content FROM app_com_oursky_ourd."record" WHERE _id = '2'`).
					Scan(&content)
				So(err, ShouldBeNil)
				So(content, ShouldEqual, "original2")
			})
		})

		Convey("Begin on a Begin'ed db returns ErrDatabaseTxDidBegin", func() {
			So(db.Begin(), ShouldBeNil)
			err := db.Begin()
			So(err, ShouldEqual, oddb.ErrDatabaseTxDidBegin)
		})

		Convey("Commit/Rollback on a non-Begin'ed db returns ErrDatabaseTxDidNotBegin", func() {
			So(db.Commit(), ShouldEqual, oddb.ErrDatabaseTxDidNotBegin)
			So(db.Rollback(), ShouldEqual, oddb.ErrDatabaseTxDidNotBegin)
		})

		Convey("Begin/Commit/Rollback on a Commit'ed db returns ErrDatabaseTxDone", func() {
			So(db.Begin(), ShouldBeNil)
			So(db.Commit(), ShouldBeNil)

			So(db.Begin(), ShouldEqual, oddb.ErrDatabaseTxDone)
			So(db.Commit(), ShouldEqual, oddb.ErrDatabaseTxDone)
			So(db.Rollback(), ShouldEqual, oddb.ErrDatabaseTxDone)
		})

		Convey("Begin/Commit/Rollback on a Rollback'ed db returns ErrDatabaseTxDone", func() {
			So(db.Begin(), ShouldBeNil)
			So(db.Rollback(), ShouldBeNil)

			So(db.Begin(), ShouldEqual, oddb.ErrDatabaseTxDone)
			So(db.Commit(), ShouldEqual, oddb.ErrDatabaseTxDone)
			So(db.Rollback(), ShouldEqual, oddb.ErrDatabaseTxDone)
		})
	})
}
