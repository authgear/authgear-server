package skydb

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMalformedPredicate(t *testing.T) {
	Convey("Predicate with IN", t, func() {
		Convey("keypath operand types", func() {
			predicate := Predicate{
				Operator: In,
				Children: []interface{}{
					Expression{
						Type:  KeyPath,
						Value: "categories",
					},
					Expression{
						Type:  KeyPath,
						Value: "favoriteCategory",
					},
				},
			}
			err := predicate.Validate()
			So(err, ShouldNotBeNil)
		})

		Convey("string operand types", func() {
			predicate := Predicate{
				Operator: In,
				Children: []interface{}{
					Expression{
						Type:  Literal,
						Value: "interesting",
					},
					Expression{
						Type:  Literal,
						Value: "interesting",
					},
				},
			}
			err := predicate.Validate()
			So(err, ShouldNotBeNil)
		})

		Convey("literal array on left hand side", func() {
			predicate := Predicate{
				Operator: In,
				Children: []interface{}{
					Expression{
						Type: Literal,
						Value: []interface{}{
							"interesting",
							"funny",
						},
					},
					Expression{
						Type:  KeyPath,
						Value: "category",
					},
				},
			}
			err := predicate.Validate()
			So(err, ShouldNotBeNil)
		})

		Convey("literal string on right hand side", func() {
			predicate := Predicate{
				Operator: In,
				Children: []interface{}{
					Expression{
						Type:  KeyPath,
						Value: "category",
					},
					Expression{
						Type:  Literal,
						Value: "interesting",
					},
				},
			}
			err := predicate.Validate()
			So(err, ShouldNotBeNil)
		})
	})
}
