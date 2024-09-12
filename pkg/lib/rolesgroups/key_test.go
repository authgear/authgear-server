package rolesgroups

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestValidateKey(t *testing.T) {
	Convey("ValidateKey", t, func() {
		So(ValidateKey(""), ShouldBeError, `invalid value:
<root>: minLength
  map[actual:0 expected:1]
<root>: pattern
  map[actual: expected:^[a-zA-Z_][a-zA-Z0-9:_]*$]`)

		So(ValidateKey("a0123456789012345678901234567890123456789"), ShouldBeError, `invalid value:
<root>: maxLength
  map[actual:41 expected:40]`)

		So(ValidateKey("1"), ShouldBeError, `invalid value:
<root>: pattern
  map[actual:1 expected:^[a-zA-Z_][a-zA-Z0-9:_]*$]`)

		So(ValidateKey("#$%^&*AND*&&^%$#"), ShouldBeError, `invalid value:
<root>: pattern
  map[actual:#$%^&*AND*&&^%$# expected:^[a-zA-Z_][a-zA-Z0-9:_]*$]`)

		So(ValidateKey("user#123ok"), ShouldBeError, `invalid value:
<root>: pattern
  map[actual:user#123ok expected:^[a-zA-Z_][a-zA-Z0-9:_]*$]`)

		So(ValidateKey("GOOD key"), ShouldBeError, `invalid value:
<root>: pattern
  map[actual:GOOD key expected:^[a-zA-Z_][a-zA-Z0-9:_]*$]`)

		So(ValidateKey("0admin"), ShouldBeError, `invalid value:
<root>: pattern
  map[actual:0admin expected:^[a-zA-Z_][a-zA-Z0-9:_]*$]`)

		So(ValidateKey("authgear:user"), ShouldBeError, `invalid value:
<root>: format
  map[error:key cannot start with the preserved prefix: `+"`"+`authgear:`+"`"+` format:x_role_group_key]`)

		So(ValidateKey("manager:SUPER_ADMIN1"), ShouldBeNil)
		So(ValidateKey("manager"), ShouldBeNil)
	})
}
