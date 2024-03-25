package redisqueue

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

var QueueUserImport = "user-import"
var QueueUserReindex = "user-reindex"

var ProducerDependencySet = wire.NewSet(
	NewUserImportProducer,
	NewUserReindexProducer,
)

type UserImportProducer struct {
	*Producer
}

func NewUserImportProducer(redis *appredis.Handle, clock clock.Clock) *UserImportProducer {
	return &UserImportProducer{
		&Producer{
			QueueName: QueueUserImport,
			Redis:     redis,
			Clock:     clock,
		},
	}
}

type UserReindexProducer struct {
	*Producer
}

func NewUserReindexProducer(redis *appredis.Handle, clock clock.Clock) *UserReindexProducer {
	return &UserReindexProducer{
		&Producer{
			QueueName: QueueUserReindex,
			Redis:     redis,
			Clock:     clock,
		},
	}
}
