package vettedposutil

import (
	"go/token"
	"strconv"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewVettedPositionsFromFile(t *testing.T) {

	Convey("NewVettedPositionsFromFile", t, func() {
		p := func(posString string) token.Position {
			parts := strings.Split(posString, ":")
			line, err := strconv.Atoi(parts[1])
			So(err, ShouldBeNil)
			column, err := strconv.Atoi(parts[2])
			So(err, ShouldBeNil)
			return token.Position{
				Filename: parts[0],
				Line:     line,
				Column:   column,
			}
		}

		poss, err := NewVettedPositionsFromFile("testdata/simple.txt")
		So(err, ShouldBeNil)

		So(poss.CheckAndMarkUsed("c", p("/Users/johndoe/path/to/file:2:3")), ShouldBeFalse)
		So(poss.CheckAndMarkUsed("a", p("/Users/johndoe/path/to/file:1:2")), ShouldBeTrue)
		So(poss.CheckAndMarkUsed("a", p("/Users/johndoe/path/to/file:1:2")), ShouldBeTrue)

		So(poss.Err(), ShouldBeError, `unused vetted positions:
/path/to/file:1:2: b
/path/to/file:2:3: a
`)
	})
}
