package pq

import (
	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/core/auth/role"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/record/dependency/record"
)

type recordStore struct {
	roleStore role.Store

	canMigrate bool

	sqlBuilder  db.SQLBuilder
	sqlExecutor db.SQLExecutor
	logger      *logrus.Entry
}

func NewRecordStore(
	roleStore role.Store,
	canMigrate bool,
	sqlBuilder db.SQLBuilder,
	sqlExecutor db.SQLExecutor,
	logger *logrus.Entry,
) record.Store {
	return &recordStore{
		roleStore:   roleStore,
		canMigrate:  canMigrate,
		sqlBuilder:  sqlBuilder,
		sqlExecutor: sqlExecutor,
		logger:      logger,
	}
}
