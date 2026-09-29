//go:build wireinject

package auditlogstreaming

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/globalredis"
	"github.com/authgear/authgear-server/pkg/util/backgroundjob"
)

func newRunnable(
	redis *globalredis.Handle,
	interval config.DurationString,
	appContextResolver AppContextResolver,
	senderFactory SenderFactory,
) backgroundjob.Runnable {
	panic(wire.Build(RunnableDependencySet))
}
