//+build wireinject

package task

import (
	"context"

	"github.com/google/wire"

	pkg "github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/core/async"
)

func newPwHouseKeeperTask(ctx context.Context, m pkg.DependencyMap) async.Task {
	wire.Build(
		pkg.CommonDependencySet,
		wire.Struct(new(PwHousekeeperTask), "*"),
		wire.Bind(new(async.Task), new(*PwHousekeeperTask)),
	)
	return nil
}

func newSendMessagesTask(ctx context.Context, m pkg.DependencyMap) async.Task {
	wire.Build(
		pkg.CommonDependencySet,
		wire.Struct(new(SendMessagesTask), "*"),
		wire.Bind(new(async.Task), new(*SendMessagesTask)),
	)
	return nil
}
