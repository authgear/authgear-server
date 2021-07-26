package secrets

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestShannonEntropy(t *testing.T) {
	Convey("ShannonEntropy", t, func() {
		So(ShannonEntropy(""), ShouldEqual, 0)
		So(ShannonEntropy("a"), ShouldEqual, 0)
		So(ShannonEntropy("aa"), ShouldEqual, 0)

		So(ShannonEntropy("The quick brown fox jumps over the lazy dog"), ShouldEqual, 215)

		So(ShannonEntropy("secret"), ShouldEqual, 18)

		So(ShannonEntropy("651fa8cf695de007335f2699fd4c7894c199bea4285d36196f35ea202ed4264f"), ShouldEqual, 256)
	})
}
