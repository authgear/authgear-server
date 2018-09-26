package principal

import (
	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/core/db"
)

type Store interface {
	CreatePrincipal(Principal) error
}

type StoreImpl struct {
	sqlBuilder  db.SQLBuilder
	sqlExecutor db.SQLExecutor
	logger      *logrus.Entry
}

func NewStore(builder db.SQLBuilder, executor db.SQLExecutor, logger *logrus.Entry) *StoreImpl {
	return &StoreImpl{
		sqlBuilder:  builder,
		sqlExecutor: executor,
		logger:      logger,
	}
}

func (s StoreImpl) CreatePrincipal(principal Principal) error {
	builder := s.sqlBuilder.Insert(s.sqlBuilder.TableName("principal")).Columns(
		"id",
		"provider",
		"user_id",
	).Values(
		principal.ID,
		principal.Provider,
		principal.UserID,
	)

	_, err := s.sqlExecutor.ExecWith(builder)
	return err
}
