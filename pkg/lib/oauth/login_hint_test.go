package oauth

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestParseLoginHint(t *testing.T) {
	Convey("ParseLoginHint", t, func() {
		good := func(s string, expected LoginHint) {
			actual, err := ParseLoginHint(s)
			So(err, ShouldBeNil)
			So(*actual, ShouldResemble, expected)
		}

		bad := func(s string, errString string) {
			_, err := ParseLoginHint(s)
			So(err, ShouldBeError, errString)
		}

		bad("", "invalid login_hint: ")
		bad("https://authgear.com/login_hint?", "invalid login_hint type: ")

		good("https://authgear.com/login_hint?type=anonymous&jwt=jwt", LoginHint{
			Type: LoginHintTypeAnonymous,
			JWT:  "jwt",
		})
		good("https://authgear.com/login_hint?type=anonymous&promotion_code=code", LoginHint{
			Type:          LoginHintTypeAnonymous,
			PromotionCode: "code",
		})
		good("https://authgear.com/login_hint?type=app_session_token&app_session_token=token", LoginHint{
			Type:            LoginHintTypeAppSessionToken,
			AppSessionToken: "token",
		})
	})
}
