package task

import (
	"context"

	pkg "github.com/skygeario/skygear-server/pkg/auth"

	"github.com/skygeario/skygear-server/pkg/core/async"
)

type taskFunc func(ctx context.Context, param interface{}) error

func (f taskFunc) Run(ctx context.Context, param interface{}) error {
	return f(ctx, param)
}

func MakeTask(deps pkg.DependencyMap, factory func(ctx context.Context, m pkg.DependencyMap) async.Task) async.Task {
	return taskFunc(func(ctx context.Context, param interface{}) error {
		task := factory(ctx, deps)
		return task.Run(ctx, param)
	})
}
