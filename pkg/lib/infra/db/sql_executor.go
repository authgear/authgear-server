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
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/authgear/authgear-server/pkg/lib/infra/db/global"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/tenant"
	"github.com/authgear/authgear-server/pkg/util/errorutil"
)

type TenantSQLExecutor struct {
	SQLExecutor
}

func NewTenantSQLExecutor(c context.Context, handle *tenant.Handle) *TenantSQLExecutor {
	return &TenantSQLExecutor{
		SQLExecutor{
			Context:  c,
			Database: handle,
		},
	}
}

type GlobalSQLExecutor struct {
	SQLExecutor
}

func NewGlobalSQLExecutor(c context.Context, handle *global.Handle) *GlobalSQLExecutor {
	return &GlobalSQLExecutor{
		SQLExecutor{
			Context:  c,
			Database: handle,
		},
	}
}

type SQLExecutor struct {
	Context  context.Context
	Database Handle
}

func (e *SQLExecutor) ExecWith(sqlizeri sq.Sqlizer) (sql.Result, error) {
	db, err := e.Database.Conn()
	if err != nil {
		return nil, err
	}
	sql, args, err := sqlizeri.ToSql()
	if err != nil {
		return nil, err
	}
	result, err := db.ExecContext(e.Context, sql, args...)
	if err != nil {
		if isWriteConflict(err) {
			panic(ErrWriteConflict)
		}
		return nil, errorutil.WithDetails(err, errorutil.Details{"sql": errorutil.SafeDetail.Value(sql)})
	}
	return result, nil
}

func (e *SQLExecutor) QueryWith(sqlizeri sq.Sqlizer) (*sqlx.Rows, error) {
	db, err := e.Database.Conn()
	if err != nil {
		return nil, err
	}
	sql, args, err := sqlizeri.ToSql()
	if err != nil {
		return nil, err
	}
	result, err := db.QueryxContext(e.Context, sql, args...)
	if err != nil {
		if isWriteConflict(err) {
			panic(ErrWriteConflict)
		}
		return nil, errorutil.WithDetails(err, errorutil.Details{"sql": errorutil.SafeDetail.Value(sql)})
	}
	return result, nil
}

func (e *SQLExecutor) QueryRowWith(sqlizeri sq.Sqlizer) (*sqlx.Row, error) {
	db, err := e.Database.Conn()
	if err != nil {
		return nil, err
	}
	sql, args, err := sqlizeri.ToSql()
	if err != nil {
		if isWriteConflict(err) {
			panic(ErrWriteConflict)
		}
		return nil, errorutil.WithDetails(err, errorutil.Details{"sql": errorutil.SafeDetail.Value(sql)})
	}
	return db.QueryRowxContext(e.Context, sql, args...), nil
}

func isWriteConflict(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		// 40001: serialization_failure
		// 40P01: deadlock_detected
		return pqErr.Code == "40001" || pqErr.Code == "40P01"
	}
	return false
}
