package deps

import (
	"context"

	"github.com/skygeario/skygear-server/pkg/auth/config"
	"github.com/skygeario/skygear-server/pkg/task"
	"github.com/skygeario/skygear-server/pkg/task/executors"
	"github.com/skygeario/skygear-server/pkg/task/queue"
)

func ProvideCaptureTaskContext(ctx context.Context, config *config.Config) queue.CaptureTaskContext {
	return func() *task.Context {
		return &task.Context{
			Config: config,
		}
	}
}

func ProvideRestoreTaskContext(p *RootProvider) executors.RestoreTaskContext {
	return func(taskCtx *task.Context) context.Context {
		ctx := context.Background()
		rp := p.NewRequestProvider(ctx, nil, taskCtx.Config)
		ctx = WithRequestProvider(ctx, rp)
		return ctx
	}
}
