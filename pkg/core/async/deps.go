package async

import (
	"context"

	"github.com/google/wire"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
)

func ProvideTaskQueue(
	ctx context.Context,
	txContext db.TxContext,
	tenantConfig *config.TenantConfiguration,
	taskExecutor *Executor,
) Queue {
	return NewQueue(
		ctx,
		txContext,
		tenantConfig,
		taskExecutor,
	)
}

var DependencySet = wire.NewSet(ProvideTaskQueue)
