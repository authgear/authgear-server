//go:build wireinject
// +build wireinject

package redisqueue

import (
	"context"

	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/cloudstorage"
	"github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/lib/elasticsearch"
	"github.com/authgear/authgear-server/pkg/lib/userexport"
	"github.com/authgear/authgear-server/pkg/lib/userimport"
)

func newUserImportService(ctx context.Context, p *deps.AppProvider) *userimport.UserImportService {
	panic(wire.Build(
		deps.RedisQueueDependencySet,
		deps.CommonDependencySet,
	))
}

func newUserExportService(ctx context.Context, p *deps.AppProvider) *userexport.UserExportService {
	panic(wire.Build(
		deps.RedisQueueDependencySet,
		deps.CommonDependencySet,
		NewCloudStorage,
		wire.Struct(new(cloudstorage.Provider), "*"),
	))
}

func newElasticsearchService(ctx context.Context, p *deps.AppProvider) *elasticsearch.Service {
	panic(wire.Build(
		deps.RedisQueueDependencySet,
		deps.CommonDependencySet,
	))
}
