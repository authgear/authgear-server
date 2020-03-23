package audit

import (
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	. "github.com/skygeario/skygear-server/pkg/core/skytest"
)

func TestPasswordViolation(t *testing.T) {
	Convey("PasswordViolation", t, func() {
		v := PasswordViolation{
			Reason: PasswordTooShort,
			Info: map[string]interface{}{
				"min_length": 8,
				"pw_length":  6,
			},
		}
		b, err := json.Marshal(v)
		So(err, ShouldBeNil)
		So(b, ShouldEqualJSON, `{"kind":"PasswordTooShort","min_length":8,"pw_length":6}`)
	})
}
