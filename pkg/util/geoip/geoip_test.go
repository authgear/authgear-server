package geoip

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestIPString(t *testing.T) {
	Convey("IPString", t, func() {
		ipStr := "42.200.192.29"

		// This serves as the documentation of the version of the file.
		metadata := reader.Metadata()
		//nolint:gosec // G115
		sec := int64(metadata.BuildEpoch)
		build := time.Unix(sec, 0).UTC().Format(time.RFC3339)
		So(build, ShouldEqual, "2025-12-12T09:02:25Z")

		info, ok := IPString(ipStr)
		So(ok, ShouldBeTrue)
		So(info.CountryCode, ShouldEqual, "HK")
		So(info.EnglishCountryName, ShouldEqual, "Hong Kong")
	})
}
