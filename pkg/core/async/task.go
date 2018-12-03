package async

import (
	"context"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
)

type TaskFactory interface {
	NewTask(context TaskContext) Task
}

type Task interface {
	Run(param interface{}) error
}

type TaskFunc func(param interface{}) error

func (t TaskFunc) Run(param interface{}) error {
	return t(param)
}

type TaskContext struct {
	context.Context
	RequestID    string
	TenantConfig config.TenantConfiguration
}

func NewTaskContext(r *http.Request) TaskContext {
	return TaskContext{
		Context:      db.InitDBContext(context.Background()),
		RequestID:    r.Header.Get("X-Skygear-Request-ID"),
		TenantConfig: config.GetTenantConfig(r),
	}
}
