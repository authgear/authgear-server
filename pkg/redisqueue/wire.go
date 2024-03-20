//go:build wireinject
// +build wireinject

package redisqueue

import (
	"context"

	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/lib/userimport"
)

func newUserImportService(ctx context.Context, p *deps.AppProvider) *userimport.UserImportService {
	panic(wire.Build(
		deps.RedisQueueDependencySet,
		deps.CommonDependencySet,
	))
}
