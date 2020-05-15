package interaction

import (
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestInteractionJSONSerialization(t *testing.T) {
	Convey("InteractionJSONSerialization", t, func() {
		Convey("private fields are not serialized", func() {
			i := &Interaction{
				Intent: &IntentLogin{},
			}
			i.committed = true
			b, err := json.Marshal(i)
			So(err, ShouldBeNil)

			var ii Interaction
			err = json.Unmarshal(b, &ii)
			So(err, ShouldBeNil)

			So(ii.committed, ShouldBeFalse)
		})
	})
}
