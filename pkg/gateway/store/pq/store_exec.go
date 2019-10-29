package pq

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"

	"github.com/skygeario/skygear-server/pkg/core/errors"
)

func (s *Store) QueryRowWith(sqlizeri sq.Sqlizer) (*sqlx.Row, error) {
	sql, args, err := sqlizeri.ToSql()
	if err != nil {
		return nil, errors.WithDetails(err, errors.Details{"sql": errors.SafeDetail.Value(sql)})
	}
	return s.DB.QueryRowxContext(s.context, sql, args...), nil
}

func (s *Store) QueryWith(sqlizeri sq.Sqlizer) (*sqlx.Rows, error) {
	sql, args, err := sqlizeri.ToSql()
	if err != nil {
		return nil, err
	}
	result, err := s.DB.QueryxContext(s.context, sql, args...)
	if err != nil {
		return nil, errors.WithDetails(err, errors.Details{"sql": errors.SafeDetail.Value(sql)})
	}
	return result, nil
}
