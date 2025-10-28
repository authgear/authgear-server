package config_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

func TestAppHostSuffixesCheckIsDefaultDomain(t *testing.T) {
	Convey("AppHostSuffixes.CheckIsDefaultDomain", t, func() {
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

func TestAppHostSuffixesCheckToWhatsappCloudAPIBizOpaqueCallbackData(t *testing.T) {
	Convey("AppHostSuffixes.ToWhatsappCloudAPIBizOpaqueCallbackData", t, func() {
		Convey("nil suffixes should return empty string", func() {
			var nilinput config.AppHostSuffixes
			So(nilinput.ToWhatsappCloudAPIBizOpaqueCallbackData(), ShouldEqual, "")
		})

		Convey("one suffix should return non empty string", func() {
			one := config.AppHostSuffixes([]string{
				".authgear.cloud",
			})
			actual := one.ToWhatsappCloudAPIBizOpaqueCallbackData()
			So(actual, ShouldEqual, "fc27ef62d6acb6da5ebb30649e0bb0ec4f7c2973f3cb3eac0974424037911209")
			So(len(actual), ShouldBeLessThan, 512)
		})

		Convey("more than suffixes should return non empty string", func() {
			more := config.AppHostSuffixes([]string{
				".authgearapps.com",
				".authgear-apps.com",
				".authgear.cloud",
			})
			actual := more.ToWhatsappCloudAPIBizOpaqueCallbackData()
			So(actual, ShouldEqual, "a9d7ce80bfdc9c48cb29727cad433258e698c71d5205ca1ae5e0cbf325e09371")
			So(len(actual), ShouldBeLessThan, 512)
		})
	})
}
