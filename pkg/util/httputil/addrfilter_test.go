package httputil_test

import (
	"net/netip"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func TestIsPubliclyRoutable(t *testing.T) {
	Convey("IsPubliclyRoutable", t, func() {
		mustRejectV4OrV6 := []string{
			"127.0.0.1",
			"127.1.2.3",
			"::1",
			"10.0.0.1",
			"172.16.0.1",
			"192.168.1.1",
			"fc00::1",
			"fd12::1",
			"169.254.169.254",
			"169.254.0.1",
			"fe80::1",
			"224.0.0.1",
			"ff02::1",
			"0.0.0.0",
			"::",
			"100.64.0.1",
			"192.0.2.1",
			"198.18.0.1",
			"198.51.100.1",
			"203.0.113.1",
			"240.0.0.1",
			"255.255.255.255",
			"2001:db8::1",
			"2002::1",
			"64:ff9b::7f00:1",
			"2001::1",
		}
		for _, s := range mustRejectV4OrV6 {
			Convey("rejects "+s, func() {
				addr := netip.MustParseAddr(s)
				So(httputil.IsPubliclyRoutable(addr), ShouldBeFalse)
			})
		}

		Convey("rejects the zero netip.Addr{}", func() {
			So(httputil.IsPubliclyRoutable(netip.Addr{}), ShouldBeFalse)
		})

		Convey("rejects ::ffff:127.0.0.1 (regresses if Unmap() is removed)", func() {
			addr := netip.MustParseAddr("::ffff:127.0.0.1")
			So(httputil.IsPubliclyRoutable(addr), ShouldBeFalse)
		})

		mustAccept := []string{
			"1.1.1.1",
			"8.8.8.8",
			"93.184.216.34",
			"2606:4700::1111",
		}
		for _, s := range mustAccept {
			Convey("accepts "+s, func() {
				addr := netip.MustParseAddr(s)
				So(httputil.IsPubliclyRoutable(addr), ShouldBeTrue)
			})
		}

		Convey("accepts ::ffff:1.1.1.1 (4-in-6 form of a public address)", func() {
			addr := netip.MustParseAddr("::ffff:1.1.1.1")
			So(httputil.IsPubliclyRoutable(addr), ShouldBeTrue)
		})
	})
}
