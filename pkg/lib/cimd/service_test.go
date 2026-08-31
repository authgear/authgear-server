package cimd_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/cimd"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/lib/oauthclient"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

const testClientID = "https://mcp-client.example.com/oauth/client-metadata.json"

var testDocument = []byte(`{
  "client_id": "https://mcp-client.example.com/oauth/client-metadata.json",
  "client_name": "Example MCP Client",
  "redirect_uris": ["http://127.0.0.1:3000/callback"],
  "grant_types": ["authorization_code", "refresh_token"],
  "response_types": ["code"],
  "token_endpoint_auth_method": "none"
}`)

// stubCommands / stubQueries / stubDatabase are hand-rolled stubs (no
// generated mocks in this package) since the test matrix mostly needs call
// counting and canned return values, not strict call-order assertions.

type stubCommands struct {
	upsertFn    func(*oauthclient.UpsertCIMDClientOptions) (*oauthclient.Client, bool, error)
	upsertCalls atomic.Int64
}

func (s *stubCommands) UpsertCIMDClient(ctx context.Context, options *oauthclient.UpsertCIMDClientOptions) (*oauthclient.Client, bool, error) {
	s.upsertCalls.Add(1)
	if s.upsertFn != nil {
		return s.upsertFn(options)
	}
	return &oauthclient.Client{ClientID: options.ClientID, Source: model.OAuthClientSourceCIMD}, true, nil
}

func (s *stubCommands) LockForClientCount(ctx context.Context, source oauthclient.Source) error {
	return nil
}

func (s *stubCommands) CountClientsBySource(ctx context.Context, source model.OAuthClientSource) (uint64, error) {
	return 0, nil
}

type stubQueries struct {
	getFn func() (*oauthclient.Client, error)
	calls atomic.Int64
}

func (s *stubQueries) GetClientByClientID(ctx context.Context, clientID string) (*oauthclient.Client, error) {
	s.calls.Add(1)
	if s.getFn != nil {
		return s.getFn()
	}
	return nil, oauthclient.ErrDynamicClientNotFound
}

type stubDatabase struct {
	inTx        bool
	withTxCalls atomic.Int64
}

func (s *stubDatabase) WithTx(ctx context.Context, do func(ctx context.Context) error) error {
	s.withTxCalls.Add(1)
	return do(ctx)
}

func (s *stubDatabase) IsInTx(ctx context.Context) bool { return s.inTx }

// documentServer is an httptest.Server (TLS, trusted via its own
// srv.Client()) that counts every request it receives, so "fetcher never
// called" is a direct assertion rather than an inference.
type documentServer struct {
	*httptest.Server
	hits atomic.Int64
}

func newDocumentServer(handler http.HandlerFunc) *documentServer {
	ds := &documentServer{}
	ds.Server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ds.hits.Add(1)
		handler(w, r)
	}))
	return ds
}

// fetcherFor builds a Fetcher whose HTTP clients are ds's own trusting
// client (httptest.Server.Client() already trusts that server's
// certificate). Address filtering is Part 2's own test suite's concern,
// not Service's -- these tests exercise Service's branching, not SSRF
// policy.
func fetcherFor(ds *documentServer) *cimd.Fetcher {
	return &cimd.Fetcher{
		HTTPClients:        &cimd.CIMDHTTPClients{Strict: ds.Client(), Insecure: ds.Client()},
		OAuthFeatureConfig: &config.OAuthFeatureConfig{},
		AppID:              "test-app",
	}
}

func newTestSingleFlight(t *testing.T) (*cimd.FetchSingleFlight, *miniredis.Miniredis) {
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
	return &cimd.FetchSingleFlight{Redis: &appredis.Handle{Handle: rh}, AppID: "test-app"}, mr
}

// newWorkingSingleFlight is used wherever the test doesn't care about
// single-flight behavior specifically but EnsureClientResolved's step (5)
// still calls Acquire on its way to a fetch -- the zero-value
// FetchSingleFlight has a nil Redis handle and panics.
func newWorkingSingleFlight(t *testing.T) *cimd.FetchSingleFlight {
	sf, _ := newTestSingleFlight(t)
	return sf
}

func enabledOAuthConfig() *config.OAuthConfig {
	return &config.OAuthConfig{
		ClientIDMetadataDocument: &config.OAuthClientIDMetadataDocumentConfig{Enabled: true},
	}
}

func TestServiceEnsureClientResolved(t *testing.T) {
	ctx := context.Background()

	Convey("Service.EnsureClientResolved", t, func() {
		Convey("CIMD disabled: nil, fetcher and Queries never called", func() {
			ds := newDocumentServer(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write(testDocument) })
			defer ds.Close()
			queries := &stubQueries{}
			svc := &cimd.Service{
				OAuthConfig:  &config.OAuthConfig{ClientIDMetadataDocument: &config.OAuthClientIDMetadataDocumentConfig{Enabled: false}},
				Fetcher:      fetcherFor(ds),
				Queries:      queries,
				Database:     &stubDatabase{},
				RateLimiter:  cimd.ProvideNoopServiceRateLimiter(),
				UsageLimiter: cimd.ProvideNoopServiceUsageLimiter(),
				SingleFlight: newWorkingSingleFlight(t),
			}
			err := svc.EnsureClientResolved(ctx, testClientID)
			So(err, ShouldBeNil)
			So(ds.hits.Load(), ShouldEqual, int64(0))
			So(queries.calls.Load(), ShouldEqual, int64(0))
		})

		for _, id := range []string{"dcrc_x", "my-client", ""} {
			id := id
			Convey("non-URL client_id ("+id+"): nil, fetcher never called", func() {
				ds := newDocumentServer(func(w http.ResponseWriter, r *http.Request) {})
				defer ds.Close()
				svc := &cimd.Service{
					OAuthConfig:  enabledOAuthConfig(),
					Fetcher:      fetcherFor(ds),
					Queries:      &stubQueries{},
					Database:     &stubDatabase{},
					RateLimiter:  cimd.ProvideNoopServiceRateLimiter(),
					UsageLimiter: cimd.ProvideNoopServiceUsageLimiter(),
					SingleFlight: newWorkingSingleFlight(t),
				}
				err := svc.EnsureClientResolved(ctx, id)
				So(err, ShouldBeNil)
				So(ds.hits.Load(), ShouldEqual, int64(0))
			})
		}

		Convey("client_id matches a static client: nil, fetcher never called", func() {
			ds := newDocumentServer(func(w http.ResponseWriter, r *http.Request) {})
			defer ds.Close()
			oauthConfig := enabledOAuthConfig()
			oauthConfig.Clients = []config.OAuthClientConfig{{ClientID: testClientID}}
			svc := &cimd.Service{
				OAuthConfig:  oauthConfig,
				Fetcher:      fetcherFor(ds),
				Queries:      &stubQueries{},
				Database:     &stubDatabase{},
				RateLimiter:  cimd.ProvideNoopServiceRateLimiter(),
				UsageLimiter: cimd.ProvideNoopServiceUsageLimiter(),
				SingleFlight: newWorkingSingleFlight(t),
			}
			err := svc.EnsureClientResolved(ctx, testClientID)
			So(err, ShouldBeNil)
			So(ds.hits.Load(), ShouldEqual, int64(0))
		})

		Convey("host fails allowed_domains: CIMDUnresolvable, fetcher never called", func() {
			ds := newDocumentServer(func(w http.ResponseWriter, r *http.Request) {})
			defer ds.Close()
			oauthConfig := enabledOAuthConfig()
			oauthConfig.ClientIDMetadataDocument.AllowedDomains = []string{"only-this.example.com"}
			svc := &cimd.Service{
				OAuthConfig:  oauthConfig,
				Fetcher:      fetcherFor(ds),
				Queries:      &stubQueries{},
				Database:     &stubDatabase{},
				RateLimiter:  cimd.ProvideNoopServiceRateLimiter(),
				UsageLimiter: cimd.ProvideNoopServiceUsageLimiter(),
				SingleFlight: newWorkingSingleFlight(t),
			}
			err := svc.EnsureClientResolved(ctx, testClientID)
			So(apierrors.IsKind(err, cimd.CIMDUnresolvable), ShouldBeTrue)
			So(ds.hits.Load(), ShouldEqual, int64(0))
		})

		Convey("no record, fetch+validate ok: UpsertCIMDClient called once, options match the document, created observed", func() {
			var ds *documentServer
			ds = newDocumentServer(func(w http.ResponseWriter, r *http.Request) {
				doc := `{"client_id":"` + ds.URL + `/x","redirect_uris":["http://127.0.0.1:3000/callback"],"grant_types":["authorization_code","refresh_token"],"response_types":["code"]}`
				_, _ = w.Write([]byte(doc))
			})
			defer ds.Close()

			var gotOptions *oauthclient.UpsertCIMDClientOptions
			var gotCreated bool
			commands := &stubCommands{upsertFn: func(o *oauthclient.UpsertCIMDClientOptions) (*oauthclient.Client, bool, error) {
				gotOptions = o
				gotCreated = true
				return &oauthclient.Client{ClientID: o.ClientID, Source: model.OAuthClientSourceCIMD}, true, nil
			}}
			svc := &cimd.Service{
				OAuthConfig:  enabledOAuthConfig(),
				Fetcher:      fetcherFor(ds),
				Commands:     commands,
				Queries:      &stubQueries{},
				Database:     &stubDatabase{},
				RateLimiter:  cimd.ProvideNoopServiceRateLimiter(),
				UsageLimiter: cimd.ProvideNoopServiceUsageLimiter(),
				SingleFlight: newWorkingSingleFlight(t),
			}

			clientID := ds.URL + "/x"
			err := svc.EnsureClientResolved(ctx, clientID)
			So(err, ShouldBeNil)
			So(commands.upsertCalls.Load(), ShouldEqual, int64(1))
			So(gotCreated, ShouldBeTrue)
			So(gotOptions.ClientID, ShouldEqual, clientID)
			So(gotOptions.RedirectURIs, ShouldResemble, []string{"http://127.0.0.1:3000/callback"})
			So(gotOptions.GrantTypes, ShouldResemble, []string{"authorization_code", "refresh_token"})
			So(gotOptions.ResponseTypes, ShouldResemble, []string{"code"})
		})

		Convey("record exists, LastFetchedAt 30 minutes ago: nil, fetcher never called", func() {
			mockClock := clock.NewMockClockAt("2026-08-31T12:00:00Z")
			fetchedAt := mockClock.NowUTC().Add(-30 * time.Minute)
			ds := newDocumentServer(func(w http.ResponseWriter, r *http.Request) {})
			defer ds.Close()
			queries := &stubQueries{getFn: func() (*oauthclient.Client, error) {
				return &oauthclient.Client{ClientID: testClientID, Source: model.OAuthClientSourceCIMD, LastFetchedAt: &fetchedAt}, nil
			}}
			svc := &cimd.Service{
				OAuthConfig:  enabledOAuthConfig(),
				Clock:        mockClock,
				Fetcher:      fetcherFor(ds),
				Queries:      queries,
				Database:     &stubDatabase{},
				RateLimiter:  cimd.ProvideNoopServiceRateLimiter(),
				UsageLimiter: cimd.ProvideNoopServiceUsageLimiter(),
				SingleFlight: newWorkingSingleFlight(t),
			}
			err := svc.EnsureClientResolved(ctx, testClientID)
			So(err, ShouldBeNil)
			So(ds.hits.Load(), ShouldEqual, int64(0))
		})

		Convey("record exists, LastFetchedAt 61 minutes ago, fetch ok: upsert called, created == false", func() {
			mockClock := clock.NewMockClockAt("2026-08-31T12:00:00Z")
			fetchedAt := mockClock.NowUTC().Add(-61 * time.Minute)
			var ds *documentServer
			ds = newDocumentServer(func(w http.ResponseWriter, r *http.Request) {
				doc := `{"client_id":"` + ds.URL + `/x","redirect_uris":["http://127.0.0.1:3000/callback"]}`
				_, _ = w.Write([]byte(doc))
			})
			defer ds.Close()
			queries := &stubQueries{getFn: func() (*oauthclient.Client, error) {
				return &oauthclient.Client{ClientID: ds.URL + "/x", Source: model.OAuthClientSourceCIMD, LastFetchedAt: &fetchedAt}, nil
			}}
			commands := &stubCommands{upsertFn: func(o *oauthclient.UpsertCIMDClientOptions) (*oauthclient.Client, bool, error) {
				return &oauthclient.Client{ClientID: o.ClientID, Source: model.OAuthClientSourceCIMD}, false, nil
			}}
			svc := &cimd.Service{
				OAuthConfig:  enabledOAuthConfig(),
				Clock:        mockClock,
				Fetcher:      fetcherFor(ds),
				Commands:     commands,
				Queries:      queries,
				Database:     &stubDatabase{},
				RateLimiter:  cimd.ProvideNoopServiceRateLimiter(),
				UsageLimiter: cimd.ProvideNoopServiceUsageLimiter(),
				SingleFlight: newWorkingSingleFlight(t),
			}
			err := svc.EnsureClientResolved(ctx, ds.URL+"/x")
			So(err, ShouldBeNil)
			So(commands.upsertCalls.Load(), ShouldEqual, int64(1))
			So(ds.hits.Load(), ShouldEqual, int64(1))
		})

		Convey("record exists, LastFetchedAt NULL: fetch attempted (never fresh)", func() {
			ds := newDocumentServer(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			})
			defer ds.Close()
			queries := &stubQueries{getFn: func() (*oauthclient.Client, error) {
				return &oauthclient.Client{ClientID: testClientID, Source: model.OAuthClientSourceCIMD, LastFetchedAt: nil}, nil
			}}
			svc := &cimd.Service{
				OAuthConfig:  enabledOAuthConfig(),
				Clock:        clock.NewMockClockAt("2026-08-31T12:00:00Z"),
				Fetcher:      fetcherFor(ds),
				Commands:     &stubCommands{},
				Queries:      queries,
				Database:     &stubDatabase{},
				RateLimiter:  cimd.ProvideNoopServiceRateLimiter(),
				UsageLimiter: cimd.ProvideNoopServiceUsageLimiter(),
				SingleFlight: newWorkingSingleFlight(t),
			}
			// A NULL-LastFetchedAt record with a failing refetch serves stale (nil).
			err := svc.EnsureClientResolved(ctx, testClientID)
			So(err, ShouldBeNil)
			So(ds.hits.Load(), ShouldEqual, int64(1))
		})

		Convey("record exists, fetch fails (multiple modes): stale served, upsert not called", func() {
			cases := map[string]http.HandlerFunc{
				"404":          func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNotFound) },
				"invalid json": func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("not json")) },
				"client_id mismatch": func(w http.ResponseWriter, r *http.Request) {
					_, _ = w.Write([]byte(`{"client_id":"https://wrong.example.com/x","redirect_uris":["https://x/cb"]}`))
				},
				"oversize": func(w http.ResponseWriter, r *http.Request) {
					_, _ = w.Write(make([]byte, cimd.MaxDocumentBytes+1))
				},
			}
			for name, handler := range cases {
				name, handler := name, handler
				Convey(name, func() {
					ds := newDocumentServer(handler)
					defer ds.Close()
					fetchedAt := time.Now().Add(-2 * time.Hour)
					queries := &stubQueries{getFn: func() (*oauthclient.Client, error) {
						return &oauthclient.Client{ClientID: testClientID, Source: model.OAuthClientSourceCIMD, LastFetchedAt: &fetchedAt}, nil
					}}
					commands := &stubCommands{}
					svc := &cimd.Service{
						OAuthConfig:  enabledOAuthConfig(),
						Clock:        clock.NewMockClockAt("2026-08-31T12:00:00Z"),
						Fetcher:      fetcherFor(ds),
						Commands:     commands,
						Queries:      queries,
						Database:     &stubDatabase{},
						RateLimiter:  cimd.ProvideNoopServiceRateLimiter(),
						UsageLimiter: cimd.ProvideNoopServiceUsageLimiter(),
						SingleFlight: newWorkingSingleFlight(t),
					}
					err := svc.EnsureClientResolved(ctx, testClientID)
					So(err, ShouldBeNil)
					So(commands.upsertCalls.Load(), ShouldEqual, int64(0))
				})
			}
		})

		Convey("no record, fetch fails: every distinct failure mode yields the SAME error", func() {
			cases := map[string]http.HandlerFunc{
				"404":          func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNotFound) },
				"invalid json": func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("not json")) },
				"client_id mismatch": func(w http.ResponseWriter, r *http.Request) {
					_, _ = w.Write([]byte(`{"client_id":"https://wrong.example.com/x","redirect_uris":["https://x/cb"]}`))
				},
				"oversize": func(w http.ResponseWriter, r *http.Request) {
					_, _ = w.Write(make([]byte, cimd.MaxDocumentBytes+1))
				},
			}
			var firstErr error
			for _, handler := range cases {
				ds := newDocumentServer(handler)
				svc := &cimd.Service{
					OAuthConfig:  enabledOAuthConfig(),
					Clock:        clock.NewMockClockAt("2026-08-31T12:00:00Z"),
					Fetcher:      fetcherFor(ds),
					Commands:     &stubCommands{},
					Queries:      &stubQueries{},
					Database:     &stubDatabase{},
					RateLimiter:  cimd.ProvideNoopServiceRateLimiter(),
					UsageLimiter: cimd.ProvideNoopServiceUsageLimiter(),
					SingleFlight: newWorkingSingleFlight(t),
				}
				err := svc.EnsureClientResolved(ctx, testClientID)
				ds.Close()
				So(apierrors.IsKind(err, cimd.CIMDUnresolvable), ShouldBeTrue)
				if firstErr == nil {
					firstErr = err
				} else {
					// Every mode must be byte-identical -- the SSRF-oracle
					// invariant.
					So(err.Error(), ShouldEqual, firstErr.Error())
				}
			}
		})

		Convey("single-flight not acquired, record exists: nil, fetcher never called", func() {
			sf, _ := newTestSingleFlight(t)
			acquired, err := sf.Acquire(ctx, testClientID)
			So(err, ShouldBeNil)
			So(acquired, ShouldBeTrue) // consume the lock so Service's own Acquire fails

			ds := newDocumentServer(func(w http.ResponseWriter, r *http.Request) {})
			defer ds.Close()
			fetchedAt := time.Now().Add(-2 * time.Hour)
			queries := &stubQueries{getFn: func() (*oauthclient.Client, error) {
				return &oauthclient.Client{ClientID: testClientID, Source: model.OAuthClientSourceCIMD, LastFetchedAt: &fetchedAt}, nil
			}}
			svc := &cimd.Service{
				OAuthConfig:  enabledOAuthConfig(),
				Clock:        clock.NewMockClockAt("2026-08-31T12:00:00Z"),
				Fetcher:      fetcherFor(ds),
				Queries:      queries,
				Database:     &stubDatabase{},
				RateLimiter:  cimd.ProvideNoopServiceRateLimiter(),
				UsageLimiter: cimd.ProvideNoopServiceUsageLimiter(),
				SingleFlight: sf,
			}
			err = svc.EnsureClientResolved(ctx, testClientID)
			So(err, ShouldBeNil)
			So(ds.hits.Load(), ShouldEqual, int64(0))
		})

		Convey("single-flight not acquired, no record: CIMDUnresolvable, fetcher never called", func() {
			sf, _ := newTestSingleFlight(t)
			_, err := sf.Acquire(ctx, testClientID)
			So(err, ShouldBeNil)

			ds := newDocumentServer(func(w http.ResponseWriter, r *http.Request) {})
			defer ds.Close()
			svc := &cimd.Service{
				OAuthConfig:  enabledOAuthConfig(),
				Fetcher:      fetcherFor(ds),
				Queries:      &stubQueries{},
				Database:     &stubDatabase{},
				RateLimiter:  cimd.ProvideNoopServiceRateLimiter(),
				UsageLimiter: cimd.ProvideNoopServiceUsageLimiter(),
				SingleFlight: sf,
			}
			err = svc.EnsureClientResolved(ctx, testClientID)
			So(apierrors.IsKind(err, cimd.CIMDUnresolvable), ShouldBeTrue)
			So(ds.hits.Load(), ShouldEqual, int64(0))
		})

		Convey("single-flight Acquire errors (Redis down): fetch proceeds", func() {
			sf, mr := newTestSingleFlight(t)
			mr.Close()

			ds := newDocumentServer(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			})
			defer ds.Close()
			svc := &cimd.Service{
				OAuthConfig:  enabledOAuthConfig(),
				Fetcher:      fetcherFor(ds),
				Queries:      &stubQueries{},
				Database:     &stubDatabase{},
				RateLimiter:  cimd.ProvideNoopServiceRateLimiter(),
				UsageLimiter: cimd.ProvideNoopServiceUsageLimiter(),
				SingleFlight: sf,
			}
			err := svc.EnsureClientResolved(ctx, testClientID)
			So(apierrors.IsKind(err, cimd.CIMDUnresolvable), ShouldBeTrue)
			So(ds.hits.Load(), ShouldEqual, int64(1))
		})

		Convey("Queries returns an infrastructure error: returned unchanged, not CIMDUnresolvable", func() {
			infraErr := errors.New("connection refused")
			ds := newDocumentServer(func(w http.ResponseWriter, r *http.Request) {})
			defer ds.Close()
			queries := &stubQueries{getFn: func() (*oauthclient.Client, error) { return nil, infraErr }}
			svc := &cimd.Service{
				OAuthConfig:  enabledOAuthConfig(),
				Fetcher:      fetcherFor(ds),
				Queries:      queries,
				Database:     &stubDatabase{},
				RateLimiter:  cimd.ProvideNoopServiceRateLimiter(),
				UsageLimiter: cimd.ProvideNoopServiceUsageLimiter(),
				SingleFlight: newWorkingSingleFlight(t),
			}
			err := svc.EnsureClientResolved(ctx, testClientID)
			So(err, ShouldEqual, infraErr)
			So(apierrors.IsKind(err, cimd.CIMDUnresolvable), ShouldBeFalse)
			So(ds.hits.Load(), ShouldEqual, int64(0))
		})

		Convey("called with IsInTx(ctx) == true: panics", func() {
			svc := &cimd.Service{
				OAuthConfig: enabledOAuthConfig(),
				Database:    &stubDatabase{inTx: true},
			}
			So(func() {
				_ = svc.EnsureClientResolved(ctx, testClientID)
			}, ShouldPanic)
		})

		Convey("existing row with Source: DCR under a URL client_id: nil, fetcher never called", func() {
			ds := newDocumentServer(func(w http.ResponseWriter, r *http.Request) {})
			defer ds.Close()
			queries := &stubQueries{getFn: func() (*oauthclient.Client, error) {
				return &oauthclient.Client{ClientID: testClientID, Source: model.OAuthClientSourceDCR}, nil
			}}
			svc := &cimd.Service{
				OAuthConfig:  enabledOAuthConfig(),
				Fetcher:      fetcherFor(ds),
				Queries:      queries,
				Database:     &stubDatabase{},
				RateLimiter:  cimd.ProvideNoopServiceRateLimiter(),
				UsageLimiter: cimd.ProvideNoopServiceUsageLimiter(),
				SingleFlight: newWorkingSingleFlight(t),
			}
			err := svc.EnsureClientResolved(ctx, testClientID)
			So(err, ShouldBeNil)
			So(ds.hits.Load(), ShouldEqual, int64(0))
		})
	})
}
