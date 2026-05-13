package declarative

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestStandardAttributeForChannel(t *testing.T) {
	Convey("standardAttributeForChannel", t, func() {
		Convey("nil stdAttrs returns empty string", func() {
			So(standardAttributeForChannel(nil, AccountRecoveryChannelEmail), ShouldEqual, "")
		})
		Convey("email+phone stdAttrs, channel=email returns email", func() {
			stdAttrs := map[string]any{
				"email":        "user@example.com",
				"phone_number": "+85291234567",
			}
			So(standardAttributeForChannel(stdAttrs, AccountRecoveryChannelEmail), ShouldEqual, "user@example.com")
		})
		Convey("email+phone stdAttrs, channel=sms returns phone", func() {
			stdAttrs := map[string]any{
				"email":        "user@example.com",
				"phone_number": "+85291234567",
			}
			So(standardAttributeForChannel(stdAttrs, AccountRecoveryChannelSMS), ShouldEqual, "+85291234567")
		})
		Convey("email+phone stdAttrs, channel=whatsapp returns phone", func() {
			stdAttrs := map[string]any{
				"email":        "user@example.com",
				"phone_number": "+85291234567",
			}
			So(standardAttributeForChannel(stdAttrs, AccountRecoveryChannelWhatsapp), ShouldEqual, "+85291234567")
		})
		Convey("only email, channel=sms returns empty string", func() {
			stdAttrs := map[string]any{
				"email": "user@example.com",
			}
			So(standardAttributeForChannel(stdAttrs, AccountRecoveryChannelSMS), ShouldEqual, "")
		})
	})
}
