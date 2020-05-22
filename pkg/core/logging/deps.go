package logging

import (
	"context"

	"github.com/google/wire"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/sentry"
)

func ProvideLoggerFactory(ctx context.Context, c *config.TenantConfiguration) Factory {
	logHook := NewDefaultLogHook(c.DefaultSensitiveLoggerValues())
	sentryHook := sentry.NewLogHookFromContext(ctx)
	return NewFactory(logHook, sentryHook)
}

var DependencySet = wire.NewSet(
	ProvideLoggerFactory,
)
