package auditlogstreaming

import (
	"context"
	"errors"
	"log/slog"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/globalredis"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

// drainMutexName serialises delivery across background replicas:
// claimPendingScript and drainScript are individually atomic, but two
// replicas could still each claim a different half of the pending apps
// and deliver concurrently, which would interleave batches across
// connections to the same stream.
const drainMutexName = "audit-log-stream:drain"

var RunnableLogger = slogutil.NewLogger("audit-log-streaming-runner")

// AppContextResolver resolves an app ID to its current config for the
// duration of fn. Satisfied by *configsource.Controller.
type AppContextResolver interface {
	ResolveContext(ctx context.Context, appID string, fn func(context.Context, *config.AppContext) error) error
}

// SenderFactory builds the Sender for one app from its resolved config.
type SenderFactory interface {
	MakeSender(appID string, appCtx *config.AppContext) Sender
}

type Runnable struct {
	Consumer           *Consumer
	AppContextResolver AppContextResolver
	SenderFactory      SenderFactory
	Redis              *globalredis.Handle
	Interval           config.DurationString
}

// Run executes one tick: claim every app with a non-empty queue, and
// deliver each app's batch. Failure is per app, never per entry -- there
// is no per-entry acknowledgement to act on -- so Run returns nil unless
// claiming the pending set itself failed.
func (r *Runnable) Run(ctx context.Context) error {
	logger := RunnableLogger.GetLogger(ctx)

	// The mutex expiry is generous relative to the interval: a tick that
	// legitimately takes a while (many apps, many streams, slow
	// collectors) must not have its lock expire out from under it and let
	// a second replica start draining concurrently.
	err := r.Redis.WithMutexExpiry(ctx, drainMutexName, 2*r.Interval.Duration(), func() error {
		return r.drain(ctx, logger)
	})
	if errors.Is(err, redis.ErrMutexNotAcquired) {
		logger.Debug(ctx, "another replica is draining, skipping this tick")
		return nil
	}
	return err
}

func (r *Runnable) drain(ctx context.Context, logger slogutil.NamedLogger) error {
	appIDs, err := r.Consumer.ClaimPending(ctx)
	if err != nil {
		return err
	}

	for _, appID := range appIDs {
		entries, err := r.Consumer.Drain(ctx, appID)
		if err != nil {
			logger.WithError(err).Error(ctx, "failed to drain audit log stream queue",
				slog.String("app_id", appID))
			// Drain's own Redis round trip failed, not the delivery: the
			// atomic drainScript never ran, so the queue is untouched and
			// its entries are not lost -- but ClaimPending already
			// removed appID from the pending set, so without this,
			// nothing would ever look at that queue again until it next
			// enqueues (or its TTL expires it). Re-add it so the next
			// tick retries.
			if markErr := r.Consumer.MarkPending(ctx, appID, queueTTLFor(r.Interval.Duration())); markErr != nil {
				logger.WithError(markErr).Error(ctx, "failed to re-mark app pending after a failed drain",
					slog.String("app_id", appID))
			}
			continue
		}
		if len(entries) == 0 {
			continue
		}

		err = r.AppContextResolver.ResolveContext(ctx, appID, func(ctx context.Context, appCtx *config.AppContext) error {
			cfg := appCtx.Config
			streams := cfg.AppConfig.Telemetry.AuditLogs.Streams
			disabled := *cfg.FeatureConfig.Telemetry.AuditLogs.Streaming.Disabled
			if disabled || len(streams) == 0 {
				// The batch is already gone -- Drain cleared the queue
				// unconditionally -- so a project that stopped streaming
				// between enqueue and drain simply does not deliver.
				return nil
			}

			sender := r.SenderFactory.MakeSender(appID, appCtx)
			sender.Send(ctx, entries)
			return nil
		})
		if err != nil {
			logger.WithError(err).Error(ctx, "failed to resolve app context for audit log streaming",
				slog.String("app_id", appID))
		}
	}

	return nil
}
