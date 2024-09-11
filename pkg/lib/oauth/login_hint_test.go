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
		good("https://authgear.com/login_hint?type=login_id&email=test%40example.com", LoginHint{
			Type:         LoginHintTypeLoginID,
			LoginIDEmail: "test@example.com",
		})
		good("https://authgear.com/login_hint?type=login_id&phone=%2B12345", LoginHint{
			Type:         LoginHintTypeLoginID,
			LoginIDPhone: "+12345",
		})
		good("https://authgear.com/login_hint?type=login_id&username=test", LoginHint{
			Type:            LoginHintTypeLoginID,
			LoginIDUsername: "test",
		})
	})

	Convey("LoginHint.String()", t, func() {
		Convey("login_id with email", func() {
			hint := LoginHint{
				Type:         LoginHintTypeLoginID,
				LoginIDEmail: "test@example.com",
				Enforce:      true,
			}
			So(hint.String(), ShouldEqual, "https://authgear.com/login_hint?email=test%40example.com&enforce=true&type=login_id")
		})
		Convey("login_id with phone", func() {
			hint := LoginHint{
				Type:         LoginHintTypeLoginID,
				LoginIDPhone: "+12345",
			}
			So(hint.String(), ShouldEqual, "https://authgear.com/login_hint?phone=%2B12345&type=login_id")
		})
		Convey("login_id with username", func() {
			hint := LoginHint{
				Type:            LoginHintTypeLoginID,
				LoginIDUsername: "test",
			}
			So(hint.String(), ShouldEqual, "https://authgear.com/login_hint?type=login_id&username=test")
		})
		Convey("app_session_token", func() {
			hint := LoginHint{
				Type:            LoginHintTypeAppSessionToken,
				AppSessionToken: "test",
			}
			So(hint.String(), ShouldEqual, "https://authgear.com/login_hint?app_session_token=test&type=app_session_token")
		})
		Convey("anonymous", func() {
			hint := LoginHint{
				Type:          LoginHintTypeAnonymous,
				JWT:           "testjwt",
				PromotionCode: "testcode",
			}
			So(hint.String(), ShouldEqual, "https://authgear.com/login_hint?jwt=testjwt&promotion_code=testcode&type=anonymous")
		})
	})
}
