package password

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestParsePasswordFormat(t *testing.T) {
	Convey("parsePasswordFormat", t, func() {
		Convey("should parse correct password hash", func() {
			h := []byte("$2a$10$4yzWhYLTp56Aire5CaS2EuUQjs0TiDa83faJe095mUeajNJUyrJDK")
			id, data, err := parsePasswordFormat(h)
			So(err, ShouldBeNil)
			So(string(id), ShouldEqual, "2a")
			So(string(data), ShouldEqual, "10$4yzWhYLTp56Aire5CaS2EuUQjs0TiDa83faJe095mUeajNJUyrJDK")
		})
		Convey("should reject incorrectly formatted password hash", func() {
			h := []byte("$test")
			_, _, err := parsePasswordFormat(h)
			So(err, ShouldBeError)
		})
	})
}
