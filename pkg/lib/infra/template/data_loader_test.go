package template

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDataLoader(t *testing.T) {
	Convey("DataLoader", t, func() {
		loader := &DataLoader{}
		cases := []struct {
			Content string
		}{
			{""},
			{"a"},
			{"with space"},
			{`<!DOCTYPE html><html></html>`},
		}
		for _, c := range cases {
			actual, err := loader.Load(DataURIWithContent(c.Content))
			So(err, ShouldBeNil)
			So(actual, ShouldEqual, c.Content)
		}
	})
}
