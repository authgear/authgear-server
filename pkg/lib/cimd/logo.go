package cimd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/crypto"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

const (
	// logoCacheTTL. One hour matches RefetchInterval, so a client that
	// changes its logo_uri sees the change take effect on roughly the same
	// timescale as any other metadata change.
	logoCacheTTL = 1 * time.Hour

	// logoNegativeCacheTTL. A failed logo fetch IS negatively cached, unlike
	// a failed document fetch (Service.EnsureClientResolved never caches a
	// document failure). The reasoning differs because the consequence
	// differs: a failed document fetch makes a client unusable, so spec §
	// Error Handling requires immediate retry; a failed logo fetch costs a
	// missing image, and retrying it on every consent-screen render would
	// be a free amplifier pointed at whatever host the document named.
	logoNegativeCacheTTL = 10 * time.Minute

	// singleFlightPurposeLogo namespaces LogoService's single-flight lock,
	// distinct from Service's document-fetch lock (singleFlightPurposeDocument)
	// despite sharing FetchSingleFlight's one implementation and TTL constant.
	singleFlightPurposeLogo = "cimd-logo-fetch"
)

// logoWaitPollInterval and logoWaitMaxDuration bound waitForCachedLogo:
// how long a request that lost the single-flight race waits for the
// winner to finish and populate the cache before giving up.
// logoWaitMaxDuration is comfortably above FetchTimeout (5s) -- long
// enough for a legitimate in-flight fetch to complete -- and comfortably
// below fetchLockTTL (10s), so a holder that died mid-fetch does not make
// every waiter block for the lock's full TTL.
//
// var, not const, and measured with plain time.Now/time.After rather than
// LogoService.Clock: this loop waits on real wall-clock time (an actual
// Redis poll, an actual sleep), and mixing that with an injected clock.Clock
// -- which a test may freeze -- is a proven flakiness trap (see the CIMD
// service tests). Tests instead shrink these vars to keep the real sleep
// negligible.
var (
	logoWaitPollInterval = 100 * time.Millisecond
	logoWaitMaxDuration  = 6 * time.Second
)

var (
	// ErrLogoNegativeCached and ErrLogoFetchInFlight never leave Get --
	// every logo-specific failure collapses to ErrLogoUnavailable() before
	// it returns, matching LogoFetcher's own sentinels (which likewise
	// never leave Get). They exist only to make the concrete cause logged
	// at the point of collapse, the same way Service.EnsureClientResolved
	// logs the concrete document-fetch failure before returning the
	// uniform CIMDUnresolvable.
	ErrLogoNegativeCached = errors.New("cimd: logo fetch previously failed")
	ErrLogoFetchInFlight  = errors.New("cimd: a logo fetch is already in flight")
)

// CIMDLogoUnavailable covers every logo-specific failure: fetch error,
// non-2xx, oversize, disallowed or mismatched content type, invalid
// logo_uri, negative cache hit, a fetch already in flight, and a
// rate-limit refusal. One Kind, one reason -- the client_logo endpoint must
// not report on the reachability of the host the document named.
// apierrors.NotFound is chosen so the Name's own HTTP status (404) already
// matches what the handler writes.
var CIMDLogoUnavailable = apierrors.NotFound.WithReason("CIMDLogoUnavailable")

func ErrLogoUnavailable() error {
	return CIMDLogoUnavailable.New("client logo is unavailable")
}

var LogoServiceLogger = slogutil.NewLogger("cimd-logo-service")

// cachedLogo is the Redis payload. Found distinguishes a cached negative
// result from "not cached at all", which Get's own found-ness handles
// separately. Body is a []byte field, so encoding/json base64-encodes it
// automatically -- no manual encoding needed.
type cachedLogo struct {
	Found       bool      `json:"found"`
	ContentType string    `json:"content_type,omitempty"`
	FetchedAt   time.Time `json:"fetched_at,omitempty"`
	Body        []byte    `json:"body,omitempty"`
	// SourceURI is stored so a logo_uri change on refetch invalidates
	// implicitly: Get compares the caller's logo_uri against this and
	// treats a mismatch as a miss. Without it, a client that changed its
	// logo would keep serving the old image for up to logoCacheTTL after
	// the document refetch already landed.
	SourceURI string `json:"source_uri,omitempty"`
}

func logoCacheKey(appID config.AppID, clientID string) string {
	// clientID is hashed for the same reason redisKeyDynamicClient hashes
	// it: it is an attacker-influenced URL containing ':'.
	return fmt.Sprintf("app:%s:cimd-logo:%s", appID, crypto.SHA256String(clientID))
}

// LogoResult is what Get returns on success.
type LogoResult struct {
	ContentType string
	Body        []byte
	FetchedAt   time.Time
}

// LogoLimiter is the one ratelimit.Limiter method LogoService needs -- the
// same shape as cimd.Limiter (RateLimiter's collaborator), kept as its own
// name so LogoService's dependency is self-describing at its call site.
type LogoLimiter interface {
	Allow(ctx context.Context, spec ratelimit.BucketSpec) (*ratelimit.FailedReservation, error)
}

// LogoService serves a dynamic client's logo, fetching it lazily on first
// request and caching the result in Redis so N pods perform one fetch, not
// N, and so a consent-screen render never waits on a 5-second outbound
// call on anything but the very first request for a given logo.
type LogoService struct {
	Redis        *appredis.Handle
	AppID        config.AppID
	Clock        clock.Clock
	Fetcher      *LogoFetcher
	SingleFlight *FetchSingleFlight
	RateLimiter  LogoLimiter
}

// Get returns the cached logo, fetching it on a miss. Every logo-specific
// failure mode -- negative cache hit, no concurrent fetch finishing in
// time, rate limited, or any LogoFetcher error -- collapses to
// CIMDLogoUnavailable here, the one collapse boundary for this subsystem
// (mirrors Service.EnsureClientResolved's uniform-error rule for the
// document fetch). A Redis or rate-limiter infrastructure error is returned
// UNCHANGED, never collapsed, so a Redis outage surfaces as a 500 rather
// than masquerading as "this client has no logo". (The cache read and the
// single-flight Acquire below are the exception: those two specifically
// degrade rather than fail, each with its own comment explaining why.)
func (s *LogoService) Get(ctx context.Context, clientID string, logoURI string) (*LogoResult, error) {
	logger := LogoServiceLogger.GetLogger(ctx)

	cached, found, err := s.getCached(ctx, clientID)
	if err != nil {
		// A cache failure must never take the endpoint down; fall through
		// to a fetch, same posture as oauthclient.ClientCache. Still worth a
		// warning: it is either a real Redis problem or a corrupt cache
		// entry, and either is worth an operator's attention even though
		// this request itself will succeed.
		logger.WithError(err).
			With(slog.String("client_id", clientID)).
			Warn(ctx, "cimd: failed to read logo cache; fetching directly")
		found = false
	}
	if found && cached.SourceURI == logoURI {
		if cached.Found {
			return &LogoResult{ContentType: cached.ContentType, Body: cached.Body, FetchedAt: cached.FetchedAt}, nil
		}
		logger.WithError(ErrLogoNegativeCached).
			With(slog.String("client_id", clientID)).
			Info(ctx, "cimd: refusing to serve a client logo")
		return nil, ErrLogoUnavailable()
	}
	// found && cached.SourceURI != logoURI falls through as a miss: the
	// client changed its logo_uri on a refetch, so the old cache entry no
	// longer describes what should be served.

	acquired, err := s.SingleFlight.Acquire(ctx, singleFlightPurposeLogo, clientID)
	if err != nil {
		// A Redis failure degrades to a possible stampede, which beats
		// refusing every request -- same posture as Service.EnsureClientResolved.
		// Still logged: silently treating "lock acquisition failed" the same
		// as "lock acquired" hides a real Redis problem from anyone who
		// isn't specifically looking for a stampede.
		logger.WithError(err).
			With(slog.String("client_id", clientID)).
			Warn(ctx, "cimd: failed to acquire logo fetch single-flight lock; proceeding without it")
		acquired = true
	}
	if !acquired {
		// Someone else -- another request on this pod, or another pod
		// entirely -- is already fetching this exact logo. Unlike the
		// document fetch, a cold logo has no stale record to fall back to,
		// so failing immediately here would turn an ordinary fetch race
		// (e.g. two concurrent consent-screen renders for a brand-new
		// client) into a broken image for whichever request lost the race.
		// Wait for the holder to finish and read its result from the cache
		// instead of refusing outright.
		if result, ok := s.waitForCachedLogo(ctx, clientID, logoURI); ok {
			return result, nil
		}
		logger.WithError(ErrLogoFetchInFlight).
			With(slog.String("client_id", clientID)).
			Info(ctx, "cimd: refusing to serve a client logo")
		return nil, ErrLogoUnavailable()
	}

	// Consumed only once a fetch is actually about to happen -- after the
	// cache read and after single-flight acquisition, exactly as in
	// Part 4 §3.3 for the document fetch.
	failedReservation, err := s.RateLimiter.Allow(ctx, NewBucketSpecCIMDLogoPerClient(clientID))
	if err != nil {
		return nil, err
	}
	if rateLimitErr := failedReservation.Error(); rateLimitErr != nil {
		// Deliberately writes NO cache entry: a throttled attempt must not
		// poison the negative cache for logoNegativeCacheTTL.
		logger.Info(ctx, "cimd: logo fetch rate limited", slog.String("client_id", clientID))
		return nil, ErrLogoUnavailable()
	}

	body, contentType, fetchErr := s.Fetcher.Fetch(ctx, logoURI)
	if fetchErr != nil {
		logger.WithError(fetchErr).
			With(slog.String("client_id", clientID)).
			Info(ctx, "cimd: failed to fetch client logo")
		if setErr := s.setCached(ctx, clientID, &cachedLogo{Found: false, SourceURI: logoURI}, logoNegativeCacheTTL); setErr != nil {
			logger.WithError(setErr).Warn(ctx, "cimd: failed to write negative logo cache entry")
		}
		return nil, ErrLogoUnavailable()
	}

	fetchedAt := s.Clock.NowUTC()
	if setErr := s.setCached(ctx, clientID, &cachedLogo{
		Found:       true,
		ContentType: contentType,
		FetchedAt:   fetchedAt,
		Body:        body,
		SourceURI:   logoURI,
	}, logoCacheTTL); setErr != nil {
		// A failed SET must not fail the request -- the result was fetched
		// successfully and is returned regardless; the next request simply
		// fetches again.
		logger.WithError(setErr).Warn(ctx, "cimd: failed to write logo cache entry")
	}

	return &LogoResult{ContentType: contentType, Body: body, FetchedAt: fetchedAt}, nil
}

func (s *LogoService) getCached(ctx context.Context, clientID string) (entry *cachedLogo, found bool, err error) {
	err = s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		data, err := conn.Get(ctx, logoCacheKey(s.AppID, clientID)).Bytes()
		if errors.Is(err, goredis.Nil) {
			found = false
			return nil
		} else if err != nil {
			return err
		}
		var cached cachedLogo
		if err := json.Unmarshal(data, &cached); err != nil {
			return err
		}
		found = true
		entry = &cached
		return nil
	})
	if err != nil {
		return nil, false, err
	}
	return entry, found, nil
}

func (s *LogoService) setCached(ctx context.Context, clientID string, payload *cachedLogo, ttl time.Duration) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		_, err := conn.Set(ctx, logoCacheKey(s.AppID, clientID), data, ttl).Result()
		return err
	})
}

// waitForCachedLogo polls the cache for up to logoWaitMaxDuration, waiting
// for whichever request holds the single-flight lock to finish and write
// its result. ok is false on timeout, on ctx cancellation, or as soon as a
// negative cache entry shows the holder's fetch already failed -- there is
// no point waiting out the rest of the deadline for an answer that has
// already arrived. A transient cache-read error while waiting is treated
// as "not yet" and retried rather than given up on immediately, since the
// point of waiting at all is to tolerate exactly this kind of blip.
func (s *LogoService) waitForCachedLogo(ctx context.Context, clientID string, logoURI string) (result *LogoResult, ok bool) {
	logger := LogoServiceLogger.GetLogger(ctx)
	loggedReadError := false
	deadline := time.Now().Add(logoWaitMaxDuration)
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return nil, false
		case <-time.After(logoWaitPollInterval):
		}

		cached, found, err := s.getCached(ctx, clientID)
		if err != nil {
			if !loggedReadError {
				// Logged once per wait, not once per poll: at
				// logoWaitPollInterval this loop can run dozens of times
				// before its deadline, and a sustained Redis outage would
				// otherwise turn one failed wait into a log flood.
				logger.WithError(err).
					With(slog.String("client_id", clientID)).
					Warn(ctx, "cimd: failed to read logo cache while waiting for an in-flight fetch")
				loggedReadError = true
			}
			continue
		}
		if !found || cached.SourceURI != logoURI {
			continue
		}
		if !cached.Found {
			return nil, false
		}
		return &LogoResult{ContentType: cached.ContentType, Body: cached.Body, FetchedAt: cached.FetchedAt}, true
	}
	return nil, false
}

// NewBucketSpecCIMDLogoPerClient is a fixed, non-configurable bucket, unlike
// the document-fetch buckets: a logo fetch requires a persisted client
// record, and records are only created/refreshed by a document fetch, which
// is already limited per project and per IP and capped by the
// oauth_client_cimd quota. What is NOT already bounded is repeated misses
// for a SINGLE client if cache writes are failing or being evicted -- a
// per-client_id concern, so it gets a per-client_id bucket and nothing
// else. Burst 2/minute is deliberately strict: a legitimate cache miss
// needs exactly one fetch, and the negative cache then suppresses retries
// for ten minutes, so 2/minute already allows a retry that legitimate
// traffic never needs.
func NewBucketSpecCIMDLogoPerClient(clientID string) ratelimit.BucketSpec {
	return ratelimit.NewBucketSpec(
		ratelimit.RateLimitOAuthCIMDLogoPerClient,
		ratelimit.RateLimitGroupOAuthCIMDLogo,
		&config.RateLimitConfig{
			Enabled: new(true),
			Period:  "1m",
			Burst:   2,
		},
		ratelimit.OAuthCIMDLogoPerClient,
		// Hashed, unlike Part 4's IP argument: BucketSpec.Key() joins
		// arguments with ":", and a client_id is a URL containing ":", so
		// passing it raw would let one client_id's key collide with
		// another's namespace -- the same reason redisKeyDynamicClient
		// hashes it.
		crypto.SHA256String(clientID),
	)
}
