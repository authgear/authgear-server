package verifycode

import (
	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/core/db"
)

type Store interface {
	CreateVerifyCode(code *VerifyCode) error
}

type storeImpl struct {
	sqlBuilder  db.SQLBuilder
	sqlExecutor db.SQLExecutor
	logger      *logrus.Entry
}

func newStore(builder db.SQLBuilder, executor db.SQLExecutor, logger *logrus.Entry) *storeImpl {
	return &storeImpl{
		sqlBuilder:  builder,
		sqlExecutor: executor,
		logger:      logger,
	}
}

func NewStore(builder db.SQLBuilder, executor db.SQLExecutor, logger *logrus.Entry) Store {
	return newStore(builder, executor, logger)
}

func (s *storeImpl) CreateVerifyCode(code *VerifyCode) (err error) {
	builder := s.sqlBuilder.Insert(s.sqlBuilder.FullTableName("verify_code")).Columns(
		"id",
		"user_id",
		"record_key",
		"record_value",
		"code",
		"consumed",
		"created_at",
	).Values(
		code.ID,
		code.UserID,
		code.RecordKey,
		code.RecordValue,
		code.Code,
		code.Consumed,
		code.CreatedAt,
	)

	_, err = s.sqlExecutor.ExecWith(builder)
	return nil
}
