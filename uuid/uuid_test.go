package uuid

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"

	"fmt"
	"regexp"
)

func TestNew(t *testing.T) {
	Convey("uuid", t, func() {
		uuid := New()

		Convey("is in format xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx", func() {
			const uuid4Pattern = `[a-f0-9]{8}-[a-f0-9]{4}-4[a-f0-9]{3}-[89ab][a-f0-9]{3}-[a-f0-9]{12}`
			matched, err := regexp.MatchString(uuid4Pattern, uuid)

			So(err, ShouldBeNil)
			So(matched, ShouldBeTrue)
		})
	})
}
