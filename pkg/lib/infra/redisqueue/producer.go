package redisqueue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"time"

	goredis "github.com/go-redis/redis/v8"

	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/util/base32"
	"github.com/authgear/authgear-server/pkg/util/clock"
	corerand "github.com/authgear/authgear-server/pkg/util/rand"
)

const (
	idAlphabet string = base32.Alphabet
	idLength   int    = 32
)

var rng *rand.Rand = corerand.SecureRand

var TTL = 24 * time.Hour

type Producer struct {
	QueueName string
	Redis     *appredis.Handle
	Clock     clock.Clock
}

func newTaskID(taskIDPrefix string) string {
	return fmt.Sprintf("%v_%v", taskIDPrefix, corerand.StringWithAlphabet(idLength, idAlphabet, rng))
}

func (p *Producer) NewTask(appID string, input json.RawMessage, taskIDPrefix string) *Task {
	id := newTaskID(taskIDPrefix)
	now := p.Clock.NowUTC()
	task := &Task{
		ID:        id,
		CreatedAt: &now,
		Status:    TaskStatusPending,
		AppID:     appID,
		Input:     input,
	}
	return task
}

func (p *Producer) EnqueueTask(ctx context.Context, task *Task) error {
	return p.Redis.WithConnContext(ctx, func(conn redis.Redis_6_0_Cmdable) error {
		taskBytes, err := json.Marshal(task)
		if err != nil {
			return err
		}

		key := task.RedisKey()
		_, err = conn.SetNX(ctx, key, taskBytes, TTL).Result()
		if err != nil {
			return err
		}

		queueItem := task.ToQueueItem()
		queueItemBytes, err := json.Marshal(queueItem)
		if err != nil {
			return err
		}

		queueKey := RedisKeyForQueue(p.QueueName)

		_, err = conn.LPush(ctx, queueKey, queueItemBytes).Result()
		if err != nil {
			return err
		}

		return nil
	})
}

func (p *Producer) GetTask(ctx context.Context, item *QueueItem) (*Task, error) {
	var task Task
	err := p.Redis.WithConnContext(ctx, func(conn redis.Redis_6_0_Cmdable) error {
		taskBytes, err := conn.Get(ctx, item.RedisKey()).Bytes()
		if errors.Is(err, goredis.Nil) {
			return ErrTaskNotFound
		}
		if err != nil {
			return err
		}

		err = json.Unmarshal(taskBytes, &task)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &task, nil
}
