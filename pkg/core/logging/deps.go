package logging

import (
	"context"

	"github.com/google/wire"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/sentry"
)

type RequestID string

func ProvideLoggerFactory(ctx context.Context, rid RequestID, c *config.TenantConfiguration) Factory {
	logHook := NewDefaultLogHook(c.DefaultSensitiveLoggerValues())
	sentryHook := sentry.NewLogHookFromContext(ctx)
	return NewFactoryFromRequestID(string(rid), logHook, sentryHook)
}

var DependencySet = wire.NewSet(
	ProvideLoggerFactory,
)
