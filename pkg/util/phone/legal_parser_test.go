package phone

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestLegalParser(t *testing.T) {
	parser := &legalParser{}
	Convey("Phone", t, func() {
		Convey("checkE164", func() {
			good := "+85223456789"
			So(parser.CheckE164(good), ShouldBeNil)

			bad := " +85223456789 "
			So(parser.CheckE164(bad), ShouldBeError, "not in E.164 format")

			withLetter := "+85222a"
			So(parser.CheckE164(withLetter), ShouldBeError, "not in E.164 format")

			invalid := "+85212345678"
			So(parser.CheckE164(invalid), ShouldBeNil)

			tooShort := "+85222"
			So(parser.CheckE164(tooShort), ShouldBeNil)

			plus := "+"
			So(parser.CheckE164(plus), ShouldBeError, "not in E.164 format")

			plusCountryCode := "+852"
			So(parser.CheckE164(plusCountryCode), ShouldBeError, "not in E.164 format")

			nonsense := "a"
			So(parser.CheckE164(nonsense), ShouldBeError, "not in E.164 format")

			empty := ""
			So(parser.CheckE164(empty), ShouldBeError, "not in E.164 format")
		})
	})
}
