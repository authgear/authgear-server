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
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/skygeario/skygear-server/pkg/server/logging"
)

var log = logging.LoggerEntry("skydb")

const VersionTableName = "_version"

var ErrMigrationDisabled = errors.New("skydb/pq/migration: migration disabled")

type Revision interface {
	Version() string
	Up(tx *sqlx.Tx) error
	Down(tx *sqlx.Tx) error
}

func findRevisionIndex(needle string) int {
	for i := range revisions {
		if revisions[i].Version() == needle {
			return i
		}
	}
	return -1
}

var findRevisions = func(current string, target string) []Revision {
	full := &fullMigration{}
	if current == "" && target == full.Version() {
		return []Revision{full}
	}

	currentIndex := -1
	targetIndex := -1

	if current != "" {
		currentIndex = findRevisionIndex(current)
		if currentIndex == -1 {
			panic("not found")
		}
	}

	if target != "" {
		targetIndex = findRevisionIndex(target)
		if targetIndex == -1 {
			panic("not found")
		}
	}

	if currentIndex == targetIndex {
		return []Revision{}
	}

	var results []Revision
	if targetIndex > currentIndex {
		for i := currentIndex + 1; i <= targetIndex; i++ {
			results = append(results, revisions[i])
		}
	} else {
		for i := currentIndex; i >= targetIndex+1; i-- {
			results = append(results, revisions[i])
		}
	}
	return results
}

func executeSchemaMigrations(tx *sqlx.Tx, schema string, original string, target string, downgrade bool) (err error) {
	if err = ensureSchema(tx, schema); err != nil {
		return err
	}

	currentRevision := original
	revs := findRevisions(original, target)
	for i := range revs {
		revision := revs[i]
		if downgrade {
			err = revision.Down(tx)
		} else {
			err = revision.Up(tx)
		}

		if err != nil {
			log.Errorf(`Error executing schema migration "%s" -> "%s": %v`,
				currentRevision, revision.Version(), err)
			return err
		}
		log.Infof(`Executed schema migration "%s" -> "%s".`,
			currentRevision, revision.Version())

		currentRevision = revision.Version()
	}

	if err = ensureVersionTable(tx, schema); err != nil {
		return err
	}

	return setVersionNum(tx, original, target)
}

func EnsureLatest(db *sqlx.DB, schema string, allowMigration bool) error {
	tx, err := db.Beginx()
	if err != nil {
		log.Errorf(`Unable to begin transaction for schema migration: %v`, err)
		return err
	}
	defer tx.Rollback()

	versionNum, err := currentVersionNum(tx, schema)
	if err != nil {
		log.Errorf(`Unable to detetermine current schema version: %v`, err)
		return err
	}

	full := &fullMigration{}
	if versionNum == "" {
		log.Debugf(`Database schema is uninitialized. Latest schema: "%s"`, full.Version())
	} else if versionNum == full.Version() {
		log.Debugf(`Database schema "%s" matches the latest schema "%s".`, versionNum, full.Version())
	} else {
		log.Debugf(`Database schema "%s" does not match the latest schema "%s".`, versionNum, full.Version())
	}

	if versionNum == full.Version() {
		// no migration required
		return nil
	}

	if !allowMigration {
		log.Warnf(`Database schema does not match latest schema but migration is disabled.`)
		return ErrMigrationDisabled
	}

	log.Infof(`Database schema requires migration.`)

	if err := executeSchemaMigrations(tx, schema, versionNum, full.Version(), false); err != nil {
		return fmt.Errorf("skydb/pq: failed to init database: %v", err)
	}

	if err := tx.Commit(); err != nil {
		log.Errorf(`Unable to commit transaction for schema migration: %v`, err)
		return fmt.Errorf("skydb/pq: failed to commit DDL: %v", err)
	}

	return nil
}

func ensureSchema(tx *sqlx.Tx, schema string) error {
	_, err := tx.Exec(fmt.Sprintf(`CREATE SCHEMA IF NOT EXISTS %s;`, schema))
	if err != nil {
		return err
	}

	// Due to database/sql connection polling, this function must be
	// executed within an transaction because the connection can be
	// different for each execution otherwise.
	_, err = tx.Exec(fmt.Sprintf(`SET search_path TO %s, public;`, schema))
	return err
}

func ensureVersionTable(tx *sqlx.Tx, schema string) error {
	_, err := tx.Exec(fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s.%s (
	version_num character varying(32) NOT NULL
);`, schema, VersionTableName))
	return err
}

func tableExists(tx *sqlx.Tx, schema string, table string) (bool, error) {
	var exists bool
	err := tx.QueryRowx(`SELECT EXISTS (
    SELECT 1 
    FROM   pg_catalog.pg_class c
    JOIN   pg_catalog.pg_namespace n ON n.oid = c.relnamespace
    WHERE  n.nspname = $1
    AND    c.relname = $2
);`, schema, table).Scan(&exists)
	return exists, err
}

func versionTableExists(tx *sqlx.Tx, schema string) (bool, error) {
	return tableExists(tx, schema, VersionTableName)
}

func currentVersionNum(tx *sqlx.Tx, schema string) (string, error) {
	exists, err := versionTableExists(tx, schema)
	if !exists {
		log.Debugf(`Version table "%s" does not exist in schema "%s".`, VersionTableName, schema)
		return "", nil
	}
	if err != nil {
		return "", err
	}

	var versionNum string
	err = tx.QueryRowx(fmt.Sprintf(`SELECT version_num FROM %s.%s`, schema, VersionTableName)).
		Scan(&versionNum)
	if err != nil && err != sql.ErrNoRows {
		return "", err
	}

	log.Debugf(`Current version of database schema is "%s".`, versionNum)
	return versionNum, nil
}

func setVersionNum(tx *sqlx.Tx, original string, target string) error {
	if original == "" {
		_, err := tx.Exec(fmt.Sprintf(`INSERT INTO %s (version_num) VALUES ($1);`, VersionTableName), target)
		return err
	}

	_, err := tx.Exec(fmt.Sprintf(`UPDATE %s SET version_num = $1 WHERE version_num = $2;`, VersionTableName), target, original)
	return err
}

func getAllRecordTables(tx *sqlx.Tx) ([]string, error) {
	rows, err := tx.Queryx(`
	SELECT tablename FROM pg_tables WHERE schemaname=current_schema()
		AND tablename NOT LIKE '_%';
	`)
	if err != nil {
		return nil, err
	}

	var results []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, err
		}
		results = append(results, tableName)
	}
	return results, nil
}
