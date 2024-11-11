package event

import (
	"context"
	"fmt"

	appdb "github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
)

type StoreImpl struct {
	SQLBuilder  *appdb.SQLBuilder
	SQLExecutor *appdb.SQLExecutor
}

func (s *StoreImpl) NextSequenceNumber(ctx context.Context) (seq int64, err error) {
	builder := s.SQLBuilder.WithoutAppID().
		Select(fmt.Sprintf("nextval('%s')", s.SQLBuilder.TableName("_auth_event_sequence")))
	row, err := s.SQLExecutor.QueryRowWith(ctx, builder)
	if err != nil {
		return
	}
	err = row.Scan(&seq)
	return
}

func NewStoreImpl(
	sqlBuilder *appdb.SQLBuilder,
	sqlExecutor *appdb.SQLExecutor) *StoreImpl {
	return &StoreImpl{
		SQLBuilder:  sqlBuilder,
		SQLExecutor: sqlExecutor,
	}
}
