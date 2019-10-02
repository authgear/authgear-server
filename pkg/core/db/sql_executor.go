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
	context   context.Context
	dbContext Context
	logger    *logrus.Entry
}

func NewSQLExecutor(ctx context.Context, dbContext Context, loggerFactory logging.Factory) SQLExecutor {
	return SQLExecutor{
		context:   ctx,
		dbContext: dbContext,
		logger:    loggerFactory.NewLogger("sql-executor"),
	}
}

func (e *SQLExecutor) Get(dest interface{}, query string, args ...interface{}) (err error) {
	err = e.dbContext.DB().GetContext(e.context, dest, query, args...)
	if err != nil {
		e.logger.WithField("sql", query).WithError(err).Errorln("Failed to execute SQL with sql.Get")
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
	result, err = e.dbContext.DB().ExecContext(e.context, query, args...)

	if err != nil {
		e.logger.WithField("sql", query).WithError(err).Errorln("Failed to execute SQL with sql.Exec")
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
	rows, err = e.dbContext.DB().QueryxContext(e.context, query, args...)
	if err != nil {
		e.logger.WithField("sql", query).WithError(err).Errorln("Failed to execute SQL with sql.Queryx")
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
	row = e.dbContext.DB().QueryRowxContext(e.context, query, args...)
	return
}

func (e *SQLExecutor) QueryRowWith(sqlizeri sq.Sqlizer) *sqlx.Row {
	sql, args, err := sqlizeri.ToSql()
	if err != nil {
		panic(err)
	}
	return e.QueryRowx(sql, args...)
}
