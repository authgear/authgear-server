package httputil_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func TestIsLoopbackHost(t *testing.T) {
	Convey("IsLoopbackHost", t, func() {
		accepted := []string{"localhost", "127.0.0.1", "::1"}
		for _, host := range accepted {
			host := host
			Convey("accepts "+host, func() {
				So(httputil.IsLoopbackHost(host), ShouldBeTrue)
			})
		}

		rejected := []string{"127.0.0.2", "0.0.0.0", "example.com", "", "localhost.example.com"}
		for _, host := range rejected {
			host := host
			Convey("rejects "+host, func() {
				So(httputil.IsLoopbackHost(host), ShouldBeFalse)
			})
		}
	})
}
