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

		Convey("GenerateSchemaString", func() {
			newFloat := func(f float64) *float64 {
				return &f
			}

			s := &Service{
				Config: &config.CustomAttributesConfig{
					Attributes: []*config.CustomAttributesAttributeConfig{
						&config.CustomAttributesAttributeConfig{
							Pointer: "/string",
							Type:    "string",
						},
						&config.CustomAttributesAttributeConfig{
							Pointer: "/number",
							Type:    "number",
							Minimum: newFloat(1),
							Maximum: newFloat(2),
						},
						&config.CustomAttributesAttributeConfig{
							Pointer: "/integer",
							Type:    "integer",
							Minimum: newFloat(3),
							Maximum: newFloat(4),
						},
						&config.CustomAttributesAttributeConfig{
							Pointer: "/enum",
							Type:    "enum",
							Enum:    []string{"a", "b"},
						},
						&config.CustomAttributesAttributeConfig{
							Pointer: "/x_phone",
							Type:    "phone_number",
						},
						&config.CustomAttributesAttributeConfig{
							Pointer: "/x_email",
							Type:    "email",
						},
						&config.CustomAttributesAttributeConfig{
							Pointer: "/x_url",
							Type:    "url",
						},
						&config.CustomAttributesAttributeConfig{
							Pointer: "/alpha2",
							Type:    "alpha2",
						},
					},
				},
			}

			test := func(pointers []string, schemaStr string) {
				actual, err := s.GenerateSchemaString(pointers)
				So(err, ShouldBeNil)
				So(actual, ShouldEqualJSON, schemaStr)
			}

			test([]string{
				"/string",
				"/number",
				"/integer",
				"/enum",
				"/x_phone",
				"/x_email",
				"/x_url",
				"/alpha2",
			}, `
{
    "type": "object",
    "properties": {
        "string": {
            "type": "string"
        },
        "number": {
            "maximum": 2,
            "minimum": 1,
            "type": "number"
        },
        "integer": {
            "maximum": 4,
            "minimum": 3,
            "type": "integer"
        },
        "enum": {
            "enum": [
                "a",
                "b"
            ],
            "type": "string"
        },
        "x_email": {
            "format": "email",
            "type": "string"
        },
        "x_phone": {
            "format": "phone",
            "type": "string"
        },
        "x_url": {
            "format": "uri",
            "type": "string"
        },
        "alpha2": {
            "format": "iso3166-1-alpha-2",
            "type": "string"
        }
    }
}
			`)

			test(nil, `
{
    "type": "object",
    "properties": {
    }
}
			`)

			test([]string{"/string"}, `
{
    "type": "object",
    "properties": {
        "string": {
            "type": "string"
        }
    }
}
			`)
		})
	})
}
