package elasticsearch

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestEscapeRegexp(t *testing.T) {
	Convey("EscapeRegexp", t, func() {
		literal := `.?+*|{}[]()"\#@&<>~`
		actual := EscapeRegexp(literal)
		expected := `\.\?\+\*\|\{\}\[\]\(\)\"\\\#\@\&\<\>\~`
		So(actual, ShouldEqual, expected)
	})
}
