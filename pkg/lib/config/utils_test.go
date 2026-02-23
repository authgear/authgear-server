package config

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestIntersectAllowlist(t *testing.T) {
	Convey("IntersectAllowlist", t, func() {
		Convey("should return appAllowlist if featureAllowlist is empty", func() {
			appAllowlist := []string{"HK", "US"}
			featureAllowlist := []string{}
			result := IntersectAllowlist(appAllowlist, featureAllowlist)
			So(result, ShouldResemble, appAllowlist)
		})

		Convey("should return empty if appAllowlist is empty", func() {
			appAllowlist := []string{}
			featureAllowlist := []string{"HK", "US"}
			result := IntersectAllowlist(appAllowlist, featureAllowlist)
			So(result, ShouldBeEmpty)
		})

		Convey("should return intersection of two lists", func() {
			appAllowlist := []string{"HK", "US", "TW"}
			featureAllowlist := []string{"HK", "TW", "GB"}
			result := IntersectAllowlist(appAllowlist, featureAllowlist)
			So(result, ShouldResemble, []string{"HK", "TW"})
		})

		Convey("should return empty if no overlap", func() {
			appAllowlist := []string{"HK", "US"}
			featureAllowlist := []string{"GB", "JP"}
			result := IntersectAllowlist(appAllowlist, featureAllowlist)
			So(result, ShouldBeEmpty)
		})

		Convey("should preserve order of appAllowlist", func() {
			appAllowlist := []string{"US", "HK", "TW"}
			featureAllowlist := []string{"HK", "US"}
			result := IntersectAllowlist(appAllowlist, featureAllowlist)
			So(result, ShouldResemble, []string{"US", "HK"})
		})
	})
}
