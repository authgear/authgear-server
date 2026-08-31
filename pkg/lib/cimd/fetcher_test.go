package cimd

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"net/url"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

// stubResolver lets a test control DNS resolution without touching the
// network or standing up a real DNS server.
type stubResolver struct {
	addrs []netip.Addr
	calls atomic.Int64
}

func (r *stubResolver) LookupNetIP(ctx context.Context, network, host string) ([]netip.Addr, error) {
	r.calls.Add(1)
	return r.addrs, nil
}

// countingTCPServer is a bare TCP listener (no TLS, no HTTP) used to prove
// SafeDialer either does or does not attempt a connection: hits counts
// accepted connections.
type countingTCPServer struct {
	ln   net.Listener
	hits atomic.Int64
}

func newCountingTCPServer(t *testing.T, addr string) *countingTCPServer {
	t.Helper()
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	s := &countingTCPServer{ln: ln}
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			s.hits.Add(1)
			conn.Close()
		}
	}()
	return s
}

func (s *countingTCPServer) Close() { s.ln.Close() }
func (s *countingTCPServer) Port() int {
	return s.ln.Addr().(*net.TCPAddr).Port
}

// firstPrivateIPv4 finds a real, non-loopback RFC 1918 address already
// assigned to this host, so the "AllowNonPublicAddresses covers more than
// loopback" cases can dial something genuinely reachable rather than a
// hardcoded address that may not resolve to this machine in every
// environment.
func firstPrivateIPv4(t *testing.T) netip.Addr {
	t.Helper()
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		t.Fatal(err)
	}
	for _, a := range addrs {
		ipNet, ok := a.(*net.IPNet)
		if !ok {
			continue
		}
		ip := ipNet.IP.To4()
		if ip == nil {
			continue
		}
		addr, ok := netip.AddrFromSlice(ip)
		if !ok {
			continue
		}
		if addr.IsLoopback() {
			continue
		}
		if addr.IsPrivate() {
			return addr
		}
	}
	t.Skip("no private IPv4 address assigned to this host")
	return netip.Addr{}
}

func TestSafeDialerDialContext(t *testing.T) {
	Convey("SafeDialer.DialContext", t, func() {
		ctx := context.Background()

		Convey("a mixed resolution (one public, one blocked address) rejects the whole hostname and never dials", func() {
			srv := newCountingTCPServer(t, "127.0.0.1:0")
			defer srv.Close()

			resolver := &stubResolver{addrs: []netip.Addr{
				netip.MustParseAddr("1.1.1.1"),
				netip.MustParseAddr("127.0.0.1"),
			}}
			d := &SafeDialer{Resolver: resolver}

			_, err := d.DialContext(ctx, "tcp", net.JoinHostPort("example.com", "80"))
			So(errors.Is(err, errBlockedAddress), ShouldBeTrue)
			So(srv.hits.Load(), ShouldEqual, int64(0))
		})

		Convey("a resolved loopback address is blocked when AllowNonPublicAddresses is false", func() {
			resolver := &stubResolver{addrs: []netip.Addr{netip.MustParseAddr("127.0.0.1")}}
			d := &SafeDialer{Resolver: resolver, AllowNonPublicAddresses: false}

			_, err := d.DialContext(ctx, "tcp", net.JoinHostPort("example.com", "80"))
			So(errors.Is(err, errBlockedAddress), ShouldBeTrue)
		})

		Convey("a resolved loopback address succeeds when AllowNonPublicAddresses is true", func() {
			srv := newCountingTCPServer(t, "127.0.0.1:0")
			defer srv.Close()

			resolver := &stubResolver{addrs: []netip.Addr{netip.MustParseAddr("127.0.0.1")}}
			d := &SafeDialer{Resolver: resolver, AllowNonPublicAddresses: true}

			conn, err := d.DialContext(ctx, "tcp", net.JoinHostPort("example.com", strconv.Itoa(srv.Port())))
			So(err, ShouldBeNil)
			conn.Close()
		})

		Convey("AllowNonPublicAddresses covers every non-public range, not only loopback", func() {
			privateIP := firstPrivateIPv4(t)
			srv := newCountingTCPServer(t, net.JoinHostPort(privateIP.String(), "0"))
			defer srv.Close()

			resolver := &stubResolver{addrs: []netip.Addr{privateIP}}
			d := &SafeDialer{Resolver: resolver, AllowNonPublicAddresses: true}

			conn, err := d.DialContext(ctx, "tcp", net.JoinHostPort("example.com", strconv.Itoa(srv.Port())))
			So(err, ShouldBeNil)
			conn.Close()
		})

		Convey("169.254.169.254 (cloud metadata) is blocked when AllowNonPublicAddresses is false", func() {
			resolver := &stubResolver{addrs: []netip.Addr{netip.MustParseAddr("169.254.169.254")}}
			d := &SafeDialer{Resolver: resolver, AllowNonPublicAddresses: false}

			_, err := d.DialContext(ctx, "tcp", net.JoinHostPort("example.com", "80"))
			So(errors.Is(err, errBlockedAddress), ShouldBeTrue)
		})

		Convey("169.254.169.254 (cloud metadata) is NOT blocked by the address filter when AllowNonPublicAddresses is true -- documents exactly how dangerous the flag is", func() {
			resolver := &stubResolver{addrs: []netip.Addr{netip.MustParseAddr("169.254.169.254")}}
			d := &SafeDialer{Resolver: resolver, AllowNonPublicAddresses: true}

			dialCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			defer cancel()
			_, err := d.DialContext(dialCtx, "tcp", net.JoinHostPort("example.com", "80"))
			// The point of this test is that the address filter did not
			// reject it -- whatever happens next (timeout, unreachable) is
			// an ordinary network outcome in this sandbox, not the
			// security control under test.
			So(errors.Is(err, errBlockedAddress), ShouldBeFalse)
		})

		Convey("no DNS lookup for an IP-literal host, and it is still policed", func() {
			resolver := &stubResolver{}
			d := &SafeDialer{Resolver: resolver, AllowNonPublicAddresses: false}

			_, err := d.DialContext(ctx, "tcp", net.JoinHostPort("127.0.0.1", "80"))
			So(errors.Is(err, errBlockedAddress), ShouldBeTrue)
			So(resolver.calls.Load(), ShouldEqual, int64(0))
		})

		Convey("an IP-literal loopback host is blocked by default, without any DNS lookup", func() {
			srv := newCountingTCPServer(t, "127.0.0.1:0")
			defer srv.Close()
			resolver := &stubResolver{}
			d := &SafeDialer{Resolver: resolver}

			_, err := d.DialContext(ctx, "tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(srv.Port())))
			So(errors.Is(err, errBlockedAddress), ShouldBeTrue)
			So(resolver.calls.Load(), ShouldEqual, int64(0))
			So(srv.hits.Load(), ShouldEqual, int64(0))
		})

		Convey("an IP-literal private-address host succeeds when AllowNonPublicAddresses is true -- the e2e shape", func() {
			privateIP := firstPrivateIPv4(t)
			srv := newCountingTCPServer(t, net.JoinHostPort(privateIP.String(), "0"))
			defer srv.Close()
			resolver := &stubResolver{}
			d := &SafeDialer{Resolver: resolver, AllowNonPublicAddresses: true}

			conn, err := d.DialContext(ctx, "tcp", net.JoinHostPort(privateIP.String(), strconv.Itoa(srv.Port())))
			So(err, ShouldBeNil)
			conn.Close()
			So(resolver.calls.Load(), ShouldEqual, int64(0))
		})

		Convey("an empty resolution is blocked", func() {
			resolver := &stubResolver{addrs: nil}
			d := &SafeDialer{Resolver: resolver}

			_, err := d.DialContext(ctx, "tcp", net.JoinHostPort("example.com", "80"))
			So(errors.Is(err, errBlockedAddress), ShouldBeTrue)
		})
	})
}

// newLoopbackHTTPClient builds an http.Client wired through SafeDialer
// (AllowNonPublicAddresses: true, since httptest servers bind to loopback)
// with rootCAs trusted, so Fetch-level tests exercise the real dial path
// while only address-policy tests (above) need to assert on address
// policy. resolver, if non-nil, overrides hostname resolution; nil means
// the request's own host is used (works for httptest's default
// "https://127.0.0.1:PORT" URLs, since 127.0.0.1 is an IP literal and never
// touches the resolver).
func newLoopbackHTTPClient(rootCAs *x509.CertPool, resolver netipResolver) *http.Client {
	dialer := &SafeDialer{Resolver: resolver, AllowNonPublicAddresses: true}
	transport := &http.Transport{
		DialContext:     dialer.DialContext,
		TLSClientConfig: &tls.Config{RootCAs: rootCAs},
	}
	return httputil.NewExternalClientWithOptions(FetchTimeout, httputil.ExternalClientOptions{
		FollowRedirect: false,
		Transport:      transport,
	})
}

func certPool(srv *httptest.Server) *x509.CertPool {
	pool := x509.NewCertPool()
	pool.AddCert(srv.Certificate())
	return pool
}

func addCert(pool *x509.CertPool, srv *httptest.Server) {
	pool.AddCert(srv.Certificate())
}

func TestFetcherFetch(t *testing.T) {
	Convey("Fetcher.Fetch", t, func() {
		ctx := context.Background()
		body := bytes.Repeat([]byte("a"), 100)
		jsonBody := append([]byte(`{"padding":"`), append(body, []byte(`"}`)...)...)

		Convey("200 with a small JSON body returns the bytes", func() {
			srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write(jsonBody)
			}))
			defer srv.Close()

			f := &Fetcher{HTTPClients: &CIMDHTTPClients{Strict: newLoopbackHTTPClient(certPool(srv), nil)}}
			u, _ := url.Parse(srv.URL)
			got, err := f.Fetch(ctx, u)
			So(err, ShouldBeNil)
			So(got, ShouldResemble, jsonBody)
		})

		Convey("exactly MaxDocumentBytes is accepted", func() {
			exact := bytes.Repeat([]byte("a"), MaxDocumentBytes)
			srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write(exact)
			}))
			defer srv.Close()

			f := &Fetcher{HTTPClients: &CIMDHTTPClients{Strict: newLoopbackHTTPClient(certPool(srv), nil)}}
			u, _ := url.Parse(srv.URL)
			got, err := f.Fetch(ctx, u)
			So(err, ShouldBeNil)
			So(len(got), ShouldEqual, MaxDocumentBytes)
		})

		Convey("MaxDocumentBytes+1 is refused, via progressive enforcement rather than Content-Length", func() {
			tooLarge := bytes.Repeat([]byte("a"), MaxDocumentBytes+1)
			srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// No explicit Content-Length header is set before writing,
				// so the server falls back to chunked transfer encoding:
				// this exercises the unknown-length path, proving the +1
				// read limit -- not a declared header -- is what catches an
				// oversize body.
				_, _ = w.Write(tooLarge)
			}))
			defer srv.Close()

			f := &Fetcher{HTTPClients: &CIMDHTTPClients{Strict: newLoopbackHTTPClient(certPool(srv), nil)}}
			u, _ := url.Parse(srv.URL)
			_, err := f.Fetch(ctx, u)
			So(errors.Is(err, ErrResponseTooLarge), ShouldBeTrue)
		})

		Convey("a 301 redirect is refused (0 redirects followed) and the target is never requested", func() {
			targetHits := &atomic.Int64{}
			target := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				targetHits.Add(1)
				_, _ = w.Write(jsonBody)
			}))
			defer target.Close()

			srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, target.URL, http.StatusMovedPermanently)
			}))
			defer srv.Close()

			pool := certPool(srv)
			addCert(pool, target)
			f := &Fetcher{HTTPClients: &CIMDHTTPClients{Strict: newLoopbackHTTPClient(pool, nil)}}
			u, _ := url.Parse(srv.URL)
			_, err := f.Fetch(ctx, u)
			So(errors.Is(err, ErrResponseNotOK), ShouldBeTrue)
			So(targetHits.Load(), ShouldEqual, int64(0))
		})

		for _, status := range []int{http.StatusNotFound, http.StatusInternalServerError} {
			status := status
			Convey("a non-2xx status is refused: "+strconv.Itoa(status), func() {
				srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(status)
				}))
				defer srv.Close()

				f := &Fetcher{HTTPClients: &CIMDHTTPClients{Strict: newLoopbackHTTPClient(certPool(srv), nil)}}
				u, _ := url.Parse(srv.URL)
				_, err := f.Fetch(ctx, u)
				So(errors.Is(err, ErrResponseNotOK), ShouldBeTrue)
			})
		}

		Convey("204 No Content is a 2xx per spec's literal 'MUST be 2xx': Fetch succeeds with an empty body -- ParseAndValidate rejects it as not a JSON object, not Fetch", func() {
			srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			}))
			defer srv.Close()

			f := &Fetcher{HTTPClients: &CIMDHTTPClients{Strict: newLoopbackHTTPClient(certPool(srv), nil)}}
			u, _ := url.Parse(srv.URL)
			got, err := f.Fetch(ctx, u)
			So(err, ShouldBeNil)
			So(got, ShouldBeEmpty)
		})

		Convey("a server that never responds times out well under the request context deadline", func() {
			srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				select {
				case <-time.After(10 * time.Second):
				case <-r.Context().Done():
				}
			}))
			defer srv.Close()

			f := &Fetcher{HTTPClients: &CIMDHTTPClients{Strict: newLoopbackHTTPClient(certPool(srv), nil)}}
			u, _ := url.Parse(srv.URL)

			start := time.Now()
			_, err := f.Fetch(ctx, u)
			elapsed := time.Since(start)
			So(err, ShouldNotBeNil)
			So(elapsed, ShouldBeLessThan, 8*time.Second)
		})

		Convey("hostname verification survives the custom DialContext: a cert for the wrong hostname fails", func() {
			srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write(jsonBody)
			}))
			defer srv.Close()

			serverAddr := srv.Listener.Addr().(*net.TCPAddr)
			resolver := &stubResolver{addrs: []netip.Addr{netip.MustParseAddr(serverAddr.IP.String())}}
			f := &Fetcher{HTTPClients: &CIMDHTTPClients{Strict: newLoopbackHTTPClient(certPool(srv), resolver)}}

			// The cert is valid for "example.com", not this hostname -- the
			// SafeDialer resolves it to the real server's address anyway
			// (that's the point being tested), but certificate verification
			// must still fail on the hostname mismatch.
			u, _ := url.Parse("https://wrong-hostname.invalid:" + strconv.Itoa(serverAddr.Port) + "/")
			_, err := f.Fetch(ctx, u)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "certificate")
		})
	})
}

func TestFetcherClientFor(t *testing.T) {
	Convey("Fetcher.clientFor", t, func() {
		strict := &http.Client{}
		insecure := &http.Client{}
		clients := &CIMDHTTPClients{Strict: strict, Insecure: insecure}
		u, _ := url.Parse("https://mcp-client.example.com/oauth/client-metadata.json")

		Convey("feature config absent -> Strict, nothing logged", func() {
			var buf bytes.Buffer
			ctx := slogutil.SetContextLogger(context.Background(), slog.New(slogutil.NewHandlerForTesting(slog.LevelWarn, &buf)))
			f := &Fetcher{HTTPClients: clients, OAuthFeatureConfig: nil, AppID: "test-app"}

			got := f.clientFor(ctx, u)
			So(got, ShouldEqual, strict)
			So(buf.String(), ShouldBeEmpty)
		})

		Convey("insecure_fetch_address_allowed: false -> Strict, nothing logged", func() {
			var buf bytes.Buffer
			ctx := slogutil.SetContextLogger(context.Background(), slog.New(slogutil.NewHandlerForTesting(slog.LevelWarn, &buf)))
			f := &Fetcher{
				HTTPClients: clients,
				OAuthFeatureConfig: &config.OAuthFeatureConfig{
					ClientIDMetadataDocument: &config.OAuthClientIDMetadataDocumentFeatureConfig{
						InsecureFetchAddressAllowed: boolPtr(false),
					},
				},
				AppID: "test-app",
			}

			got := f.clientFor(ctx, u)
			So(got, ShouldEqual, strict)
			So(buf.String(), ShouldBeEmpty)
		})

		Convey("insecure_fetch_address_allowed: true -> Insecure, and a Warn record is emitted with app_id, host and flag name", func() {
			var buf bytes.Buffer
			ctx := slogutil.SetContextLogger(context.Background(), slog.New(slogutil.NewHandlerForTesting(slog.LevelWarn, &buf)))
			f := &Fetcher{
				HTTPClients: clients,
				OAuthFeatureConfig: &config.OAuthFeatureConfig{
					ClientIDMetadataDocument: &config.OAuthClientIDMetadataDocumentFeatureConfig{
						InsecureFetchAddressAllowed: boolPtr(true),
					},
				},
				AppID: "test-app",
			}

			got := f.clientFor(ctx, u)
			So(got, ShouldEqual, insecure)
			logged := buf.String()
			So(logged, ShouldContainSubstring, "test-app")
			So(logged, ShouldContainSubstring, "mcp-client.example.com")
			So(logged, ShouldContainSubstring, "oauth.client_id_metadata_document.insecure_fetch_address_allowed")
		})
	})
}

func boolPtr(b bool) *bool { return &b }
