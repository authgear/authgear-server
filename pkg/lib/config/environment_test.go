package config_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

func TestAppHostSuffixes(t *testing.T) {
	Convey("AppHostSuffixes", t, func() {
		prod := config.AppHostSuffixes([]string{
			".authgearapps.com",
			".authgear-apps.com",
		})
		local := config.AppHostSuffixes([]string{
			".localhost:3100",
		})

		test := func(suffixes config.AppHostSuffixes, host string, expected bool) {
			actual := suffixes.CheckIsDefaultDomain(host)
			So(actual, ShouldEqual, expected)
		}

		test(prod, "myapp.authgearapps.com", true)
		test(prod, "myapp.authgear-apps.com", true)
		test(prod, "accounts.portal.authgear-apps.com", false)
		test(prod, "accounts.portal.authgear.com", false)
		test(prod, "example.com", false)

		test(local, "myapp.localhost:3100", true)
		test(local, "accounts.portal.localhost:3100", false)
	})
}
