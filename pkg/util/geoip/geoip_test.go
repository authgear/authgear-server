package geoip

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestIPString(t *testing.T) {
	Convey("IPString", t, func() {
		ipStr := "42.200.192.29"

		db, err := Open("../../../GeoLite2-Country.mmdb")
		So(err, ShouldBeNil)

		info, ok := db.IPString(ipStr)
		So(ok, ShouldBeTrue)
		So(info.CountryCode, ShouldEqual, "HK")
		So(info.EnglishCountryName, ShouldEqual, "Hong Kong")
	})
}
