package sso

import (
	"testing"

	"github.com/lestrrat-go/jwx/jwk"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/clock"
)

func TestStateCodec(t *testing.T) {

	Convey("StateCodec", t, func() {
		key, err := jwk.New([]byte("secret"))
		So(err, ShouldBeNil)

		set := jwk.Set{Keys: []jwk.Key{key}}
		codec := &StateCodec{
			AppID: "app",
			Clock: clock.NewMockClock(),
			Credentials: &config.JWTKeyMaterials{
				Set: set,
			},
		}

		state := State{
			UserID: "user",
			Extra: map[string]string{
				"foo": "bar",
			},
			Action:      "action",
			HashedNonce: "nonce",
		}
		out, err := codec.EncodeState(state)
		So(err, ShouldBeNil)

		decoded, err := codec.DecodeState(out)
		So(err, ShouldBeNil)

		So(decoded, ShouldResemble, &state)
	})
}
