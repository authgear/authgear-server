package cimd

import (
	"context"
	"errors"
	"net"
	"net/netip"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
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
