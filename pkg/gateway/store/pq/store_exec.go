package pq

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

func (s *Store) QueryRowx(query string, args ...interface{}) (row *sqlx.Row) {
	row = s.DB.QueryRowxContext(s.context, query, args...)
	s.logger.WithField("sql", query).Debugln("Executed SQL with sql.QueryRowx")
	return
}

func (s *Store) QueryRowWith(sqlizeri sq.Sqlizer) *sqlx.Row {
	sql, args, err := sqlizeri.ToSql()
	if err != nil {
		panic(err)
	}
	return s.QueryRowx(sql, args...)
}

func (s *Store) Queryx(query string, args ...interface{}) (rows *sqlx.Rows, err error) {
	rows, err = s.DB.QueryxContext(s.context, query, args...)
	if err != nil {
		s.logger.WithField("sql", query).WithError(err).Errorln("Failed to execute SQL with sql.Queryx")
	}
	return
}

func (s *Store) QueryWith(sqlizeri sq.Sqlizer) (*sqlx.Rows, error) {
	sql, args, err := sqlizeri.ToSql()
	if err != nil {
		panic(err)
	}
	return s.Queryx(sql, args...)
}
