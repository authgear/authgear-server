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
	"fmt"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	. "github.com/smartystreets/goconvey/convey"
)

func testSchemaName() string {
	return "app_io_skygear_test"
}

func getTestDB(t *testing.T) *sqlx.DB {
	defaultTo := func(envvar string, value string) {
		if os.Getenv(envvar) == "" {
			os.Setenv(envvar, value)
		}
	}
	defaultTo("PGDATABASE", "skygear_test")
	defaultTo("PGSSLMODE", "disable")

	db, err := sqlx.Open("postgres", "")
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func cleanupDB(t *testing.T, db *sqlx.DB, schema string) {
	_, err := db.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schema))
	if err != nil {
		t.Fatal(err)
	}
}

func executeInTransaction(t *testing.T, db *sqlx.DB, f func(*sqlx.Tx)) {
	tx, err := db.Beginx()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()

	f(tx)
}

type naiveRevision struct {
	VersionNum string
	UpFunc     func(tx *sqlx.Tx) error
	DownFunc   func(tx *sqlx.Tx) error
}

func (r *naiveRevision) Version() string        { return r.VersionNum }
func (r *naiveRevision) Up(tx *sqlx.Tx) error   { return r.UpFunc(tx) }
func (r *naiveRevision) Down(tx *sqlx.Tx) error { return r.DownFunc(tx) }

func TestSchemaAndVersion(t *testing.T) {
	schema := testSchemaName()

	Convey("Schema", t, func() {
		db := getTestDB(t)
		defer cleanupDB(t, db, schema)

		Convey("ensure schema is created and selected", func() {
			executeInTransaction(t, db, func(tx *sqlx.Tx) {
				So(ensureSchema(tx, schema), ShouldBeNil)
				So(tx.Commit(), ShouldBeNil)
			})

			var schemaExists bool
			db.QueryRowx(`SELECT EXISTS(SELECT 1 FROM pg_catalog.pg_namespace WHERE nspname = $1);`, schema).Scan(&schemaExists)
			So(schemaExists, ShouldBeTrue)

			var currentSchema string
			db.QueryRowx(`SELECT current_schema();`).Scan(&currentSchema)
			So(currentSchema, ShouldEqual, schema)
		})
	})

	Convey("Version Table", t, func() {
		db := getTestDB(t)
		defer cleanupDB(t, db, schema)

		executeInTransaction(t, db, func(tx *sqlx.Tx) {
			if err := ensureSchema(tx, schema); err != nil {
				t.Fatal(err)
			}
			tx.Commit()
		})

		Convey("version table not exists", func() {
			executeInTransaction(t, db, func(tx *sqlx.Tx) {
				exists, err := versionTableExists(tx, schema)
				So(exists, ShouldBeFalse)
				So(err, ShouldBeNil)
			})
		})

		Convey("version table exists", func() {
			_, err := db.Exec(fmt.Sprintf(`CREATE TABLE %s._version();`, schema))
			if err != nil {
				t.Fatal(err)
			}
			executeInTransaction(t, db, func(tx *sqlx.Tx) {
				exists, err := versionTableExists(tx, schema)
				So(exists, ShouldBeTrue)
				So(err, ShouldBeNil)
			})
		})

		Convey("ensure version table", func() {
			executeInTransaction(t, db, func(tx *sqlx.Tx) {
				So(ensureVersionTable(tx, schema), ShouldBeNil)
				tx.Commit()
			})

			executeInTransaction(t, db, func(tx *sqlx.Tx) {
				exists, err := versionTableExists(tx, schema)
				So(exists, ShouldBeTrue)
				So(err, ShouldBeNil)
			})
		})

		Convey("current version number without table", func() {
			executeInTransaction(t, db, func(tx *sqlx.Tx) {
				versionNum, err := currentVersionNum(tx, schema)
				So(versionNum, ShouldEqual, "")
				So(err, ShouldBeNil)
			})
		})

		Convey("current version number with table and empty row", func() {
			executeInTransaction(t, db, func(tx *sqlx.Tx) {
				ensureVersionTable(tx, schema)
				tx.Commit()
			})

			executeInTransaction(t, db, func(tx *sqlx.Tx) {
				versionNum, err := currentVersionNum(tx, schema)
				So(versionNum, ShouldEqual, "")
				So(err, ShouldBeNil)
			})
		})

		Convey("current version number with table and a row", func() {
			executeInTransaction(t, db, func(tx *sqlx.Tx) {
				ensureVersionTable(tx, schema)
				tx.Exec(fmt.Sprintf(`INSERT INTO %s._version (version_num) VALUES ($1);`, schema), "version-num")
				tx.Commit()
			})

			executeInTransaction(t, db, func(tx *sqlx.Tx) {
				versionNum, err := currentVersionNum(tx, schema)
				So(versionNum, ShouldEqual, "version-num")
				So(err, ShouldBeNil)
			})
		})

		Convey("set version num from empty", func() {
			executeInTransaction(t, db, func(tx *sqlx.Tx) {
				ensureVersionTable(tx, schema)
				So(setVersionNum(tx, "", "new-version"), ShouldBeNil)
				tx.Commit()
			})

			executeInTransaction(t, db, func(tx *sqlx.Tx) {
				versionNum, _ := currentVersionNum(tx, schema)
				So(versionNum, ShouldEqual, "new-version")
			})
		})

		Convey("set version num from old version", func() {
			executeInTransaction(t, db, func(tx *sqlx.Tx) {
				ensureVersionTable(tx, schema)
				tx.Exec(fmt.Sprintf(`INSERT INTO %s._version (version_num) VALUES ($1);`, schema), "old-version")
				So(setVersionNum(tx, "old-version", "new-version"), ShouldBeNil)
				tx.Commit()
			})

			executeInTransaction(t, db, func(tx *sqlx.Tx) {
				versionNum, _ := currentVersionNum(tx, schema)
				So(versionNum, ShouldEqual, "new-version")
			})
		})
	})
}

func TestMigration(t *testing.T) {
	schema := testSchemaName()

	Convey("Execute", t, func() {
		db := getTestDB(t)
		defer cleanupDB(t, db, schema)

		originalFindRevisions := findRevisions
		defer func() {
			findRevisions = originalFindRevisions
		}()

		firstTableRevision := &naiveRevision{
			VersionNum: "version1",
			UpFunc: func(tx *sqlx.Tx) error {
				_, err := tx.Exec(`CREATE TABLE _first();`)
				return err
			},
			DownFunc: func(tx *sqlx.Tx) error {
				_, err := tx.Exec(`DROP TABLE _first;`)
				return err
			},
		}

		secondTableRevision := &naiveRevision{
			VersionNum: "version2",
			UpFunc: func(tx *sqlx.Tx) error {
				_, err := tx.Exec(`CREATE TABLE _second();`)
				return err
			},
			DownFunc: func(tx *sqlx.Tx) error {
				_, err := tx.Exec(`DROP TABLE _second;`)
				return err
			},
		}

		Convey("execute upgrade", func() {
			findRevisions = func(original string, target string) []Revision {
				if original == "version2" {
					So(original, ShouldEqual, "version2")
					So(target, ShouldEqual, "version1")
					return []Revision{secondTableRevision}
				}
				So(original, ShouldEqual, "")
				So(target, ShouldEqual, "version2")
				return []Revision{firstTableRevision, secondTableRevision}
			}

			// upgrade schema
			executeInTransaction(t, db, func(tx *sqlx.Tx) {
				So(executeSchemaMigrations(tx, schema, "", "version2", false), ShouldBeNil)
				So(tx.Commit(), ShouldBeNil)
			})

			executeInTransaction(t, db, func(tx *sqlx.Tx) {
				exists, err := tableExists(tx, schema, "_first")
				So(exists, ShouldBeTrue)
				So(err, ShouldBeNil)
				exists, err = tableExists(tx, schema, "_second")
				So(exists, ShouldBeTrue)
				So(err, ShouldBeNil)
				versionNum, err := currentVersionNum(tx, schema)
				So(versionNum, ShouldEqual, "version2")
				So(err, ShouldBeNil)
			})

			// downgrade schema
			executeInTransaction(t, db, func(tx *sqlx.Tx) {
				So(executeSchemaMigrations(tx, schema, "version2", "version1", true), ShouldBeNil)
				So(tx.Commit(), ShouldBeNil)
			})

			executeInTransaction(t, db, func(tx *sqlx.Tx) {
				exists, err := tableExists(tx, schema, "_first")
				So(exists, ShouldBeTrue)
				So(err, ShouldBeNil)
				exists, err = tableExists(tx, schema, "_second")
				So(exists, ShouldBeFalse)
				So(err, ShouldBeNil)
				versionNum, err := currentVersionNum(tx, schema)
				So(versionNum, ShouldEqual, "version1")
				So(err, ShouldBeNil)
			})
		})
	})
}
