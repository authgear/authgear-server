package customattrs

import (
	"testing"

	"github.com/authgear/authgear-server/pkg/lib/config"
	. "github.com/smartystreets/goconvey/convey"
)

func TestService(t *testing.T) {
	Convey("Service", t, func() {
		Convey("FromStorageForm", func() {
			s := &Service{
				Config: &config.CustomAttributesConfig{
					Attributes: []*config.CustomAttributesAttributeConfig{
						&config.CustomAttributesAttributeConfig{
							ID:      "0000",
							Pointer: "/a",
							Type:    "string",
						},
						&config.CustomAttributesAttributeConfig{
							ID:      "0001",
							Pointer: "/b",
							Type:    "string",
						},
					},
				},
			}

			Convey("transform to representation from", func() {
				actual, err := s.FromStorageForm(map[string]interface{}{
					"0000": "a",
					"0001": "b",
				})
				So(err, ShouldBeNil)
				So(actual, ShouldResemble, T{
					"a": "a",
					"b": "b",
				})
			})

			Convey("ignore unknown attributes", func() {
				actual, err := s.FromStorageForm(map[string]interface{}{
					"0002": "c",
				})
				So(err, ShouldBeNil)
				So(actual, ShouldResemble, T{})
			})
		})
	})
}
