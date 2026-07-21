package analytic

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/analyticredis"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

// firstAuthDedupTTL bounds how long a per-client dedup key lives. It matches the
// 90-day audit retention. A client that re-authenticates after the key expires
// may re-emit application.first_auth, but the deterministic uuid keeps PostHog
// aggregation correct.
const firstAuthDedupTTL = 90 * 24 * time.Hour

var FirstAuthSinkLogger = slogutil.NewLogger("posthog-first-auth-sink")

// FirstAuthSink forwards a single application.first_auth event to PostHog the
// first time each OAuth client authenticates. It is an event.Sink: it runs
// synchronously in the event service's DidCommitTx, but delivery to PostHog is
// detached (fire-and-forget) so it never adds latency to, or fails, the auth.
type FirstAuthSink struct {
	Clock         clock.Clock
	AnalyticRedis *analyticredis.Handle
	Posthog       *PosthogService
}

func (s *FirstAuthSink) ReceiveBlockingEvent(ctx context.Context, e *event.Event) error {
	return nil
}

func (s *FirstAuthSink) ReceiveNonBlockingEvent(ctx context.Context, e *event.Event) error {
	logger := FirstAuthSinkLogger.GetLogger(ctx)

	// Only successful-auth events count as a "first auth".
	if e.Type != nonblocking.UserAuthenticated && e.Type != nonblocking.M2MTokenCreated {
		return nil
	}

	appID := e.Context.AppID
	clientID := e.Context.ClientID
	if appID == "" || clientID == "" {
		return nil
	}

	// Degrade gracefully when PostHog or the analytic Redis is not configured.
	if s.Posthog.PosthogCredentials == nil || s.AnalyticRedis == nil {
		return nil
	}

	firstAuthAt := s.Clock.NowUTC()

	won, err := s.markFirstAuth(ctx, appID, clientID, firstAuthAt)
	if err != nil {
		// Never fail the auth because of an analytics side effect.
		logger.WithError(err).Error(ctx, "failed to mark first auth")
		return nil
	}
	if !won {
		return nil
	}

	// Detach from the request so delivery survives request cancellation, and
	// do not block the auth response on the PostHog round-trip.
	detachedCtx := context.WithoutCancel(ctx)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error(detachedCtx, "panic forwarding first_auth event", slog.Any("recovered", r))
			}
		}()

		evt, err := buildFirstAuthEvent(appID, clientID, firstAuthAt)
		if err != nil {
			logger.WithError(err).Error(detachedCtx, "failed to build first_auth event")
			return
		}
		if err := s.Posthog.Batch(detachedCtx, []json.RawMessage{evt}); err != nil {
			logger.WithError(err).Error(detachedCtx, "failed to forward first_auth to posthog")
		}
	}()

	return nil
}

// markFirstAuth records the first-auth dedup key. It returns true only when the
// key did not already exist, i.e. this is the first auth seen for the client
// within the TTL window.
func (s *FirstAuthSink) markFirstAuth(ctx context.Context, appID string, clientID string, at time.Time) (bool, error) {
	key := firstAuthDedupKey(appID, clientID)
	var keyWasSet bool
	err := s.AnalyticRedis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		var err error
		keyWasSet, err = conn.SetNX(ctx, key, at.UTC().Format(time.RFC3339), firstAuthDedupTTL).Result()
		return err
	})
	if err != nil {
		return false, err
	}
	return keyWasSet, nil
}

func firstAuthDedupKey(appID string, clientID string) string {
	return fmt.Sprintf("app:%s:posthog-first-auth:%s", appID, clientID)
}
