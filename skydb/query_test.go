package skydb

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestExpression(t *testing.T) {
	Convey("is empty", t, func() {
		expr := Expression{}
		So(expr.IsEmpty(), ShouldBeTrue)
	})
}

func TestMalformedPredicate(t *testing.T) {
	Convey("Predicate with Equal", t, func() {
		Convey("comparing array", func() {
			predicate := Predicate{
				Operator: Equal,
				Children: []interface{}{
					Expression{
						Type:  KeyPath,
						Value: "categories",
					},
					Expression{
						Type:  Literal,
						Value: []interface{}{},
					},
				},
			}
			err := predicate.Validate()
			So(err, ShouldNotBeNil)
		})

		Convey("comparing map", func() {
			predicate := Predicate{
				Operator: Equal,
				Children: []interface{}{
					Expression{
						Type:  KeyPath,
						Value: "categories",
					},
					Expression{
						Type:  Literal,
						Value: map[string]interface{}{},
					},
				},
			}
			err := predicate.Validate()
			So(err, ShouldNotBeNil)
		})
	})

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

	Convey("Predicate with User Discover", t, func() {
		Convey("cannot be combined", func() {
			predicate := Predicate{
				Not,
				[]interface{}{
					Predicate{
						Functional,
						[]interface{}{
							Expression{
								Type: Function,
								Value: UserDiscoverFunc{
									Emails: []string{
										"john.doe@example.com",
										"jane.doe@example.com",
									},
								},
							},
						},
					},
				},
			}

			err := predicate.Validate()
			So(err, ShouldNotBeNil)
		})
	})
}
