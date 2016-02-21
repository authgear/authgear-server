package migration

import (
	"database/sql"
	"errors"
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"
)

const VersionTableName = "_version"

var ErrMigrationDisabled = errors.New("skydb/pq/migration: migration disabled")

type Revision interface {
	Version() string
	Up(tx *sqlx.Tx) error
	Down(tx *sqlx.Tx) error
}

var findRevisions = func(original string, target string, downgrade bool) []Revision {
	return []Revision{
		&revision_full{},
	}
}

func executeSchemaMigrations(tx *sqlx.Tx, schema string, original string, target string, downgrade bool) (err error) {
	if err = ensureSchema(tx, schema); err != nil {
		return err
	}

	currentRevision := original
	revs := findRevisions(original, target, downgrade)
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

	if err = setVersionNum(tx, original, target); err != nil {
		return err
	}

	return nil
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

	if versionNum == "" {
		log.Debugf(`Database schema is uninitialized.`)
	} else if versionNum == DbVersionNum {
		log.Debugf(`Database schema "%s" matches the latest schema "%s".`, versionNum, DbVersionNum)
	} else {
		log.Debugf(`Database schema "%s" does not match the latest schema "%s".`, versionNum, DbVersionNum)
	}

	if versionNum == DbVersionNum {
		// no migration required
		return nil
	}

	if !allowMigration {
		log.Warnf(`Database schema does not match latest schema but migration is disabled.`)
		return ErrMigrationDisabled
	}

	log.Infof(`Database schema requires migration.`)

	if err := executeSchemaMigrations(tx, schema, versionNum, DbVersionNum, false); err != nil {
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

	// Due to database/sql connection polling, this function must be
	// executed within an transaction because the connection can be
	// different for each execution otherwise.
	_, err = tx.Exec(fmt.Sprintf(`SET search_path TO %s;`, schema))
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

const DbVersionNum = "551bc42a839"
