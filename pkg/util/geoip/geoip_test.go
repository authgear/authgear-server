package geoip

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestIPString(t *testing.T) {
	Convey("IPString", t, func() {
		ipStr := "42.200.192.29"

		db, err := Open("../../../GeoLite2-Country.mmdb")
		So(err, ShouldBeNil)

		// This serves as the documentation of the version of the file.
		metadata := db.reader.Metadata()
		build := time.Unix(int64(metadata.BuildEpoch), 0).UTC().Format(time.RFC3339)
		So(build, ShouldEqual, "2022-09-01T15:36:01Z")

		info, ok := db.IPString(ipStr)
		So(ok, ShouldBeTrue)
		So(info.CountryCode, ShouldEqual, "HK")
		So(info.EnglishCountryName, ShouldEqual, "Hong Kong")
	})
}
