package httputil_test

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strconv"
	"strings"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

type stubResolver struct {
	addrs []netip.Addr
	err   error
}

func (r *stubResolver) LookupNetIP(ctx context.Context, network, host string) ([]netip.Addr, error) {
	return r.addrs, r.err
}

func TestNewSSRFSafeExternalClient(t *testing.T) {
	Convey("NewSSRFSafeExternalClient", t, func() {
		ctx := context.Background()

		// A real server on loopback stands in for anything inside the
		// deployment's own network.
		reached := false
		internal := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reached = true
			w.WriteHeader(200)
		}))
		defer internal.Close()

		Convey("refuses a loopback destination, and does not reach it", func() {
			client := httputil.NewSSRFSafeExternalClient(5*time.Second, httputil.SSRFSafeExternalClientOptions{})

			req, err := http.NewRequestWithContext(ctx, "GET", internal.URL, nil)
			So(err, ShouldBeNil)

			_, err = client.Do(req)
			So(err, ShouldNotBeNil)
			So(errors.Is(err, httputil.ErrBlockedAddress), ShouldBeTrue)
			So(reached, ShouldBeFalse)
		})

		Convey("reaches it when the deployment has allowed non-public addresses", func() {
			client := httputil.NewSSRFSafeExternalClient(5*time.Second, httputil.SSRFSafeExternalClientOptions{
				AllowNonPublicAddresses: true,
			})

			req, err := http.NewRequestWithContext(ctx, "GET", internal.URL, nil)
			So(err, ShouldBeNil)

			resp, err := client.Do(req)
			So(err, ShouldBeNil)
			defer resp.Body.Close()
			So(resp.StatusCode, ShouldEqual, 200)
			So(reached, ShouldBeTrue)
		})
	})
}

func TestSafeDialerAllowedHosts(t *testing.T) {
	Convey("SafeDialer.AllowedHosts", t, func() {
		ctx := context.Background()

		ln, err := net.Listen("tcp", "127.0.0.1:0")
		So(err, ShouldBeNil)
		defer ln.Close()
		addrPort, err := netip.ParseAddrPort(ln.Addr().String())
		So(err, ShouldBeNil)
		port := strconv.Itoa(int(addrPort.Port()))

		// Every case below resolves to loopback, so only the allowlist can
		// make the difference.
		resolver := &stubResolver{addrs: []netip.Addr{addrPort.Addr()}}

		Convey("an exact host is exempt from the address rules", func() {
			d := &httputil.SafeDialer{Resolver: resolver, AllowedHosts: []string{"hooks.internal.example.com"}}
			conn, err := d.DialContext(ctx, "tcp", net.JoinHostPort("hooks.internal.example.com", port))
			So(err, ShouldBeNil)
			defer conn.Close()
		})

		Convey("a *. pattern matches exactly one label", func() {
			d := &httputil.SafeDialer{Resolver: resolver, AllowedHosts: []string{"*.internal.example.com"}}

			conn, err := d.DialContext(ctx, "tcp", net.JoinHostPort("hooks.internal.example.com", port))
			So(err, ShouldBeNil)
			defer conn.Close()

			// Two labels deep does not match, and so is refused.
			_, err = d.DialContext(ctx, "tcp", net.JoinHostPort("a.b.internal.example.com", port))
			So(errors.Is(err, httputil.ErrBlockedAddress), ShouldBeTrue)

			// Neither does the apex.
			_, err = d.DialContext(ctx, "tcp", net.JoinHostPort("internal.example.com", port))
			So(errors.Is(err, httputil.ErrBlockedAddress), ShouldBeTrue)
		})

		Convey("matching is case-insensitive", func() {
			d := &httputil.SafeDialer{Resolver: resolver, AllowedHosts: []string{"Hooks.Internal.Example.COM"}}
			conn, err := d.DialContext(ctx, "tcp", net.JoinHostPort("hooks.internal.example.com", port))
			So(err, ShouldBeNil)
			defer conn.Close()
		})

		Convey("a host not on the list is still refused", func() {
			d := &httputil.SafeDialer{Resolver: resolver, AllowedHosts: []string{"hooks.internal.example.com"}}
			_, err := d.DialContext(ctx, "tcp", net.JoinHostPort("elsewhere.example.com", port))
			So(errors.Is(err, httputil.ErrBlockedAddress), ShouldBeTrue)
		})

		Convey("an empty list exempts nothing", func() {
			d := &httputil.SafeDialer{Resolver: resolver, AllowedHosts: nil}
			_, err := d.DialContext(ctx, "tcp", net.JoinHostPort("hooks.internal.example.com", port))
			So(errors.Is(err, httputil.ErrBlockedAddress), ShouldBeTrue)
		})
	})
}

func TestSafeDialerRebinding(t *testing.T) {
	Convey("SafeDialer", t, func() {
		ctx := context.Background()

		Convey("rejects the hostname if ANY resolved address is not routable", func() {
			// The rebinding shape: one public answer alongside a private one.
			// Filtering to the routable subset would connect; rejecting the
			// whole answer is what makes a mixed reply safe.
			d := &httputil.SafeDialer{Resolver: &stubResolver{addrs: []netip.Addr{
				netip.MustParseAddr("93.184.216.34"),
				netip.MustParseAddr("10.0.0.1"),
			}}}
			_, err := d.DialContext(ctx, "tcp", "rebind.example.com:80")
			So(errors.Is(err, httputil.ErrBlockedAddress), ShouldBeTrue)
		})

		Convey("rejects an IPv4-mapped IPv6 loopback", func() {
			d := &httputil.SafeDialer{Resolver: &stubResolver{addrs: []netip.Addr{
				netip.MustParseAddr("::ffff:127.0.0.1"),
			}}}
			_, err := d.DialContext(ctx, "tcp", "mapped.example.com:80")
			So(errors.Is(err, httputil.ErrBlockedAddress), ShouldBeTrue)
		})

		Convey("rejects the cloud metadata address given as an IP literal", func() {
			d := &httputil.SafeDialer{}
			_, err := d.DialContext(ctx, "tcp", "169.254.169.254:80")
			So(errors.Is(err, httputil.ErrBlockedAddress), ShouldBeTrue)
		})

		Convey("rejects an empty DNS answer rather than falling through", func() {
			d := &httputil.SafeDialer{Resolver: &stubResolver{addrs: nil}}
			_, err := d.DialContext(ctx, "tcp", "empty.example.com:80")
			So(errors.Is(err, httputil.ErrBlockedAddress), ShouldBeTrue)
		})

		Convey("surfaces a resolver failure as itself, not as a block", func() {
			resolverErr := errors.New("dns is down")
			d := &httputil.SafeDialer{Resolver: &stubResolver{err: resolverErr}}
			_, err := d.DialContext(ctx, "tcp", "broken.example.com:80")
			So(errors.Is(err, resolverErr), ShouldBeTrue)
			So(errors.Is(err, httputil.ErrBlockedAddress), ShouldBeFalse)
		})

		Convey("connects when every resolved address is routable", func() {
			// Points a public-looking name at a listener on loopback, with the
			// restriction lifted, to prove the dial path itself works.
			ln, err := net.Listen("tcp", "127.0.0.1:0")
			So(err, ShouldBeNil)
			defer ln.Close()

			addrPort, err := netip.ParseAddrPort(ln.Addr().String())
			So(err, ShouldBeNil)

			d := &httputil.SafeDialer{
				Resolver:                &stubResolver{addrs: []netip.Addr{addrPort.Addr()}},
				AllowNonPublicAddresses: true,
			}
			conn, err := d.DialContext(ctx, "tcp", net.JoinHostPort("host.example.com", strconv.Itoa(int(addrPort.Port()))))
			So(err, ShouldBeNil)
			defer conn.Close()
		})
	})
}

func TestSafeDialerBlockedAddressLog(t *testing.T) {
	Convey("SafeDialer logs what it refuses", t, func() {
		var w strings.Builder
		ctx := slogutil.SetContextLogger(context.Background(), slog.New(slogutil.NewHandlerForTesting(slog.LevelInfo, &w)))

		Convey("one ERROR naming the sink, the host and the flag to lift", func() {
			d := &httputil.SafeDialer{
				Resolver: &stubResolver{addrs: []netip.Addr{netip.MustParseAddr("169.254.169.254")}},
				Sink:     "hook.blocking_handlers",
			}

			_, err := d.DialContext(ctx, "tcp", "metadata.example.com:80")
			So(errors.Is(err, httputil.ErrBlockedAddress), ShouldBeTrue)

			So(w.String(), ShouldEqual, `level=ERROR msg="outbound fetch blocked by SSRF address policy" logger=ssrf-address-policy error="httputil: address is not publicly routable" blocked_by_ssrf_policy=true sink=hook.blocking_handlers host=metadata.example.com flag=http.insecure_fetch_address_allowed
`)
		})

		Convey("nothing when the destination is allowed", func() {
			d := &httputil.SafeDialer{
				Resolver:                &stubResolver{addrs: []netip.Addr{netip.MustParseAddr("127.0.0.1")}},
				AllowNonPublicAddresses: true,
				Sink:                    "hook.blocking_handlers",
			}

			// The dial itself fails -- nothing is listening -- which is the
			// point: a connection failure is not a policy refusal.
			_, err := d.DialContext(ctx, "tcp", "localhost:1")
			So(err, ShouldNotBeNil)
			So(errors.Is(err, httputil.ErrBlockedAddress), ShouldBeFalse)

			So(w.String(), ShouldEqual, "")
		})

		Convey("nothing when the name does not resolve", func() {
			d := &httputil.SafeDialer{
				Resolver: &stubResolver{err: errors.New("no such host")},
				Sink:     "hook.blocking_handlers",
			}

			_, err := d.DialContext(ctx, "tcp", "nonexistent.example.com:443")
			So(err, ShouldNotBeNil)

			So(w.String(), ShouldEqual, "")
		})
	})
}
