package auditlogstreaming

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/globalredis"
)

const (
	// maxQueueLength bounds a single project's queue. Beyond it the oldest
	// queued entry is dropped.
	maxQueueLength = 10000

	// minQueueTTL floors queueTTLFor so a short configured interval cannot
	// make a queue expire faster than an ordinary background worker
	// restart/deploy -- it must survive the worker being gone for a while,
	// not just one missed tick.
	minQueueTTL = 5 * time.Minute

	// queueTTLSafetyFactor is how many drain intervals a queue survives
	// undrained before it expires. Generous on purpose: a missed tick
	// (lock contention, a slow previous tick, see runnable.go's own 2x on
	// the drain lock's expiry) must not lose data, and holding entries a
	// little longer costs Redis memory, not correctness.
	queueTTLSafetyFactor = 5
)

// queueTTLFor derives the queue TTL from the configured drain interval,
// rather than using one fixed duration regardless of it: a TTL shorter
// than (or too close to) the interval silently discards every entry that
// arrives just before a gap between drains, with no error anywhere --
// the tick that would have drained it simply finds nothing there. At the
// default 1-minute interval this returns exactly the 5-minute floor, so
// existing deployments see no change.
func queueTTLFor(interval time.Duration) time.Duration {
	if t := queueTTLSafetyFactor * interval; t > minQueueTTL {
		return t
	}
	return minQueueTTL
}

// Both keys are in global Redis: the drainer is a single global worker, and
// the keys are not app-scoped in the "app:<id>:" sense. This follows
// RedisKeyForQueue in pkg/lib/infra/redisqueue/task.go, likewise a global
// key with the app ID carried in the payload rather than the key.
func redisKeyPending() string { return "audit-log-stream:pending" }
func redisKeyQueue(appID string) string {
	return fmt.Sprintf("audit-log-stream:queue:%v", appID)
}

// Redis_6_0_Cmdable (pkg/lib/infra/redis) has no LRange, LTrim, RPush or
// LPop, so all queue access goes through Lua, exactly like
// pkg/lib/ratelimit/gcra.go, pkg/lib/lockout/record.go and
// pkg/lib/usage/limit.go already do. Running through a script also makes
// each operation atomic.

// enqueueScript appends an entry, trims the queue to maxQueueLength
// (dropping the oldest first), and refreshes both keys' TTL.
//
// KEYS[1] = queue key, KEYS[2] = pending key
// ARGV[1] = entry JSON, ARGV[2] = max length, ARGV[3] = TTL seconds, ARGV[4] = app ID
var enqueueScript = goredis.NewScript(`
redis.call("RPUSH", KEYS[1], ARGV[1])
redis.call("LTRIM", KEYS[1], -tonumber(ARGV[2]), -1)
redis.call("EXPIRE", KEYS[1], tonumber(ARGV[3]))
redis.call("SADD", KEYS[2], ARGV[4])
redis.call("EXPIRE", KEYS[2], tonumber(ARGV[3]))
return redis.status_reply("OK")
`)

// claimPendingScript takes and clears the set of app IDs with a
// non-empty queue. An app that enqueues again after being claimed is
// simply re-added (SADD runs on every enqueue) and picked up on the next
// claim.
//
// KEYS[1] = pending key
var claimPendingScript = goredis.NewScript(`
local apps = redis.call("SMEMBERS", KEYS[1])
redis.call("DEL", KEYS[1])
return apps
`)

// drainScript takes and clears one app's queue in one step. Taking and
// clearing together is what implements the spec's "taking the batch
// clears the queue, whether or not the send succeeds": a failure after
// this script runs drops the batch by construction, there is nothing left
// to retry from. A failure running this script itself is a different
// case -- the script never executed, so the queue is untouched -- see
// markPendingScript.
//
// KEYS[1] = queue key
var drainScript = goredis.NewScript(`
local items = redis.call("LRANGE", KEYS[1], 0, -1)
redis.call("DEL", KEYS[1])
return items
`)

// markPendingScript re-adds an app to the pending set. Consumer.Drain
// calls this when drainScript itself fails to run (a Redis round-trip
// failure, not a delivery failure): ClaimPending already removed the app
// from the pending set, but drainScript never ran, so the app's queue is
// untouched -- only the "check this app next tick" bookkeeping was lost.
// Re-adding it is what makes that queue reachable again, exactly as if
// the app had just enqueued something.
//
// KEYS[1] = pending key
// ARGV[1] = app ID, ARGV[2] = TTL seconds
var markPendingScript = goredis.NewScript(`
redis.call("SADD", KEYS[1], ARGV[1])
redis.call("EXPIRE", KEYS[1], tonumber(ARGV[2]))
return redis.status_reply("OK")
`)

// Producer enqueues persisted audit log entries for streaming. It is
// request-scoped and holds only pointers into the resolved config.
type Producer struct {
	AppID    config.AppID
	Streams  []*config.TelemetryAuditLogStreamConfig
	Disabled bool
	Redis    *globalredis.Handle
	// Interval is the background worker's configured drain interval
	// (AUDIT_LOG_STREAMING_INTERVAL), used to derive the queue TTL -- see
	// queueTTLFor.
	Interval config.DurationString
}

// Enqueue never returns an error and never blocks on anything but the
// Redis round trip: streaming can never fail an end-user request.
func (p *Producer) Enqueue(ctx context.Context, e *event.Event) {
	// A project with no streams, or with streaming disabled, performs no
	// Redis work at all -- this is what keeps the cost proportional to
	// streaming-enabled projects rather than to all audit traffic.
	if p.Disabled || len(p.Streams) == 0 {
		return
	}

	logger := Logger.GetLogger(ctx)

	payload, err := json.Marshal(e)
	if err != nil {
		logger.WithError(err).Error(ctx, "failed to marshal audit log entry for streaming",
			slog.String("app_id", string(p.AppID)))
		return
	}

	ttl := queueTTLFor(p.Interval.Duration())
	err = p.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		return enqueueScript.Run(ctx, conn,
			[]string{redisKeyQueue(string(p.AppID)), redisKeyPending()},
			string(payload), maxQueueLength, int(ttl.Seconds()), string(p.AppID),
		).Err()
	})
	if err != nil {
		logger.WithError(err).Error(ctx, "failed to enqueue audit log entry for streaming",
			slog.String("app_id", string(p.AppID)))
	}
}

// Consumer drains the queues on the background worker's behalf.
type Consumer struct {
	Redis *globalredis.Handle
}

// ClaimPending returns every app ID with a non-empty queue, and clears the
// pending set in the same script.
func (c *Consumer) ClaimPending(ctx context.Context) ([]string, error) {
	var appIDs []string
	err := c.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		var err error
		appIDs, err = claimPendingScript.Run(ctx, conn, []string{redisKeyPending()}).StringSlice()
		return err
	})
	if err != nil {
		return nil, err
	}
	return appIDs, nil
}

// MarkPending re-adds appID to the pending set. Callers use this when
// Drain fails: ClaimPending already removed appID from the pending set,
// but if Drain's own Redis round trip then fails, drainScript never ran,
// so the app's queue is untouched -- only the "check this app"
// bookkeeping was lost. ttl should be the same TTL Enqueue would use
// (queueTTLFor(interval)), so a re-marked app does not become reachable
// for longer than an app that enqueued normally.
func (c *Consumer) MarkPending(ctx context.Context, appID string, ttl time.Duration) error {
	return c.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		return markPendingScript.Run(ctx, conn, []string{redisKeyPending()}, appID, int(ttl.Seconds())).Err()
	})
}

// Drain takes and clears one app's queue, decoding each entry.
func (c *Consumer) Drain(ctx context.Context, appID string) ([]QueuedEntry, error) {
	var raws []string
	err := c.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		var err error
		raws, err = drainScript.Run(ctx, conn, []string{redisKeyQueue(appID)}).StringSlice()
		return err
	})
	if err != nil {
		return nil, err
	}

	logger := Logger.GetLogger(ctx)
	entries := make([]QueuedEntry, 0, len(raws))
	for _, raw := range raws {
		entry, err := decodeQueuedEntry([]byte(raw))
		if err != nil {
			// The script above already popped and cleared this entry from
			// Redis, so there is nothing left to retry it from. A decode
			// failure on one entry must not cost every other entry in the
			// same batch: log and skip just this one, keep draining the rest.
			logger.WithError(err).Error(ctx, "failed to decode queued audit log entry, dropping it",
				slog.String("app_id", appID))
			continue
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

// queuedEventShadow mirrors event.Event's JSON shape for exactly the
// fields the encoder reads. It deliberately has no Payload field:
// event.Payload is an interface with methods, and encoding/json cannot
// decode a stored JSON object into an interface field that has no current
// concrete value -- confirmed to fail with "cannot unmarshal object into
// Go struct field Event.payload of type event.Payload" -- so any struct
// used to decode a queued entry must omit it rather than declare it.
type queuedEventShadow struct {
	ID      string        `json:"id"`
	Type    event.Type    `json:"type"`
	Context event.Context `json:"context"`
}

func decodeQueuedEntry(raw []byte) (QueuedEntry, error) {
	var shadow queuedEventShadow
	if err := json.Unmarshal(raw, &shadow); err != nil {
		return QueuedEntry{}, err
	}
	return QueuedEntry{
		Event: &event.Event{
			ID:      shadow.ID,
			Type:    shadow.Type,
			Context: shadow.Context,
		},
		Raw: raw,
	}, nil
}
