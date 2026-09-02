package cimd

import (
	"context"
	"errors"
	"net"
	"net/netip"
	"strconv"
	"syscall"
	"time"

	"github.com/authgear/authgear-server/pkg/util/httputil"
)

const (
	// MaxDocumentBytes is draft-ietf-oauth-client-id-metadata-document-02
	// §8.7's recommended maximum. Enforced progressively while reading the
	// body, never from Content-Length, which a server may omit or misstate.
	MaxDocumentBytes = 5120

	// FetchTimeout covers DNS resolution, TLS handshake and reading the
	// response body, because it is applied as http.Client.Timeout. It
	// matches the existing blocking-webhook per-call default
	// (hook.sync_hook_timeout_seconds) and every other
	// httputil.NewExternalClient call site in this repo, all of which use 5s.
	FetchTimeout = 5 * time.Second

	// RefetchInterval bounds how often a persisted CIMD row is refreshed.
	// Consumed in Part 3; declared here so every fixed CIMD limit is in one
	// place.
	RefetchInterval = 1 * time.Hour
)

// None of the limits above is project-configurable. docs/specs/cimd.md §
// Configuration says so outright: they are not tuning knobs, they are the
// bounds on what an unauthenticated caller can make this server do to a
// third party. A project owner raising MaxDocumentBytes or the timeout
// raises the cost that OTHER people's infrastructure pays.

var errBlockedAddress = errors.New("cimd: address is not publicly routable")

// netipResolver is the one *net.Resolver method SafeDialer needs, pulled
// out as an interface so tests can stub DNS resolution without spinning up
// a real DNS server. *net.Resolver satisfies it as-is.
type netipResolver interface {
	LookupNetIP(ctx context.Context, network, host string) ([]netip.Addr, error)
}

// SafeDialer implements docs/specs/cimd.md § SSRF Protection's first two
// rules together, because neither works without the other:
//
//   - Resolve the hostname once per fetch attempt and connect only to an
//     address validated in that same resolution -- a second, independent
//     resolution at connect time is vulnerable to DNS rebinding.
//   - Check every address a hostname resolves to, not just the first --
//     reject the whole hostname if any A/AAAA record is non-publicly-
//     routable, rather than filtering to the routable subset. A mixed
//     answer is an attack signature, not a misconfiguration to accommodate.
//
// AllowNonPublicAddresses carries insecure_fetch_address_allowed as a plain
// field rather than a config read, so the dialer stays a pure function of
// its inputs. There is no DevMode field: exactly one mechanism widens this
// policy.
type SafeDialer struct {
	Resolver                netipResolver // nil means net.DefaultResolver
	AllowNonPublicAddresses bool
}

func (d *SafeDialer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	portNum, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		return nil, err
	}

	// An IP-literal host never goes through DNS. oauthclient.ParseCIMDClientID
	// lets an IP-literal client_id through shape validation; this is where
	// it is refused.
	if literal, err := netip.ParseAddr(host); err == nil {
		if !d.allow(literal) {
			return nil, errBlockedAddress
		}
		return d.dial(ctx, network, netip.AddrPortFrom(literal, uint16(portNum)))
	}

	resolver := d.Resolver
	if resolver == nil {
		resolver = net.DefaultResolver
	}
	addrs, err := resolver.LookupNetIP(ctx, "ip", host)
	if err != nil {
		return nil, err
	}
	if len(addrs) == 0 {
		return nil, errBlockedAddress
	}
	for _, a := range addrs {
		if !d.allow(a) {
			return nil, errBlockedAddress
		}
	}

	// Connect to an address from THAT resolution. Each is already validated;
	// trying them in order keeps a dual-stack host with an unreachable AAAA
	// working.
	var lastErr error
	for _, a := range addrs {
		conn, err := d.dial(ctx, network, netip.AddrPortFrom(a, uint16(portNum)))
		if err == nil {
			return conn, nil
		}
		lastErr = err
	}
	return nil, lastErr
}

func (d *SafeDialer) allow(addr netip.Addr) bool {
	if httputil.IsPubliclyRoutable(addr) {
		return true
	}
	return d.AllowNonPublicAddresses
}

// dial connects to the already-validated address ap. Control re-validates
// the concrete syscall address as defence in depth: by construction it
// should never fire, and it exists so a future refactor that reintroduces
// hostname-based dialling cannot silently reopen the DNS-rebinding hole.
//
// Only network values "tcp", "tcp4" and "tcp6" occur here.
func (d *SafeDialer) dial(ctx context.Context, network string, ap netip.AddrPort) (net.Conn, error) {
	dialer := &net.Dialer{
		Timeout: FetchTimeout,
		Control: func(network, address string, c syscall.RawConn) error {
			parsed, err := netip.ParseAddrPort(address)
			if err != nil {
				return err
			}
			if !d.allow(parsed.Addr()) {
				return errBlockedAddress
			}
			return nil
		},
	}
	return dialer.DialContext(ctx, network, ap.String())
}
