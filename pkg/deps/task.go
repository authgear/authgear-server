package deps

import (
	"context"

	"github.com/skygeario/skygear-server/pkg/auth/config"
	"github.com/skygeario/skygear-server/pkg/core/intl"
	"github.com/skygeario/skygear-server/pkg/task"
	"github.com/skygeario/skygear-server/pkg/task/executors"
	"github.com/skygeario/skygear-server/pkg/task/queue"
)

func ProvideCaptureTaskContext(ctx context.Context, config *config.Config) queue.CaptureTaskContext {
	return func() *task.Context {
		return &task.Context{
			Config:                config,
			PreferredLanguageTags: intl.GetPreferredLanguageTags(ctx),
		}
	}
}

func ProvideRestoreTaskContext(deps *RootContainer) executors.RestoreTaskContext {
	return func(taskCtx *task.Context) context.Context {
		ctx := context.Background()
		requestContainer := deps.NewRequestContainer(ctx, nil, taskCtx.Config)
		ctx = WithRequestContainer(ctx, requestContainer)
		ctx = intl.WithPreferredLanguageTags(ctx, taskCtx.PreferredLanguageTags)
		return ctx
	}
}
