package pq

import (
	"database/sql"

	log "github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"
	sq "github.com/lann/squirrel"
)

func (c *conn) Get(dest interface{}, query string, args ...interface{}) (err error) {
	c.statementCount++
	err = c.Db().Get(dest, query, args...)
	logFields := log.Fields{
		"sql":            query,
		"args":           args,
		"error":          err,
		"executionCount": c.statementCount,
	}
	if err != nil {
		log.WithFields(logFields).Errorln("Failed to execute SQL with sql.Get")
	} else {
		log.WithFields(logFields).Debugln("Executed SQL successfully with sql.Get")
	}
	return
}

func (c *conn) GetWith(dest interface{}, sqlizeri sq.Sqlizer) (err error) {
	sql, args, err := sqlizeri.ToSql()
	if err != nil {
		panic(err)
	}
	return c.Get(dest, sql, args...)
}

func (c *conn) Exec(query string, args ...interface{}) (result sql.Result, err error) {
	c.statementCount++
	result, err = c.Db().Exec(query, args...)

	var rowsAffected int64
	if result != nil {
		var rowsAffectedError error
		rowsAffected, rowsAffectedError = result.RowsAffected()
		if rowsAffectedError != nil {
			// because the row affected is only used for logging here
			// it is okay to ignore if RowsAffected() returns error
			log.Debugf("conn: unable to get rows affected: %s", rowsAffectedError)
		}
	}

	logFields := log.Fields{
		"sql":            query,
		"args":           args,
		"error":          err,
		"executionCount": c.statementCount,
		"rowsAffected":   rowsAffected,
	}
	if err != nil {
		log.WithFields(logFields).Errorln("Failed to execute SQL with sql.Exec")
	} else {
		log.WithFields(logFields).Debugln("Executed SQL successfully with sql.Exec")
	}
	return
}

func (c *conn) ExecWith(sqlizeri sq.Sqlizer) (sql.Result, error) {
	sql, args, err := sqlizeri.ToSql()
	if err != nil {
		panic(err)
	}
	return c.Exec(sql, args...)
}

func (c *conn) Queryx(query string, args ...interface{}) (rows *sqlx.Rows, err error) {
	c.statementCount++
	rows, err = c.Db().Queryx(query, args...)
	logFields := log.Fields{
		"sql":            query,
		"args":           args,
		"error":          err,
		"executionCount": c.statementCount,
	}
	if err != nil {
		log.WithFields(logFields).Errorln("Failed to execute SQL with sql.Queryx")
	} else {
		log.WithFields(logFields).Debugln("Executed SQL successfully with sql.Queryx")
	}
	return
}

func (c *conn) QueryWith(sqlizeri sq.Sqlizer) (*sqlx.Rows, error) {
	sql, args, err := sqlizeri.ToSql()
	if err != nil {
		panic(err)
	}
	return c.Queryx(sql, args...)
}

func (c *conn) QueryRowx(query string, args ...interface{}) (row *sqlx.Row) {
	c.statementCount++
	row = c.Db().QueryRowx(query, args...)
	log.WithFields(log.Fields{
		"sql":            query,
		"args":           args,
		"executionCount": c.statementCount,
	}).Debugln("Executed SQL with sql.QueryRowx")
	return
}

func (c *conn) QueryRowWith(sqlizeri sq.Sqlizer) *sqlx.Row {
	sql, args, err := sqlizeri.ToSql()
	if err != nil {
		panic(err)
	}
	return c.QueryRowx(sql, args...)
}
