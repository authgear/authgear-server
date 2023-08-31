package attrs

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

func TestAddAbsent(t *testing.T) {
	Convey("AddAbsent", t, func() {
		test := func(l List, allAbsent []string, expected List) {
			actual := l.AddAbsent(allAbsent)
			So(actual, ShouldResemble, expected)
		}

		test(nil, nil, List{})
		test(nil, []string{"/given_name"}, List{
			{Pointer: "/given_name", Value: nil},
		})
		test(List{
			{Pointer: "/family_name", Value: "doe"},
		}, nil, List{
			{Pointer: "/family_name", Value: "doe"},
		})
		test(List{
			{Pointer: "/family_name", Value: "doe"},
		}, []string{"/given_name"}, List{
			{Pointer: "/family_name", Value: "doe"},
			{Pointer: "/given_name", Value: nil},
		})
	})
}

func TestSeparate(t *testing.T) {
	Convey("Separate", t, func() {
		type result struct {
			StdAttrs     List
			CustomAttrs  List
			UnknownAttrs List
		}

		test := func(l List, customAttrsPointers []string, expected result) {
			cfg := &config.UserProfileConfig{
				CustomAttributes: &config.CustomAttributesConfig{},
			}
			for _, p := range customAttrsPointers {
				cfg.CustomAttributes.Attributes = append(cfg.CustomAttributes.Attributes, &config.CustomAttributesAttributeConfig{
					Pointer: p,
				})
			}

			stdAttrs, customAttrs, unknownAttrs := l.Separate(cfg)
			So(stdAttrs, ShouldResemble, expected.StdAttrs)
			So(customAttrs, ShouldResemble, expected.CustomAttrs)
			So(unknownAttrs, ShouldResemble, expected.UnknownAttrs)
		}

		test(nil, nil, result{})
		test(nil, []string{"/x_age"}, result{})

		test(List{
			{Pointer: "/given_name", Value: "john"},
			{Pointer: "/x_age", Value: "42"},
			{Pointer: "/unknown", Value: nil},
		}, []string{"/x_age"}, result{
			StdAttrs: List{
				{Pointer: "/given_name", Value: "john"},
			},
			CustomAttrs: List{
				{Pointer: "/x_age", Value: "42"},
			},
			UnknownAttrs: List{
				{Pointer: "/unknown", Value: nil},
			},
		})
	})
}
