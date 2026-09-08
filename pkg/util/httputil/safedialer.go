package httputil

import (
	"context"
	"errors"
	"net"
	"net/netip"
	"strconv"
	"syscall"
	"time"
)

// ErrBlockedAddress is returned when SafeDialer refuses to connect because
// the destination is not publicly routable.
//
// Call sites that fetch a URL supplied by a project's configuration match on
// this so the refusal can be reported as a policy decision rather than as
// just another connection failure -- an operator has to be able to tell "we
// blocked this" apart from "the host was down".
var ErrBlockedAddress = errors.New("httputil: address is not publicly routable")

// NetIPResolver is the one *net.Resolver method SafeDialer needs, pulled out
// as an interface so tests can stub DNS resolution without spinning up a real
// DNS server. *net.Resolver satisfies it as-is.
type NetIPResolver interface {
	LookupNetIP(ctx context.Context, network, host string) ([]netip.Addr, error)
}

// SafeDialer refuses to connect to an address that is not publicly routable,
// which is what stops a URL taken from configuration or from a third party
// being used to reach the deployment's own network.
//
// It implements two rules together, because neither works without the other:
//
//   - Resolve the hostname once per dial and connect only to an address
//     validated in that same resolution -- a second, independent resolution
//     at connect time is vulnerable to DNS rebinding.
//   - Check every address a hostname resolves to, not just the first --
//     reject the whole hostname if any A/AAAA record is non-publicly-routable,
//     rather than filtering to the routable subset. A mixed answer is an
//     attack signature, not a misconfiguration to accommodate.
//
// AllowNonPublicAddresses is a plain field rather than a config read, so the
// dialer stays a pure function of its inputs. Whoever constructs it decides
// the policy; there is no second mechanism that widens it.
type SafeDialer struct {
	Resolver                NetIPResolver // nil means net.DefaultResolver
	AllowNonPublicAddresses bool
	// DialTimeout bounds the connect itself. Zero means no dial-specific
	// deadline, leaving whatever the context and http.Client.Timeout impose.
	DialTimeout time.Duration
	// Sink names the configuration that chose the destination, e.g.
	// "hook.blocking_handlers". It appears in the log a refusal writes; see
	// logIfBlockedAddress.
	Sink string
}

// DialContext dials addr, and writes the refusal log if the policy rejects it.
//
// The log is emitted here, not at the call sites that fetch a URL, because
// this is the only place that knows a refusal happened. A caller sees a
// transport error indistinguishable from an unreachable host, so leaving the
// log to callers means every future one has to remember to write it.
func (d *SafeDialer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	conn, err := d.dialContext(ctx, network, addr)
	if err != nil {
		host, _, splitErr := net.SplitHostPort(addr)
		if splitErr != nil {
			host = addr
		}
		logIfBlockedAddress(ctx, err, d.Sink, host)
	}
	return conn, err
}

func (d *SafeDialer) dialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	portNum, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		return nil, err
	}

	// An IP-literal host never goes through DNS.
	if literal, err := netip.ParseAddr(host); err == nil {
		if !d.allow(literal) {
			return nil, ErrBlockedAddress
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
		return nil, ErrBlockedAddress
	}
	for _, a := range addrs {
		if !d.allow(a) {
			return nil, ErrBlockedAddress
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
	if IsPubliclyRoutable(addr) {
		return true
	}
	return d.AllowNonPublicAddresses
}

// dial connects to the already-validated address ap. Control re-validates the
// concrete syscall address as defence in depth: by construction it should
// never fire, and it exists so a future refactor that reintroduces
// hostname-based dialling cannot silently reopen the DNS-rebinding hole.
//
// Only network values "tcp", "tcp4" and "tcp6" occur here.
func (d *SafeDialer) dial(ctx context.Context, network string, ap netip.AddrPort) (net.Conn, error) {
	dialer := &net.Dialer{
		Timeout: d.DialTimeout,
		Control: func(network, address string, c syscall.RawConn) error {
			parsed, err := netip.ParseAddrPort(address)
			if err != nil {
				return err
			}
			if !d.allow(parsed.Addr()) {
				return ErrBlockedAddress
			}
			return nil
		},
	}
	return dialer.DialContext(ctx, network, ap.String())
}
