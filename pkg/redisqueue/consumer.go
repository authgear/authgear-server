package redisqueue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	goredis "github.com/go-redis/redis/v8"

	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/globalredis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redisqueue"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/errorutil"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/signalutil"
)

// timeout is a reasonble number that does not block too long,
// and does not poll redis too frequently.
var timeout = 10 * time.Second

var errNoTask = errors.New("no task in queue")

type TaskProcessor func(ctx context.Context, appProvider *deps.AppProvider, task *redisqueue.Task) (output json.RawMessage, err error)

type Consumer struct {
	QueueName              string
	clock                  clock.Clock
	rootProvider           *deps.RootProvider
	configSourceController *configsource.Controller
	taskProcessor          TaskProcessor
	redis                  *globalredis.Handle
	logger                 *log.Logger
	// shutdown is for breaking the loop.
	shutdown chan struct{}
	// shutdown blocks Stop until the loop has ended.
	shutdownDone chan struct{}
	// shutdownCtx is for shutdown timeout.
	shutdownCtx context.Context
}

var _ signalutil.Daemon = &Consumer{}

func NewConsumer(queueName string, rootProvider *deps.RootProvider, configSourceController *configsource.Controller, taskProcessor TaskProcessor) *Consumer {
	return &Consumer{
		QueueName:              queueName,
		clock:                  clock.NewSystemClock(),
		rootProvider:           rootProvider,
		configSourceController: configSourceController,
		taskProcessor:          taskProcessor,
		redis: globalredis.NewHandle(
			rootProvider.RedisPool,
			&rootProvider.EnvironmentConfig.RedisConfig,
			&rootProvider.EnvironmentConfig.GlobalRedis,
			rootProvider.LoggerFactory,
		),
		logger:       rootProvider.LoggerFactory.New("redis-queue-consumer"),
		shutdown:     make(chan struct{}),
		shutdownDone: make(chan struct{}),
		shutdownCtx:  context.Background(),
	}
}

func (c *Consumer) DisplayName() string {
	return fmt.Sprintf("redis-queue:%v", c.QueueName)
}

// Start starts draining the queue and blocks indefinitely.
// It should be called with go.
func (c *Consumer) Start0(ctx context.Context) {
loop:
	for {
		select {
		case <-c.shutdown:
			c.logger.Infof("shutdown gracefully")
			break loop
		case <-c.shutdownCtx.Done():
			c.logger.Infof("shutdown context timeout")
			break loop
		default:
			c.dequeue(ctx)
		}
	}
	close(c.shutdownDone)
}

func (c *Consumer) Start(ctx context.Context, _ *log.Logger) {
	c.Start0(ctx)
}

func (c *Consumer) Stop0(ctx context.Context) {
	c.shutdownCtx = ctx
	close(c.shutdown)
	<-c.shutdownDone
}

func (c *Consumer) Stop(ctx context.Context, _ *log.Logger) error {
	c.Stop0(ctx)
	return nil
}

func (c *Consumer) dequeue(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			c.logger.WithFields(map[string]interface{}{
				"queue_name": c.QueueName,
				"error":      r,
				"stack":      errorutil.Callers(8),
			}).Error("panic occurred when running task")
		}
	}()

	var task redisqueue.Task
	var appProvider *deps.AppProvider

	err := c.redis.WithConnContext(ctx, func(conn *goredis.Conn) error {
		queueKey := redisqueue.RedisKeyForQueue(c.QueueName)

		strs, err := conn.BRPop(ctx, timeout, queueKey).Result()
		if errors.Is(err, goredis.Nil) {
			// timeout.
			return errNoTask
		}
		if err != nil {
			// other errors.
			return fmt.Errorf("BRPOP queue: %w", err)
		}

		// The first item in the array is the queue name.
		// The second item in the array is the value.
		// See https://redis.io/commands/blpop/
		queueItemBytes := []byte(strs[1])
		var queueItem redisqueue.QueueItem
		err = json.Unmarshal(queueItemBytes, &queueItem)
		if err != nil {
			return fmt.Errorf("unmarshal queue item: %w", err)
		}

		taskBytes, err := conn.Get(ctx, queueItem.RedisKey()).Bytes()
		if errors.Is(err, goredis.Nil) {
			return errors.New("task item not found")
		}
		if err != nil {
			return fmt.Errorf("get task: %w", err)
		}

		err = json.Unmarshal(taskBytes, &task)
		if err != nil {
			return fmt.Errorf("unmarshal task: %w", err)
		}

		appCtx, err := c.configSourceController.ResolveContext(queueItem.AppID)
		if err != nil {
			return fmt.Errorf("resolve app context: %w", err)
		}

		appProvider = c.rootProvider.NewAppProvider(ctx, appCtx)
		appProvider.LoggerFactory.DefaultFields["task_id"] = task.ID
		return nil
	})

	if errors.Is(err, errNoTask) {
		return
	} else if err != nil {
		c.logger.WithError(err).Error("failed to dequeue task")
		return
	}

	output, err := c.taskProcessor(ctx, appProvider, &task)
	if err != nil {
		c.logger.WithError(err).Error("failed to process task")
		return
	}

	task.Status = redisqueue.TaskStatusCompleted
	completedAt := c.clock.NowUTC()
	task.CompletedAt = &completedAt
	task.Output = output

	taskBytes, err := json.Marshal(task)
	if err != nil {
		c.logger.WithError(err).Error("failed to marshal task")
		return
	}

	err = c.redis.WithConnContext(ctx, func(conn *goredis.Conn) error {
		key := task.RedisKey()
		_, err := conn.Set(ctx, key, taskBytes, redisqueue.TTL).Result()
		if err != nil {
			c.logger.WithError(err).Error("failed to save task output")
			return err
		}
		return nil
	})
	if err != nil {
		return
	}
}
