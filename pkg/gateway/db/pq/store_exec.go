package pq

import (
	"github.com/jmoiron/sqlx"
	sq "github.com/lann/squirrel"
	"github.com/sirupsen/logrus"

	"github.com/skygeario/skygear-server/pkg/core/logging"
)

func (s *store) QueryRowx(query string, args ...interface{}) (row *sqlx.Row) {
	logger := logging.LoggerEntry("gateway")
	row = s.DB.QueryRowxContext(s.context, query, args...)
	logger.WithFields(logrus.Fields{
		"sql":  query,
		"args": args,
	}).Debugln("Executed SQL with sql.QueryRowx")
	return
}

func (s *store) QueryRowWith(sqlizeri sq.Sqlizer) *sqlx.Row {
	sql, args, err := sqlizeri.ToSql()
	if err != nil {
		panic(err)
	}
	return s.QueryRowx(sql, args...)
}
