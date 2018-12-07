package async

import (
	"context"

	"github.com/skygeario/skygear-server/pkg/core/config"
)

type TaskFactory interface {
	NewTask(ctx context.Context, taskCtx TaskContext) Task
}

type Task interface {
	Run(param interface{}) error
}

type TaskFunc func(param interface{}) error

func (t TaskFunc) Run(param interface{}) error {
	return t(param)
}

type TaskContext struct {
	RequestID    string
	TenantConfig config.TenantConfiguration
}
