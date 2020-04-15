//+build wireinject

package task

import (
	"context"

	"github.com/google/wire"

	pkg "github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/core/async"
)

var DependencySet = wire.NewSet(
	ProvideLoggingRequestID,
)

func newWelcomeEmailSendTask(ctx context.Context, m pkg.DependencyMap) async.Task {
	wire.Build(
		pkg.CommonDependencySet,
		DependencySet,
		wire.Struct(new(WelcomeEmailSendTask), "*"),
		wire.Bind(new(async.Task), new(*WelcomeEmailSendTask)),
	)
	return nil
}

func newVerifyCodeSendTask(ctx context.Context, m pkg.DependencyMap) async.Task {
	wire.Build(
		pkg.CommonDependencySet,
		DependencySet,
		wire.Struct(new(VerifyCodeSendTask), "*"),
		wire.Bind(new(async.Task), new(*VerifyCodeSendTask)),
	)
	return nil
}

func newPwHouseKeeperTask(ctx context.Context, m pkg.DependencyMap) async.Task {
	wire.Build(
		pkg.CommonDependencySet,
		DependencySet,
		wire.Struct(new(PwHousekeeperTask), "*"),
		wire.Bind(new(async.Task), new(*PwHousekeeperTask)),
	)
	return nil
}
