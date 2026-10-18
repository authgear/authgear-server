package customattrs

import (
	"context"
	"errors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/authn/customattrs"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type fakeUserQueries struct {
	custom map[string]any
}

func (q *fakeUserQueries) GetRaw(ctx context.Context, userID string) (*user.User, error) {
	return &user.User{CustomAttributes: q.custom}, nil
}

type fakeUserStore struct {
	stored map[string]any
}

func (s *fakeUserStore) UpdateCustomAttributes(ctx context.Context, userID string, storageForm map[string]any) error {
	s.stored = storageForm
	return nil
}

func TestServiceNoEventStringConstraints(t *testing.T) {
	Convey("ServiceNoEvent string constraints", t, func() {
		two := 2
		store := &fakeUserStore{}
		queries := &fakeUserQueries{custom: map[string]any{}}
		s := &ServiceNoEvent{
			Config: &config.UserProfileConfig{
				CustomAttributes: &config.CustomAttributesConfig{
					Attributes: []*config.CustomAttributesAttributeConfig{
						{ID: "0000", Pointer: "/employee_id", Type: config.CustomAttributeTypeString, Pattern: "^E[0-9]+$", MinLength: &two},
						{ID: "0001", Pointer: "/other", Type: config.CustomAttributeTypeString},
						{ID: "0002", Pointer: "/accepted", Type: config.CustomAttributeTypeBoolean},
					},
				},
			},
			UserQueries: queries,
			UserStore:   store,
		}
		for _, c := range s.Config.CustomAttributes.Attributes {
			c.SetDefaults()
		}
		ctx := context.Background()
		update := func(form map[string]string) error {
			return s.UpdateCustomAttributesWithForm(ctx, accesscontrol.RoleGreatest, "u", form)
		}

		Convey("checks a changed value", func() {
			err := update(map[string]string{"/employee_id": "12345"})
			var aErr *validation.AggregatedError
			So(err, ShouldNotBeNil)
			So(errors.As(err, &aErr), ShouldBeTrue)
			So(aErr.Errors, ShouldHaveLength, 1)
			So(aErr.Errors[0].Keyword, ShouldEqual, "pattern")
			So(aErr.Errors[0].Location, ShouldEqual, "/employee_id")
			So(store.stored, ShouldBeNil)

			So(update(map[string]string{"/employee_id": "E12345"}), ShouldBeNil)
			So(store.stored["0000"], ShouldEqual, "E12345")
		})

		Convey("passes an unchanged violating value", func() {
			queries.custom = map[string]any{"0000": "12345"}
			So(update(map[string]string{"/employee_id": "12345", "/other": "x"}), ShouldBeNil)
			So(store.stored["0001"], ShouldEqual, "x")
		})

		Convey("does not check a deletion", func() {
			queries.custom = map[string]any{"0000": "12345"}
			So(update(map[string]string{"/employee_id": ""}), ShouldBeNil)
		})

		Convey("writes a boolean from a form value", func() {
			So(update(map[string]string{"/accepted": "false"}), ShouldBeNil)
			So(store.stored["0002"], ShouldEqual, false)
			So(update(map[string]string{"/accepted": "true"}), ShouldBeNil)
			So(store.stored["0002"], ShouldEqual, true)
		})
	})
}

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
				actual, err := s.fromStorageForm(map[string]any{
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
				actual, err := s.fromStorageForm(map[string]any{
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
				So(actual, ShouldResemble, map[string]any{
					"0000": "a",
					"0001": "b",
				})
			})

			Convey("ignore absent attributes", func() {
				actual, err := s.toStorageForm(customattrs.T{
					"a": "a",
				})
				So(err, ShouldBeNil)
				So(actual, ShouldResemble, map[string]any{
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

			test := func(pointers []string, value map[string]any, errStr string) {
				err := s.validate(context.Background(), pointers, customattrs.T(value))
				if errStr == "" {
					So(err, ShouldBeNil)
				} else {
					So(err, ShouldBeError, errStr)
				}
			}

			// Validate the only invalid value.
			test([]string{"/number"}, map[string]any{
				"number": 3,
			}, `invalid value:
/number: maximum
  map[actual:3 maximum:2]`)

			// Only validate listed pointers.
			test([]string{"/number"}, map[string]any{
				"number":  3, // invalid, should be validated.
				"integer": 0, // invalid, but should not be validated
			}, `invalid value:
/number: maximum
  map[actual:3 maximum:2]`)

			// Validate all invalid values.
			test([]string{"/number", "/integer"}, map[string]any{
				"number":  3, // invalid, should be validated.
				"integer": 0, // invalid, should be validated.
			}, `invalid value:
/integer: minimum
  map[actual:0 minimum:3]
/number: maximum
  map[actual:3 maximum:2]`)

			// Validate enum
			test([]string{"/enum"}, map[string]any{
				"enum": "foobar",
			}, `invalid value:
/enum: enum
  map[actual:foobar expected:[a b]]`)
			test([]string{"/enum"}, map[string]any{
				"enum": "a",
			}, ``)

			// Validate phone number
			test([]string{"/x_phone"}, map[string]any{
				"x_phone": "foobar",
			}, `invalid value:
/x_phone: format
  map[error:not in E.164 format format:phone]`)
			test([]string{"/x_phone"}, map[string]any{
				"x_phone": "+85298765432",
			}, ``)

			// Validate email
			test([]string{"/x_email"}, map[string]any{
				"x_email": "foobar",
			}, `invalid value:
/x_email: format
  map[error:invalid email address: mail: missing '@' or angle-addr format:email]`)
			test([]string{"/x_email"}, map[string]any{
				"x_email": "user@example.com",
			}, ``)

			// Validate url
			test([]string{"/x_url"}, map[string]any{
				"x_url": "foobar",
			}, `invalid value:
/x_url: format
  map[error:input URL must be absolute format:uri]`)
			test([]string{"/x_url"}, map[string]any{
				"x_url": "http://127.0.0.1",
			}, ``)

			// Validate alpha2
			test([]string{"/alpha2"}, map[string]any{
				"alpha2": "foobar",
			}, `invalid value:
/alpha2: format
  map[error:invalid ISO 3166-1 alpha-2 code: "foobar" format:iso3166-1-alpha-2]`)
			test([]string{"/x_url"}, map[string]any{
				"alpha2": "US",
			}, ``)

		})
	})
}
