package handler

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetGearName(t *testing.T) {
	Convey("GetGearName", t, func() {
		Convey("should return gear name from path", func() {
			cases := []struct {
				path string
				gear string
			}{
				{"/_auth", "auth"},
				{"/_auth/login/", "auth"},
				{"/auth/", ""},
				{"/_auth////", "auth"},
				{"/_asset/", "asset"},
				{"/index", ""},
				{"", ""},
			}

			for _, c := range cases {
				So(getGearName(c.path), ShouldEqual, c.gear)
			}
		})
	})
}
