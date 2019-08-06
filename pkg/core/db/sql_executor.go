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

package db

import (
	"context"
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"

	"github.com/skygeario/skygear-server/pkg/core/logging"
)

type SQLExecutor struct {
	context        context.Context
	dbContext      Context
	statementCount int
}

func NewSQLExecutor(ctx context.Context, dbContext Context) SQLExecutor {
	return SQLExecutor{
		context:   ctx,
		dbContext: dbContext,
	}
}

func (e *SQLExecutor) Get(dest interface{}, query string, args ...interface{}) (err error) {
	logger := logging.CreateLoggerWithContext(e.context, "skydb").WithField("tag", "sql")
	e.statementCount++
	err = e.dbContext.DB().GetContext(e.context, dest, query, args...)
	logFields := logrus.Fields{
		"sql":            logging.StringValueFormatter(query),
		"args":           args,
		"executionCount": e.statementCount,
	}
	if err != nil {
		logger.WithFields(logFields).WithError(err).Errorln("Failed to execute SQL with sql.Get")
	} else {
		logger.WithFields(logFields).Debugln("Executed SQL successfully with sql.Get")
	}
	return
}

func (e *SQLExecutor) GetWith(dest interface{}, sqlizeri sq.Sqlizer) (err error) {
	sql, args, err := sqlizeri.ToSql()
	if err != nil {
		panic(err)
	}
	return e.Get(dest, sql, args...)
}

func (e *SQLExecutor) Exec(query string, args ...interface{}) (result sql.Result, err error) {
	logger := logging.CreateLoggerWithContext(e.context, "skydb").WithField("tag", "sql")
	e.statementCount++
	result, err = e.dbContext.DB().ExecContext(e.context, query, args...)

	var rowsAffected int64
	if result != nil {
		var rowsAffectedError error
		rowsAffected, rowsAffectedError = result.RowsAffected()
		if rowsAffectedError != nil {
			// because the row affected is only used for logging here
			// it is okay to ignore if RowsAffected() returns error
			logger.Debugf("conn: unable to get rows affected: %s", rowsAffectedError)
		}
	}

	logFields := logrus.Fields{
		"sql":            logging.StringValueFormatter(query),
		"args":           args,
		"executionCount": e.statementCount,
		"rowsAffected":   rowsAffected,
	}
	if err != nil {
		logger.WithFields(logFields).WithError(err).Errorln("Failed to execute SQL with sql.Exec")
	} else {
		logger.WithFields(logFields).Debugln("Executed SQL successfully with sql.Exec")
	}
	return
}

func (e *SQLExecutor) ExecWith(sqlizeri sq.Sqlizer) (sql.Result, error) {
	sql, args, err := sqlizeri.ToSql()
	if err != nil {
		panic(err)
	}
	return e.Exec(sql, args...)
}

func (e *SQLExecutor) Queryx(query string, args ...interface{}) (rows *sqlx.Rows, err error) {
	logger := logging.CreateLoggerWithContext(e.context, "skydb").WithField("tag", "sql")
	e.statementCount++
	rows, err = e.dbContext.DB().QueryxContext(e.context, query, args...)
	logFields := logrus.Fields{
		"sql":            logging.StringValueFormatter(query),
		"args":           args,
		"executionCount": e.statementCount,
	}
	if err != nil {
		logger.WithFields(logFields).WithError(err).Errorln("Failed to execute SQL with sql.Queryx")
	} else {
		logger.WithFields(logFields).Debugln("Executed SQL successfully with sql.Queryx")
	}
	return
}

func (e *SQLExecutor) QueryWith(sqlizeri sq.Sqlizer) (*sqlx.Rows, error) {
	sql, args, err := sqlizeri.ToSql()
	if err != nil {
		panic(err)
	}
	return e.Queryx(sql, args...)
}

func (e *SQLExecutor) QueryRowx(query string, args ...interface{}) (row *sqlx.Row) {
	logger := logging.CreateLoggerWithContext(e.context, "skydb").WithField("tag", "sql")
	e.statementCount++
	row = e.dbContext.DB().QueryRowxContext(e.context, query, args...)
	logger.WithFields(logrus.Fields{
		"sql":            logging.StringValueFormatter(query),
		"args":           args,
		"executionCount": e.statementCount,
	}).Debugln("Executed SQL with sql.QueryRowx")
	return
}

func (e *SQLExecutor) QueryRowWith(sqlizeri sq.Sqlizer) *sqlx.Row {
	sql, args, err := sqlizeri.ToSql()
	if err != nil {
		panic(err)
	}
	return e.QueryRowx(sql, args...)
}
