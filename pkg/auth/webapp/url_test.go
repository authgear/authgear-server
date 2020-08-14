package webapp

import (
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRemoveX(t *testing.T) {
	Convey("RemoveX", t, func() {
		q := url.Values{
			"a":   []string{"b"},
			"x_a": []string{"b"},
		}
		RemoveX(q)
		So(q.Encode(), ShouldEqual, "a=b")
	})
}
