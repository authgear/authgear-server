package rolesgroups

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestValidateKey(t *testing.T) {
	ctx := context.Background()
	Convey("ValidateKey", t, func() {
		So(ValidateKey(ctx, ""), ShouldBeError, `invalid value:
<root>: minLength
  map[actual:0 expected:1]
<root>: pattern
  map[actual: expected:^[a-zA-Z_][a-zA-Z0-9:_]*$]`)

		So(ValidateKey(ctx, "a0123456789012345678901234567890123456789"), ShouldBeError, `invalid value:
<root>: maxLength
  map[actual:41 expected:40]`)

		So(ValidateKey(ctx, "1"), ShouldBeError, `invalid value:
<root>: pattern
  map[actual:1 expected:^[a-zA-Z_][a-zA-Z0-9:_]*$]`)

		So(ValidateKey(ctx, "#$%^&*AND*&&^%$#"), ShouldBeError, `invalid value:
<root>: pattern
  map[actual:#$%^&*AND*&&^%$# expected:^[a-zA-Z_][a-zA-Z0-9:_]*$]`)

		So(ValidateKey(ctx, "user#123ok"), ShouldBeError, `invalid value:
<root>: pattern
  map[actual:user#123ok expected:^[a-zA-Z_][a-zA-Z0-9:_]*$]`)

		So(ValidateKey(ctx, "GOOD key"), ShouldBeError, `invalid value:
<root>: pattern
  map[actual:GOOD key expected:^[a-zA-Z_][a-zA-Z0-9:_]*$]`)

		So(ValidateKey(ctx, "0admin"), ShouldBeError, `invalid value:
<root>: pattern
  map[actual:0admin expected:^[a-zA-Z_][a-zA-Z0-9:_]*$]`)

		So(ValidateKey(ctx, "authgear:user"), ShouldBeError, `invalid value:
<root>: format
  map[error:key cannot start with the preserved prefix: `+"`"+`authgear:`+"`"+` format:x_role_group_key]`)

		So(ValidateKey(ctx, "manager:SUPER_ADMIN1"), ShouldBeNil)
		So(ValidateKey(ctx, "manager"), ShouldBeNil)
	})
}
