package redisqueue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/globalredis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redisqueue"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/util/backoff"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/errorutil"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/panicutil"
	"github.com/authgear/authgear-server/pkg/util/signalutil"
)

// timeout is a reasonble number that does not block too long,
// and does not poll redis too frequently.
var timeout = 10 * time.Second

var errNoTask = errors.New("no task in queue")

const taskQueueBucket ratelimit.BucketName = "TaskQueue"

type TaskProcessor func(ctx context.Context, appProvider *deps.AppProvider, task *redisqueue.Task) (output json.RawMessage, err error)

type Consumer struct {
	QueueName              string
	clock                  clock.Clock
	rootProvider           *deps.RootProvider
	configSourceController *configsource.Controller
	taskProcessor          TaskProcessor
	redis                  *globalredis.Handle
	logger                 *log.Logger

	dequeueBackoff *backoff.Counter
	limitBucket    ratelimit.BucketSpec
	limiter        *ratelimit.LimiterGlobal

	// shutdown is for breaking the loop.
	shutdown chan struct{}
	// shutdown blocks Stop until the loop has ended.
	shutdownDone chan struct{}
	// shutdownCtx is for shutdown timeout.
	shutdownCtx context.Context
}

var _ signalutil.Daemon = &Consumer{}

func NewConsumer(ctx context.Context, queueName string, rateLimitConfig config.RateLimitsEnvironmentConfigEntry, rootProvider *deps.RootProvider, configSourceController *configsource.Controller, taskProcessor TaskProcessor) *Consumer {
	redis := globalredis.NewHandle(
		rootProvider.RedisPool,
		&rootProvider.EnvironmentConfig.RedisConfig,
		&rootProvider.EnvironmentConfig.GlobalRedis,
		rootProvider.LoggerFactory,
	)

	return &Consumer{
		QueueName:              queueName,
		clock:                  clock.NewSystemClock(),
		rootProvider:           rootProvider,
		configSourceController: configSourceController,
		taskProcessor:          taskProcessor,
		redis:                  redis,
		logger:                 rootProvider.LoggerFactory.New("redis-queue-consumer"),
		dequeueBackoff: &backoff.Counter{
			Interval:    time.Millisecond * 500,
			MaxInterval: time.Second * 10,
		},
		limitBucket: ratelimit.NewGlobalBucketSpec(rateLimitConfig, taskQueueBucket, queueName),
		limiter: &ratelimit.LimiterGlobal{
			Logger:  ratelimit.NewLogger(rootProvider.LoggerFactory),
			Storage: ratelimit.NewGlobalStorageRedis(redis),
		},
		shutdown:     make(chan struct{}),
		shutdownDone: make(chan struct{}),
		shutdownCtx:  ctx,
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
			c.work(ctx)
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

func (c *Consumer) dequeue(ctx context.Context) (*redisqueue.Task, *deps.AppProvider, error) {
	var task redisqueue.Task
	var appProvider *deps.AppProvider

	err := c.redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
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

		appCtx, err := c.configSourceController.ResolveContext(ctx, queueItem.AppID)
		if err != nil {
			return fmt.Errorf("resolve app context: %w", err)
		}

		appProvider = c.rootProvider.NewAppProvider(ctx, appCtx)
		appProvider.LoggerFactory.DefaultFields["task_id"] = task.ID
		return nil
	})

	if err != nil {
		return nil, nil, err
	}

	return &task, appProvider, err
}

func (c *Consumer) process(
	ctx context.Context,
	task *redisqueue.Task,
	appProvider *deps.AppProvider,
) (result json.RawMessage, err error) {
	defer func() {
		if e := recover(); e != nil {
			// Transform any panic into a saml result
			err = panicutil.MakeError(e)
		}
	}()

	result, err = c.taskProcessor(ctx, appProvider, task)
	return
}

func (c *Consumer) work(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			c.logger.WithFields(map[string]interface{}{
				"queue_name": c.QueueName,
				"error":      r,
				"stack":      errorutil.Callers(8),
			}).Error("panic occurred when running task")
		}
	}()

	if backoff := c.dequeueBackoff.BackoffDuration(); backoff != 0 {
		c.logger.WithField("delay", backoff).Info("backoff from dequeue")
		select {
		case <-c.shutdown:
			return
		case <-time.After(backoff):
			break
		}
	}

	// Reserve before dequeue.
	// If we do it in reverse order, we could dequeue the job, but rate limited.
	// And the server shuts down, the job will be gone forever.
	//
	// However, reserve-before-dequeue has its own problem.
	// Assume the rate limit is 10/3m, and there are 3 workers.
	//
	// At t=0, the queue is empty.
	// Since the burst (10) is larger than the number of workers (3),
	// All workers can reserve().
	// And then they all dequeue().
	// Since the queue is empty, they are blocked by BRPOP.
	// When the are unblocked by BRPOP after timeout,
	// they cancel the reservation.
	// And the progress repeats.
	// Once the queue becomes non-empty,
	// the workers will dequeue at the same time.
	var reservation *ratelimit.Reservation
	for {
		// This loop breaks until we can have a reservation without error.
		// This usage is different from HTTP request rate limit.
		// 1. In processing HTTP request, when we encounter rate limit error, we immediately terminate the request.
		// 2. So we rely on the caller of the HTTP request to retry.
		// 3. As the caller retries before time-to-act, they are rate limited again.
		// However, in here, we need to retry ourselves.
		// So we need a forever loop to retry until we are told we can proceed.

		var failedReservation *ratelimit.FailedReservation
		var err error
		reservation, failedReservation, err = c.limiter.Reserve(ctx, c.limitBucket)
		if err != nil {
			c.logger.WithError(err).Error("failed to check rate limit")
			c.dequeueBackoff.Increment()
			return
		}

		if reservation != nil {
			// The loop breaks here.
			break
		}

		// Otherwise, we are rate limited.
		c.logger.
			WithField("tat", failedReservation.GetTimeToAct()).
			WithField("bucket_key", c.limitBucket.Key()).
			Info("task rate limited")
		select {
		case <-c.shutdown:
			return
		case <-time.After(failedReservation.GetTimeToAct().Sub(c.clock.NowUTC())):
			// We wait until time-to-act and retry.
		}
	}

	task, appProvider, err := c.dequeue(ctx)
	if errors.Is(err, errNoTask) {
		// There is actually no task.
		// Cancel the reservation
		c.logger.
			WithField("bucket_key", c.limitBucket.Key()).
			// This is Debug instead of Info because it prints periodically.
			Debug("cancel reservation due to no task")
		c.limiter.Cancel(ctx, reservation)
		return
	} else if err != nil {
		c.logger.WithError(err).Error("failed to dequeue task")
		c.dequeueBackoff.Increment()
		return
	}

	c.logger.
		WithField("bucket_key", c.limitBucket.Key()).
		Info("consume reservation")

	// Reset backoff when we can dequeue.
	c.dequeueBackoff.Reset()

	output, err := c.process(ctx, task, appProvider)

	if err != nil {
		c.logger.WithFields(map[string]interface{}{
			"queue_name": c.QueueName,
			"error":      err,
			"stack":      errorutil.Callers(8),
		}).Error("failed to process task")
		task.Error = apierrors.AsAPIError(err)
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

	err = c.redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
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
