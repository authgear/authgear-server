package migration

import (
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	. "github.com/smartystreets/goconvey/convey"
)

func TestFullMigration(t *testing.T) {
	schema := "app_com_oursky_skygear"

	Convey("execute", t, func() {
		db := getTestDB(t)
		defer cleanupDB(t, db, schema)

		Convey("execute full migration", func() {
			full := &fullMigration{}
			executeInTransaction(t, db, func(tx *sqlx.Tx) {
				ensureSchema(tx, schema)
				So(full.Up(tx), ShouldBeNil)
				So(tx.Commit(), ShouldBeNil)
			})

			executeInTransaction(t, db, func(tx *sqlx.Tx) {
				exists, err := tableExists(tx, schema, "_user")
				So(exists, ShouldBeTrue)
				So(err, ShouldBeNil)
			})
		})
	})
}
