// Copyright 2015-present Oursky Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pq

import (
	"database/sql"
	"testing"

	"github.com/skygeario/skygear-server/pkg/server/skydb"
	. "github.com/smartystreets/goconvey/convey"
)

func TestTxDB(t *testing.T) {
	Convey("TxDatabase", t, func() {
		c := getTestConn(t)
		defer cleanupConn(t, c)

		dbx := c.Db()
		db := c.PublicDB().(*database)
		So(db.Extend("record", skydb.RecordSchema{
			"content": skydb.FieldType{Type: skydb.TypeString},
		}), ShouldBeNil)

		insertRow(t, dbx, `INSERT INTO "record" `+
			`(_database_id, _id, _owner_id, _created_at, _created_by, _updated_at, _updated_by, "content") `+
			`VALUES ('', '1', '', '0001-01-01 00:00:00', '', '0001-01-01 00:00:00', '', 'original1')`)
		insertRow(t, dbx, `INSERT INTO "record" `+
			`(_database_id, _id, _owner_id, _created_at, _created_by, _updated_at, _updated_by, "content") `+
			`VALUES ('', '2', '', '0001-01-01 00:00:00', '', '0001-01-01 00:00:00', '', 'original2')`)
		insertRow(t, dbx, `INSERT INTO "record" `+
			`(_database_id, _id, _owner_id, _created_at, _created_by, _updated_at, _updated_by, "content") `+
			`VALUES ('', '3', '', '0001-01-01 00:00:00', '', '0001-01-01 00:00:00', '', 'original3')`)

		Convey("with modification after Begin", func() {
			err := db.Begin()
			So(err, ShouldBeNil)

			// create
			So(db.Save(&skydb.Record{
				ID:      skydb.NewRecordID("record", "0"),
				Data:    map[string]interface{}{"content": "new0"},
				OwnerID: "ownerID",
			}), ShouldBeNil)

			// update
			So(db.Save(&skydb.Record{
				ID:      skydb.NewRecordID("record", "1"),
				Data:    map[string]interface{}{"content": "new1"},
				OwnerID: "ownerID",
			}), ShouldBeNil)

			// delete
			So(db.Delete(skydb.NewRecordID("record", "2")), ShouldBeNil)

			Convey("Commit saves all the changes", func() {
				err = db.Commit()
				So(err, ShouldBeNil)

				var content string
				err = dbx.QueryRowx(`SELECT content FROM "record" WHERE _id = '0'`).
					Scan(&content)
				So(err, ShouldBeNil)
				So(content, ShouldEqual, "new0")

				err = dbx.QueryRowx(`SELECT content FROM "record" WHERE _id = '1'`).
					Scan(&content)
				So(err, ShouldBeNil)
				So(content, ShouldEqual, "new1")

				err = dbx.QueryRowx(`SELECT content FROM "record" WHERE _id = '2'`).
					Scan(&content)
				So(err, ShouldEqual, sql.ErrNoRows)
			})

			Convey("Rollback undo all the changes", func() {
				err = db.Rollback()
				So(err, ShouldBeNil)

				var content string
				err = dbx.QueryRowx(`SELECT content FROM "record" WHERE _id = '0'`).
					Scan(&content)
				So(err, ShouldEqual, sql.ErrNoRows)

				err = dbx.QueryRowx(`SELECT content FROM "record" WHERE _id = '1'`).
					Scan(&content)
				So(err, ShouldBeNil)
				So(content, ShouldEqual, "original1")

				err = dbx.QueryRowx(`SELECT content FROM "record" WHERE _id = '2'`).
					Scan(&content)
				So(err, ShouldBeNil)
				So(content, ShouldEqual, "original2")
			})
		})

		Convey("Begin on a Begin'ed db returns ErrDatabaseTxDidBegin", func() {
			So(db.Begin(), ShouldBeNil)
			err := db.Begin()
			So(err, ShouldEqual, skydb.ErrDatabaseTxDidBegin)
		})

		Convey("Commit/Rollback on a non-Begin'ed db returns ErrDatabaseTxDidNotBegin", func() {
			So(db.Commit(), ShouldEqual, skydb.ErrDatabaseTxDidNotBegin)
			So(db.Rollback(), ShouldEqual, skydb.ErrDatabaseTxDidNotBegin)
		})

		Convey("New transaction can begin after commit", func() {
			So(db.Begin(), ShouldBeNil)
			So(db.Commit(), ShouldBeNil)

			So(db.Begin(), ShouldEqual, nil)
		})

		Convey("New transaction can begin after rollback", func() {
			So(db.Begin(), ShouldBeNil)
			So(db.Rollback(), ShouldBeNil)

			So(db.Begin(), ShouldEqual, nil)
		})
	})
}
