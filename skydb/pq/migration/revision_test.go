package migration

import (
	"fmt"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	. "github.com/smartystreets/goconvey/convey"
)

func createTestRecordTable(tx *sqlx.Tx, tableName string) error {
	// Note: This record table schema represents the initial
	// version. There is no need to update this schema.
	_, err := tx.Exec(fmt.Sprintf(`CREATE TABLE %s (
	    _id text,
	    _database_id text,
	    _owner_id text,
	    _access jsonb,
	    PRIMARY KEY(_id, _database_id, _owner_id),
	    UNIQUE (_id)
	);`, tableName))
	return err
}

func TestRevisions(t *testing.T) {
	schema := "app_com_oursky_skygear"

	Convey("execute", t, func() {
		db := getTestDB(t)
		defer cleanupDB(t, db, schema)

		Convey("execute each revision", func() {
			executeInTransaction(t, db, func(tx *sqlx.Tx) {
				ensureSchema(tx, schema)
				for i := range revisions {
					So(revisions[i].Up(tx), ShouldBeNil)

					if i == 0 {
						err := createTestRecordTable(tx, "test")
						if err != nil {
							t.Fatal(err)
						}
					}
				}
				So(tx.Commit(), ShouldBeNil)
			})

			executeInTransaction(t, db, func(tx *sqlx.Tx) {
				exists, err := tableExists(tx, schema, "_user")
				So(exists, ShouldBeTrue)
				So(err, ShouldBeNil)
			})
		})
	})

	Convey("version num", t, func() {
		Convey("last revision version same as expected", func() {
			full := &fullMigration{}
			So(revisions[len(revisions)-1].Version(), ShouldEqual, full.Version())
		})
	})
}
