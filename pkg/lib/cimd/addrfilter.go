package cimd

import "net/netip"

// additionalBlockedPrefixes holds only what netip.Addr's own predicates do
// NOT cover. IsLoopback/IsPrivate/IsLinkLocalUnicast/IsLinkLocalMulticast/
// IsInterfaceLocalMulticast/IsMulticast/IsUnspecified already cover
// RFC 1918, fc00::/7 (via IsPrivate), loopback, link-local (incl. the
// 169.254.0.0/16 cloud-metadata range), and multicast. RFC 6890 is the
// umbrella reference for docs/specs/cimd.md § SSRF Protection.
//
// Do not "simplify" this to IsGlobalUnicast(): it returns true for
// 10.0.0.1, 172.16.0.1, fc00::1, 100.64.0.1, 192.0.2.1 and 240.0.0.1. It
// excludes only unspecified/loopback/multicast/link-local, so as an
// "is this public" test it admits every RFC 1918 address.
var additionalBlockedPrefixes = []netip.Prefix{
	netip.MustParsePrefix("100.64.0.0/10"),      // RFC 6598 CGNAT
	netip.MustParsePrefix("192.0.0.0/24"),       // RFC 6890 IETF protocol assignments
	netip.MustParsePrefix("192.0.2.0/24"),       // RFC 5737 TEST-NET-1
	netip.MustParsePrefix("198.18.0.0/15"),      // RFC 2544 benchmarking
	netip.MustParsePrefix("198.51.100.0/24"),    // RFC 5737 TEST-NET-2
	netip.MustParsePrefix("203.0.113.0/24"),     // RFC 5737 TEST-NET-3
	netip.MustParsePrefix("240.0.0.0/4"),        // RFC 1112 reserved
	netip.MustParsePrefix("255.255.255.255/32"), // limited broadcast
	netip.MustParsePrefix("2001:db8::/32"),      // RFC 3849 documentation
	netip.MustParsePrefix("2002::/16"),          // RFC 3056 6to4
	netip.MustParsePrefix("64:ff9b::/96"),       // RFC 6052 NAT64
	netip.MustParsePrefix("64:ff9b:1::/48"),     // RFC 8215 local-use NAT64
	netip.MustParsePrefix("100::/64"),           // RFC 6666 discard-only
	netip.MustParsePrefix("2001::/32"),          // RFC 4380 Teredo
}

// IsPubliclyRoutable reports whether addr is safe to connect to under
// docs/specs/cimd.md § SSRF Protection: not loopback, not private, not
// link-local, not multicast, not unspecified, and not one of the
// additional special-use ranges above -- each of which either embeds an
// arbitrary (possibly private) IPv4 address, or is otherwise not a real
// routable destination.
func IsPubliclyRoutable(addr netip.Addr) bool {
	if !addr.IsValid() {
		return false
	}
	// Unmap first: the table is written in native v4 form, so Contains
	// would miss ::ffff:10.0.0.1 without it.
	a := addr.Unmap()

	switch {
	case a.IsLoopback(),
		a.IsPrivate(),
		a.IsLinkLocalUnicast(),
		a.IsLinkLocalMulticast(),
		a.IsInterfaceLocalMulticast(),
		a.IsMulticast(),
		a.IsUnspecified():
		return false
	}

	for _, p := range additionalBlockedPrefixes {
		if p.Contains(a) {
			return false
		}
	}
	return true
}
