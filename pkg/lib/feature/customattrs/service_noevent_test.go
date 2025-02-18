package customattrs

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/authn/customattrs"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func TestServiceNoEvent(t *testing.T) {
	Convey("ServiceNoEvent", t, func() {
		Convey("fromStorageForm", func() {
			s := &ServiceNoEvent{
				Config: &config.UserProfileConfig{
					CustomAttributes: &config.CustomAttributesConfig{
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
				},
			}

			Convey("transform to representation from", func() {
				actual, err := s.fromStorageForm(map[string]interface{}{
					"0000": "a",
					"0001": "b",
				})
				So(err, ShouldBeNil)
				So(actual, ShouldResemble, customattrs.T{
					"a": "a",
					"b": "b",
				})
			})

			Convey("ignore unknown attributes", func() {
				actual, err := s.fromStorageForm(map[string]interface{}{
					"0002": "c",
				})
				So(err, ShouldBeNil)
				So(actual, ShouldResemble, customattrs.T{})
			})
		})

		Convey("toStorageForm", func() {
			s := &ServiceNoEvent{
				Config: &config.UserProfileConfig{
					CustomAttributes: &config.CustomAttributesConfig{
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
				},
			}

			Convey("transform to storage form", func() {
				actual, err := s.toStorageForm(customattrs.T{
					"a": "a",
					"b": "b",
				})
				So(err, ShouldBeNil)
				So(actual, ShouldResemble, map[string]interface{}{
					"0000": "a",
					"0001": "b",
				})
			})

			Convey("ignore absent attributes", func() {
				actual, err := s.toStorageForm(customattrs.T{
					"a": "a",
				})
				So(err, ShouldBeNil)
				So(actual, ShouldResemble, map[string]interface{}{
					"0000": "a",
				})
			})
		})

		Convey("generateSchemaString", func() {
			newFloat := func(f float64) *float64 {
				return &f
			}

			s := &ServiceNoEvent{
				Config: &config.UserProfileConfig{
					CustomAttributes: &config.CustomAttributesConfig{

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
								Type:    "country_code",
							},
						},
					},
				},
			}

			test := func(pointers []string, schemaStr string) {
				actual, err := s.generateSchemaString(pointers)
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
            "type": "string",
            "minLength": 1
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
            "type": "string",
	    "minLength": 1
        }
    }
}
			`)
		})

		Convey("validate", func() {
			newFloat := func(f float64) *float64 {
				return &f
			}

			s := &ServiceNoEvent{
				Config: &config.UserProfileConfig{
					CustomAttributes: &config.CustomAttributesConfig{
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
								Type:    "country_code",
							},
						},
					},
				},
			}

			test := func(pointers []string, value map[string]interface{}, errStr string) {
				err := s.validate(context.Background(), pointers, customattrs.T(value))
				if errStr == "" {
					So(err, ShouldBeNil)
				} else {
					So(err, ShouldBeError, errStr)
				}
			}

			// Validate the only invalid value.
			test([]string{"/number"}, map[string]interface{}{
				"number": 3,
			}, `invalid value:
/number: maximum
  map[actual:3 maximum:2]`)

			// Only validate listed pointers.
			test([]string{"/number"}, map[string]interface{}{
				"number":  3, // invalid, should be validated.
				"integer": 0, // invalid, but should not be validated
			}, `invalid value:
/number: maximum
  map[actual:3 maximum:2]`)

			// Validate all invalid values.
			test([]string{"/number", "/integer"}, map[string]interface{}{
				"number":  3, // invalid, should be validated.
				"integer": 0, // invalid, should be validated.
			}, `invalid value:
/integer: minimum
  map[actual:0 minimum:3]
/number: maximum
  map[actual:3 maximum:2]`)

			// Validate enum
			test([]string{"/enum"}, map[string]interface{}{
				"enum": "foobar",
			}, `invalid value:
/enum: enum
  map[actual:foobar expected:[a b]]`)
			test([]string{"/enum"}, map[string]interface{}{
				"enum": "a",
			}, ``)

			// Validate phone number
			test([]string{"/x_phone"}, map[string]interface{}{
				"x_phone": "foobar",
			}, `invalid value:
/x_phone: format
  map[error:not in E.164 format format:phone]`)
			test([]string{"/x_phone"}, map[string]interface{}{
				"x_phone": "+85298765432",
			}, ``)

			// Validate email
			test([]string{"/x_email"}, map[string]interface{}{
				"x_email": "foobar",
			}, `invalid value:
/x_email: format
  map[error:invalid email address: mail: missing '@' or angle-addr format:email]`)
			test([]string{"/x_email"}, map[string]interface{}{
				"x_email": "user@example.com",
			}, ``)

			// Validate url
			test([]string{"/x_url"}, map[string]interface{}{
				"x_url": "foobar",
			}, `invalid value:
/x_url: format
  map[error:input URL must be absolute format:uri]`)
			test([]string{"/x_url"}, map[string]interface{}{
				"x_url": "http://127.0.0.1",
			}, ``)

			// Validate alpha2
			test([]string{"/alpha2"}, map[string]interface{}{
				"alpha2": "foobar",
			}, `invalid value:
/alpha2: format
  map[error:invalid ISO 3166-1 alpha-2 code: "foobar" format:iso3166-1-alpha-2]`)
			test([]string{"/x_url"}, map[string]interface{}{
				"alpha2": "US",
			}, ``)

		})
	})
}
