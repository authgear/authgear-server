package pq

import (
	"github.com/jmoiron/sqlx"
	sq "github.com/lann/squirrel"
	"github.com/sirupsen/logrus"

	"github.com/skygeario/skygear-server/pkg/core/logging"
)

func (s *Store) QueryRowx(query string, args ...interface{}) (row *sqlx.Row) {
	row = s.DB.QueryRowxContext(s.context, query, args...)
	s.logger.WithFields(logrus.Fields{
		"sql":  query,
		"args": args,
	}).Debugln("Executed SQL with sql.QueryRowx")
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
	logFields := logrus.Fields{
		"sql":  logging.StringValueFormatter(query),
		"args": args,
	}
	if err != nil {
		s.logger.WithFields(logFields).WithError(err).Errorln("Failed to execute SQL with sql.Queryx")
	} else {
		s.logger.WithFields(logFields).Debugln("Executed SQL successfully with sql.Queryx")
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
