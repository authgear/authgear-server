package async

import (
	"context"

	"github.com/google/wire"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/logging"
)

func ProvideTaskQueue(
	ctx context.Context,
	txContext db.TxContext,
	requestID logging.RequestID,
	tenantConfig *config.TenantConfiguration,
	taskExecutor *Executor,
) Queue {
	return NewQueue(
		ctx,
		txContext,
		string(requestID),
		tenantConfig,
		taskExecutor,
	)
}

var DependencySet = wire.NewSet(ProvideTaskQueue)
