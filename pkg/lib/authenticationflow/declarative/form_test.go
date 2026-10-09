package declarative

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func newTestFormConfig() *config.AppConfig {
	two := 2
	cfg := &config.AppConfig{
		Localization: &config.LocalizationConfig{
			SupportedLanguages: []string{"en", "zh-HK"},
		},
		UserProfile: &config.UserProfileConfig{
			StandardAttributes: &config.StandardAttributesConfig{
				Validation: []*config.StandardAttributesValidationConfig{
					{Pointer: "/given_name", MinLength: &two},
					{Pointer: "/address/locality", AllowedCharacters: []config.AllowedCharacters{config.AllowedCharactersLetters}},
				},
			},
			CustomAttributes: &config.CustomAttributesConfig{
				Attributes: []*config.CustomAttributesAttributeConfig{
					{ID: "0", Pointer: "/x_terms", Type: config.CustomAttributeTypeBoolean},
					{ID: "1", Pointer: "/x_employee_id", Type: config.CustomAttributeTypeString, Pattern: "^E[0-9]+$"},
					{ID: "2", Pointer: "/x_age", Type: config.CustomAttributeTypeInteger},
				},
			},
		},
	}
	return cfg
}

func newTestForm(fields ...*config.AuthenticationFlowFormField) *config.AuthenticationFlowForm {
	return &config.AuthenticationFlowForm{Fields: fields}
}

func userProfileField(pointer string, required bool) *config.AuthenticationFlowFormField {
	return &config.AuthenticationFlowFormField{
		UserProfile: &config.AuthenticationFlowFormFieldUserProfile{Pointer: pointer, Required: required},
	}
}

func TestBuildFormFieldsSchema(t *testing.T) {
	Convey("buildFormFieldsSchema", t, func() {
		cfg := newTestFormConfig()

		parse := func(form *config.AuthenticationFlowForm, raw string) error {
			s := &InputSchemaFillForm{Config: cfg, Form: form}
			_, err := s.MakeInput(context.Background(), json.RawMessage(raw))
			return err
		}

		Convey("required string", func() {
			form := newTestForm(userProfileField("/given_name", true), userProfileField("/nickname", false))

			So(parse(form, `{"fields": [{"key": "/given_name", "value": "John"}]}`), ShouldBeNil)
			So(parse(form, `{"fields": [{"key": "/given_name", "value": "John"}, {"key": "/nickname", "value": "J"}]}`), ShouldBeNil)

			// missing
			So(parse(form, `{"fields": []}`), ShouldNotBeNil)
			So(parse(form, `{"fields": [{"key": "/nickname", "value": "J"}]}`), ShouldNotBeNil)
			// empty
			So(parse(form, `{"fields": [{"key": "/given_name", "value": ""}]}`), ShouldNotBeNil)
			// wrong type
			So(parse(form, `{"fields": [{"key": "/given_name", "value": 1}]}`), ShouldNotBeNil)
		})

		Convey("optional fields may be absent", func() {
			form := newTestForm(userProfileField("/nickname", false))
			So(parse(form, `{"fields": []}`), ShouldBeNil)
		})

		Convey("unknown key", func() {
			form := newTestForm(userProfileField("/given_name", false))
			So(parse(form, `{"fields": [{"key": "/family_name", "value": "Doe"}]}`), ShouldNotBeNil)
		})

		Convey("unknown property of the input and of an item", func() {
			form := newTestForm(userProfileField("/given_name", false))
			So(parse(form, `{"fields": [], "foo": 1}`), ShouldNotBeNil)
			So(parse(form, `{"fields": [{"key": "/given_name", "value": "J", "foo": 1}]}`), ShouldNotBeNil)
		})

		Convey("required boolean must be true", func() {
			form := newTestForm(userProfileField("/x_terms", true))
			So(parse(form, `{"fields": [{"key": "/x_terms", "value": true}]}`), ShouldBeNil)
			So(parse(form, `{"fields": [{"key": "/x_terms", "value": false}]}`), ShouldNotBeNil)
			So(parse(form, `{"fields": []}`), ShouldNotBeNil)
			So(parse(form, `{"fields": [{"key": "/x_terms", "value": "true"}]}`), ShouldNotBeNil)
		})

		Convey("optional boolean accepts false", func() {
			form := newTestForm(userProfileField("/x_terms", false))
			So(parse(form, `{"fields": [{"key": "/x_terms", "value": false}]}`), ShouldBeNil)
		})

		Convey("required address", func() {
			form := newTestForm(userProfileField("/address", true))
			So(parse(form, `{"fields": [{"key": "/address", "value": {"locality": "London"}}]}`), ShouldBeNil)
			So(parse(form, `{"fields": [{"key": "/address", "value": {}}]}`), ShouldNotBeNil)
			So(parse(form, `{"fields": [{"key": "/address", "value": {"foo": "x"}}]}`), ShouldNotBeNil)
			So(parse(form, `{"fields": [{"key": "/address", "value": {"locality": ""}}]}`), ShouldNotBeNil)
			So(parse(form, `{"fields": []}`), ShouldNotBeNil)
		})

		Convey("optional address accepts an empty object", func() {
			form := newTestForm(userProfileField("/address", false))
			So(parse(form, `{"fields": [{"key": "/address", "value": {}}]}`), ShouldBeNil)
		})

		Convey("optional address rejects unknown keys", func() {
			form := newTestForm(userProfileField("/address", false))
			So(parse(form, `{"fields": [{"key": "/address", "value": {"locality": "London", "foo": "x"}}]}`), ShouldNotBeNil)
		})

		Convey("locale is limited to the supported languages", func() {
			form := newTestForm(userProfileField("/locale", false))
			So(parse(form, `{"fields": [{"key": "/locale", "value": "zh-HK"}]}`), ShouldBeNil)
			So(parse(form, `{"fields": [{"key": "/locale", "value": "fr"}]}`), ShouldNotBeNil)
		})

		Convey("custom attribute types", func() {
			form := newTestForm(userProfileField("/x_age", false))
			So(parse(form, `{"fields": [{"key": "/x_age", "value": 18}]}`), ShouldBeNil)
			So(parse(form, `{"fields": [{"key": "/x_age", "value": 1.5}]}`), ShouldNotBeNil)
		})

		Convey("the shared address schema is not mutated", func() {
			form := newTestForm(userProfileField("/address", true))
			_ = parse(form, `{"fields": []}`)
			_, ok := stdattrs.SchemaBuilderAddress()["additionalProperties"]
			So(ok, ShouldBeFalse)
		})
	})
}

func TestValidateFormFields(t *testing.T) {
	Convey("validateFormFields", t, func() {
		cfg := newTestFormConfig()

		run := func(fields ...FormFieldInput) []validation.Error {
			vctx := &validation.Context{}
			validateFormFields(vctx, cfg, fields)
			err := vctx.Error("invalid form")
			if err == nil {
				return nil
			}
			var aErr *validation.AggregatedError
			So(errors.As(err, &aErr), ShouldBeTrue)
			return aErr.Errors
		}

		Convey("duplicated key", func() {
			errs := run(
				FormFieldInput{Key: "/nickname", Value: "a"},
				FormFieldInput{Key: "/nickname", Value: "b"},
			)
			So(errs, ShouldHaveLength, 1)
			So(errs[0].Location, ShouldEqual, "/fields/1/key")
		})

		Convey("standard attribute constraints", func() {
			errs := run(FormFieldInput{Key: "/given_name", Value: "J"})
			So(errs, ShouldHaveLength, 1)
			So(errs[0].Keyword, ShouldEqual, "minLength")
			So(errs[0].Location, ShouldEqual, "/fields/0/value")
		})

		Convey("custom attribute constraints", func() {
			errs := run(
				FormFieldInput{Key: "/nickname", Value: "ok"},
				FormFieldInput{Key: "/x_employee_id", Value: "12345"},
			)
			So(errs, ShouldHaveLength, 1)
			So(errs[0].Keyword, ShouldEqual, "pattern")
			So(errs[0].Location, ShouldEqual, "/fields/1/value")
		})

		Convey("address sub-attribute constraints", func() {
			errs := run(FormFieldInput{Key: "/address", Value: map[string]any{"locality": "L0ndon", "region": "x"}})
			So(errs, ShouldHaveLength, 1)
			So(errs[0].Keyword, ShouldEqual, "allowedCharacters")
			So(errs[0].Location, ShouldEqual, "/fields/0/value/locality")
		})

		Convey("a valid form", func() {
			So(run(
				FormFieldInput{Key: "/given_name", Value: "John"},
				FormFieldInput{Key: "/x_employee_id", Value: "E123"},
				FormFieldInput{Key: "/x_terms", Value: true},
			), ShouldBeNil)
		})
	})
}

func TestFormFieldsToAttrs(t *testing.T) {
	Convey("formFieldsToAttrs", t, func() {
		cfg := newTestFormConfig()
		form := newTestForm(
			userProfileField("/given_name", false),
			userProfileField("/nickname", false),
			userProfileField("/x_terms", false),
			userProfileField("/x_age", false),
		)

		std, custom := formFieldsToAttrs(cfg, form, []FormFieldInput{
			{Key: "/x_terms", Value: false},
			{Key: "/given_name", Value: "John"},
		})

		// In form order. A field absent from the input gets nil, so it is cleared.
		So(len(std), ShouldEqual, 2)
		So(std[0].Pointer, ShouldEqual, "/given_name")
		So(std[0].Value, ShouldEqual, "John")
		So(std[1].Pointer, ShouldEqual, "/nickname")
		So(std[1].Value, ShouldBeNil)

		So(len(custom), ShouldEqual, 2)
		So(custom[0].Pointer, ShouldEqual, "/x_terms")
		So(custom[0].Value, ShouldEqual, false)
		So(custom[1].Pointer, ShouldEqual, "/x_age")
		So(custom[1].Value, ShouldBeNil)
	})
}
