package deps

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/task"
)

type TaskQueueFactory func(*AppProvider) task.Queue

type TaskFunc func(ctx context.Context, param task.Param) error

func (f TaskFunc) Run(ctx context.Context, param task.Param) error {
	return f(ctx, param)
}

func ProvideCaptureTaskContext(config *config.Config, appCtx *config.AppContext) task.CaptureTaskContext {
	return func() *task.Context {
		return &task.Context{
			Config:     config,
			AppContext: appCtx,
		}
	}
}

func ProvideRestoreTaskContext(p *RootProvider) task.RestoreTaskContext {
	return func(ctx context.Context, taskCtx *task.Context) context.Context {
		rp := p.NewAppProvider(ctx, taskCtx.AppContext)
		ctx = withProvider(ctx, rp)
		return ctx
	}
}
