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

	"github.com/skygeario/skygear-server/pkg/core/errors"
)

type SQLExecutor struct {
	context   context.Context
	dbContext Context
}

func NewSQLExecutor(ctx context.Context, dbContext Context) SQLExecutor {
	return SQLExecutor{
		context:   ctx,
		dbContext: dbContext,
	}
}

func (e *SQLExecutor) ExecWith(sqlizeri sq.Sqlizer) (sql.Result, error) {
	db, err := e.dbContext.DB()
	if err != nil {
		return nil, err
	}
	sql, args, err := sqlizeri.ToSql()
	if err != nil {
		return nil, err
	}
	result, err := db.ExecContext(e.context, sql, args...)
	if err != nil {
		return nil, errors.WithDetails(err, errors.Details{"sql": errors.SafeString(sql)})
	}
	return result, nil
}

func (e *SQLExecutor) QueryWith(sqlizeri sq.Sqlizer) (*sqlx.Rows, error) {
	db, err := e.dbContext.DB()
	if err != nil {
		return nil, err
	}
	sql, args, err := sqlizeri.ToSql()
	if err != nil {
		return nil, err
	}
	result, err := db.QueryxContext(e.context, sql, args...)
	if err != nil {
		return nil, errors.WithDetails(err, errors.Details{"sql": errors.SafeString(sql)})
	}
	return result, nil
}

func (e *SQLExecutor) QueryRowWith(sqlizeri sq.Sqlizer) (*sqlx.Row, error) {
	db, err := e.dbContext.DB()
	if err != nil {
		return nil, err
	}
	sql, args, err := sqlizeri.ToSql()
	if err != nil {
		return nil, errors.WithDetails(err, errors.Details{"sql": errors.SafeString(sql)})
	}
	return db.QueryRowxContext(e.context, sql, args...), nil
}
