package cimd

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alicebob/miniredis/v2"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

func newTestRedisHandle(t *testing.T) (*appredis.Handle, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	pool := redis.NewPool()
	rh := redis.NewHandle(pool, redis.ConnectionOptions{
		RedisURL:              "redis://" + mr.Addr(),
		MaxOpenConnection:     func(i int) *int { return &i }(10),
		MaxIdleConnection:     func(i int) *int { return &i }(5),
		IdleConnectionTimeout: func(d config.DurationSeconds) *config.DurationSeconds { return &d }(300),
		MaxConnectionLifetime: func(d config.DurationSeconds) *config.DurationSeconds { return &d }(900),
	})
	return &appredis.Handle{Handle: rh}, mr
}

// stubLogoLimiter is a hand-rolled LogoLimiter: it records every call and
// returns a canned result.
type stubLogoLimiter struct {
	failWith *ratelimit.FailedReservation
	err      error
	calls    int
}

func (s *stubLogoLimiter) Allow(ctx context.Context, spec ratelimit.BucketSpec) (*ratelimit.FailedReservation, error) {
	s.calls++
	if s.err != nil {
		return nil, s.err
	}
	return s.failWith, nil
}

func newTestLogoService(t *testing.T, fetchHandler http.HandlerFunc) (*LogoService, *documentServerForLogoTest) {
	t.Helper()
	rh, _ := newTestRedisHandle(t)
	srv := &documentServerForLogoTest{}
	srv.Server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		srv.hits++
		fetchHandler(w, r)
	}))
	return &LogoService{
		Redis:        rh,
		AppID:        "test-app",
		Clock:        clock.NewMockClockAt("2026-08-31T12:00:00Z"),
		Fetcher:      logoFetcherFor(newLoopbackHTTPClient(certPool(srv.Server), nil), nil),
		SingleFlight: &FetchSingleFlight{Redis: rh, AppID: "test-app"},
		RateLimiter:  &stubLogoLimiter{},
	}, srv
}

type documentServerForLogoTest struct {
	*httptest.Server
	hits int
}

func TestLogoServiceGet(t *testing.T) {
	Convey("LogoService.Get", t, func() {
		ctx := context.Background()
		clientID := "https://mcp-client.example.com/oauth/client-metadata.json"

		Convey("cache miss: fetches, and writes a positive entry with SourceURI", func() {
			svc, srv := newTestLogoService(t, func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "image/png")
				_, _ = w.Write(pngMagic)
			})
			defer srv.Close()

			result, err := svc.Get(ctx, clientID, srv.URL)
			So(err, ShouldBeNil)
			So(result.Body, ShouldResemble, pngMagic)
			So(result.ContentType, ShouldEqual, "image/png")
			So(srv.hits, ShouldEqual, 1)

			cached, found, err := svc.getCached(ctx, clientID)
			So(err, ShouldBeNil)
			So(found, ShouldBeTrue)
			So(cached.Found, ShouldBeTrue)
			So(cached.SourceURI, ShouldEqual, srv.URL)
		})

		Convey("cache hit with matching SourceURI: the fetcher is not called", func() {
			svc, srv := newTestLogoService(t, func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "image/png")
				_, _ = w.Write(pngMagic)
			})
			defer srv.Close()

			_, err := svc.Get(ctx, clientID, srv.URL)
			So(err, ShouldBeNil)
			So(srv.hits, ShouldEqual, 1)

			result, err := svc.Get(ctx, clientID, srv.URL)
			So(err, ShouldBeNil)
			So(result.Body, ShouldResemble, pngMagic)
			So(srv.hits, ShouldEqual, 1) // still 1: the second Get was a cache hit
		})

		Convey("cache hit with a DIFFERENT SourceURI (the client changed its logo): treated as a miss and refetched", func() {
			svc, srv := newTestLogoService(t, func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "image/png")
				_, _ = w.Write(pngMagic)
			})
			defer srv.Close()

			_, err := svc.Get(ctx, clientID, srv.URL+"/old-logo.png")
			So(err, ShouldBeNil)
			So(srv.hits, ShouldEqual, 1)

			// A single-flight lock backed by a separate Redis instance,
			// standing in for the first attempt's lock having since
			// expired (the lock lives IN Redis, so a fresh *FetchSingleFlight
			// struct sharing the same Redis handle would still see it) --
			// this test is about the SourceURI-mismatch/cache-invalidation
			// behavior, not single-flight collapsing. The cache itself
			// (svc.Redis) is untouched, so cache reads still see what the
			// first Get wrote.
			freshRedis, _ := newTestRedisHandle(t)
			svc.SingleFlight = &FetchSingleFlight{Redis: freshRedis, AppID: svc.AppID}
			_, err = svc.Get(ctx, clientID, srv.URL+"/new-logo.png")
			So(err, ShouldBeNil)
			So(srv.hits, ShouldEqual, 2)
		})

		Convey("cache hit with Found: false and matching SourceURI: CIMDLogoUnavailable, fetcher not called", func() {
			svc, srv := newTestLogoService(t, func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/html")
				_, _ = w.Write(htmlBody)
			})
			defer srv.Close()

			_, err := svc.Get(ctx, clientID, srv.URL)
			So(apierrors.IsKind(err, CIMDLogoUnavailable), ShouldBeTrue)
			So(srv.hits, ShouldEqual, 1)

			_, err = svc.Get(ctx, clientID, srv.URL)
			So(apierrors.IsKind(err, CIMDLogoUnavailable), ShouldBeTrue)
			So(srv.hits, ShouldEqual, 1) // second attempt served from the negative cache
		})

		Convey("fetch failure: CIMDLogoUnavailable, and a negative entry is written", func() {
			svc, srv := newTestLogoService(t, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			})
			defer srv.Close()

			_, err := svc.Get(ctx, clientID, srv.URL)
			So(apierrors.IsKind(err, CIMDLogoUnavailable), ShouldBeTrue)

			cached, found, err := svc.getCached(ctx, clientID)
			So(err, ShouldBeNil)
			So(found, ShouldBeTrue)
			So(cached.Found, ShouldBeFalse)
			So(cached.SourceURI, ShouldEqual, srv.URL)
		})

		Convey("single-flight not acquired: CIMDLogoUnavailable, fetcher not called", func() {
			svc, srv := newTestLogoService(t, func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "image/png")
				_, _ = w.Write(pngMagic)
			})
			defer srv.Close()

			acquired, err := svc.SingleFlight.Acquire(ctx, singleFlightPurposeLogo, clientID)
			So(err, ShouldBeNil)
			So(acquired, ShouldBeTrue) // consume the lock so Get's own Acquire fails

			_, err = svc.Get(ctx, clientID, srv.URL)
			So(apierrors.IsKind(err, CIMDLogoUnavailable), ShouldBeTrue)
			So(srv.hits, ShouldEqual, 0)
		})

		Convey("Redis GET error (connection down): fetch proceeds; a failed SET does not fail the request", func() {
			rh, mr := newTestRedisHandle(t)
			srv := &documentServerForLogoTest{}
			srv.Server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				srv.hits++
				w.Header().Set("Content-Type", "image/png")
				_, _ = w.Write(pngMagic)
			}))
			defer srv.Close()
			mr.Close() // Redis is now unreachable for both GET and SET

			svc := &LogoService{
				Redis:        rh,
				AppID:        "test-app",
				Clock:        clock.NewMockClockAt("2026-08-31T12:00:00Z"),
				Fetcher:      logoFetcherFor(newLoopbackHTTPClient(certPool(srv.Server), nil), nil),
				SingleFlight: &FetchSingleFlight{Redis: rh, AppID: "test-app"},
				RateLimiter:  &stubLogoLimiter{},
			}

			result, err := svc.Get(ctx, clientID, srv.URL)
			So(err, ShouldBeNil)
			So(result.Body, ShouldResemble, pngMagic)
			So(srv.hits, ShouldEqual, 1)
		})

		Convey("rate limiter refuses: CIMDLogoUnavailable, fetcher not called, no cache entry written", func() {
			svc, srv := newTestLogoService(t, func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "image/png")
				_, _ = w.Write(pngMagic)
			})
			defer srv.Close()
			spec := NewBucketSpecCIMDLogoPerClient(clientID)
			limiter := &stubLogoLimiter{failWith: ratelimit.NewFailedReservation(spec)}
			svc.RateLimiter = limiter

			_, err := svc.Get(ctx, clientID, srv.URL)
			So(apierrors.IsKind(err, CIMDLogoUnavailable), ShouldBeTrue)
			So(srv.hits, ShouldEqual, 0)
			So(limiter.calls, ShouldEqual, 1)

			_, found, err := svc.getCached(ctx, clientID)
			So(err, ShouldBeNil)
			So(found, ShouldBeFalse)
		})

		Convey("a rate limiter infrastructure error is returned UNCHANGED, not as CIMDLogoUnavailable", func() {
			svc, srv := newTestLogoService(t, func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "image/png")
				_, _ = w.Write(pngMagic)
			})
			defer srv.Close()
			infraErr := errors.New("redis: connection refused")
			svc.RateLimiter = &stubLogoLimiter{err: infraErr}

			_, err := svc.Get(ctx, clientID, srv.URL)
			So(errors.Is(err, infraErr), ShouldBeTrue)
			So(apierrors.IsKind(err, CIMDLogoUnavailable), ShouldBeFalse)
			So(srv.hits, ShouldEqual, 0)
		})
	})
}
