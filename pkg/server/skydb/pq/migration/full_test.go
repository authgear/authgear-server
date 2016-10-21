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

package migration

import (
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	. "github.com/smartystreets/goconvey/convey"
)

func TestFullMigration(t *testing.T) {
	schema := testSchemaName()

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
