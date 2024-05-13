package hexstring_test

import (
	"math/big"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/hexstring"
)

func TestHexCmp(t *testing.T) {
	Convey("NewFromInt64 normal", t, func() {
		h1, err := hexstring.NewFromInt64(16)
		So(err, ShouldBeNil)
		So(h1, ShouldEqual, hexstring.T("0x10"))

		h2, err := hexstring.NewFromInt64(24)
		So(err, ShouldBeNil)
		So(h2, ShouldEqual, hexstring.T("0x18"))

		h3, err := hexstring.NewFromInt64(0)
		So(err, ShouldBeNil)
		So(h3, ShouldEqual, hexstring.T("0x0"))
	})

	Convey("NewFromInt64 failure", t, func() {
		h1, err := hexstring.NewFromInt64(-20)
		So(err, ShouldBeError, "value must be positive")
		So(h1, ShouldEqual, hexstring.T(""))
	})

	Convey("NewFromBigInt normal", t, func() {
		h1, err := hexstring.NewFromBigInt(big.NewInt(5029))
		So(err, ShouldBeNil)
		So(h1, ShouldEqual, hexstring.T("0x13a5"))

		h2, err := hexstring.NewFromBigInt(big.NewInt(2738))
		So(err, ShouldBeNil)
		So(h2, ShouldEqual, hexstring.T("0xab2"))

		h3, err := hexstring.NewFromBigInt(big.NewInt(0))
		So(err, ShouldBeNil)
		So(h3, ShouldEqual, hexstring.T("0x0"))
	})

	Convey("NewFromBigInt failure", t, func() {
		h1, err := hexstring.NewFromBigInt(big.NewInt(-1))
		So(err, ShouldBeError, "value must be positive")
		So(h1, ShouldEqual, hexstring.T(""))
	})

	Convey("Parse normal", t, func() {
		h1, err := hexstring.Parse("0x10")
		So(err, ShouldBeNil)
		So(h1, ShouldEqual, hexstring.T("0x10"))

		h2, err := hexstring.Parse("0x18")
		So(err, ShouldBeNil)
		So(h2, ShouldEqual, hexstring.T("0x18"))

		h3, err := hexstring.Parse("0x0")
		So(err, ShouldBeNil)
		So(h3, ShouldEqual, hexstring.T("0x0"))

		h4, err := hexstring.Parse("0x000018")
		So(err, ShouldBeNil)
		So(h4, ShouldEqual, hexstring.T("0x000018"))
	})

	Convey("Parse failure", t, func() {
		h, err := hexstring.Parse("10")
		So(err, ShouldBeError, `hex string must match the regexp "^(0x)0*([0-9a-fA-F]+)$"`)
		So(h, ShouldEqual, hexstring.T(""))

		h, err = hexstring.Parse("0xg")
		So(err, ShouldBeError, `hex string must match the regexp "^(0x)0*([0-9a-fA-F]+)$"`)
		So(h, ShouldEqual, hexstring.T(""))
	})

	Convey("Parse normalizes to lowercase", t, func() {
		h, err := hexstring.Parse("0xDEADBEEF")
		So(err, ShouldBeNil)
		So(h, ShouldEqual, hexstring.T("0xdeadbeef"))
	})

	Convey("FindSmallest normal", t, func() {
		hexStrings := []hexstring.T{
			"0x10",
			"0x18",
			"0x1",
		}
		hex, i, ok := hexstring.FindSmallest(hexStrings)

		So(ok, ShouldBeTrue)
		So(hex, ShouldEqual, hexstring.T("0x1"))
		So(i, ShouldEqual, 2)
	})
	Convey("FindSmallest failure", t, func() {
		hexStrings := []hexstring.T{}
		hex, i, ok := hexstring.FindSmallest(hexStrings)

		So(ok, ShouldBeFalse)
		So(hex, ShouldEqual, hexstring.T(""))
		So(i, ShouldEqual, -1)
	})
	Convey("Parse remove leading zeros", t, func() {
		h, err := hexstring.TrimmedParse("0x000000000000000000000DEADBEEF")
		So(err, ShouldBeNil)
		So(h, ShouldEqual, hexstring.T("0xdeadbeef"))
	})
}
