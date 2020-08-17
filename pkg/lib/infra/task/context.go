package task

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

type CaptureTaskContext func() *Context

type RestoreTaskContext func(context.Context, *Context) context.Context

type Context struct {
	Config *config.Config
}
