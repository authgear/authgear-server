package declarative

import (
	"context"
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

// fakeTranslationService is hand-written because the generated mocks are in package declarative_test.
type fakeTranslationService map[string]string

func (s fakeTranslationService) HasKey(ctx context.Context, key string) (bool, error) {
	_, ok := s[key]
	return ok, nil
}

func (s fakeTranslationService) RenderText(ctx context.Context, key string, args any) (string, error) {
	return s[key], nil
}

func TestNewFillFormDataForm(t *testing.T) {
	Convey("NewFillFormDataForm", t, func() {
		ctx := context.Background()
		cfg := newTestFormConfig()
		cfg.UserProfile.CustomAttributes.Attributes = append(cfg.UserProfile.CustomAttributes.Attributes,
			&config.CustomAttributesAttributeConfig{
				ID: "3", Pointer: "/x_plan", Type: config.CustomAttributeTypeEnum, Enum: []string{"free", "pro_plan"},
			},
			&config.CustomAttributesAttributeConfig{
				ID: "4", Pointer: "/x_country", Type: config.CustomAttributeTypeCountryCode,
			},
			&config.CustomAttributesAttributeConfig{
				ID: "5", Pointer: "/x_score", Type: config.CustomAttributeTypeNumber,
				Minimum: new(float64(0)), Maximum: new(float64(10)),
			},
		)
		translations := fakeTranslationService{}
		deps := &authflow.Dependencies{Config: cfg, Translations: translations}

		build := func(user *model.User, fields ...*config.AuthenticationFlowFormField) map[string]FillFormDataField {
			form, err := NewFillFormDataForm(ctx, deps, &config.AuthenticationFlowForm{Name: "names", Fields: fields}, user)
			So(err, ShouldBeNil)
			So(form.Name, ShouldEqual, "names")
			out := map[string]FillFormDataField{}
			for _, f := range form.Fields {
				out[f.Key] = f
			}
			So(len(out), ShouldEqual, len(fields))
			return out
		}
		emptyUser := &model.User{StandardAttributes: map[string]any{}, CustomAttributes: map[string]any{}}

		Convey("types", func() {
			standard := map[string]FillFormFieldType{
				"/name":        FillFormFieldTypeString,
				"/given_name":  FillFormFieldTypeString,
				"/family_name": FillFormFieldTypeString,
				"/middle_name": FillFormFieldTypeString,
				"/nickname":    FillFormFieldTypeString,
				"/gender":      FillFormFieldTypeString,
				"/birthdate":   FillFormFieldTypeDate,
				"/zoneinfo":    FillFormFieldTypeTimezone,
				"/locale":      FillFormFieldTypeEnum,
				"/profile":     FillFormFieldTypeURL,
				"/picture":     FillFormFieldTypeURL,
				"/website":     FillFormFieldTypeURL,
				"/address":     FillFormFieldTypeAddress,
			}
			var fields []*config.AuthenticationFlowFormField
			for pointer := range standard {
				fields = append(fields, userProfileField(pointer, false))
			}
			built := build(emptyUser, fields...)
			for pointer, typ := range standard {
				So(built[pointer].Type, ShouldEqual, typ)
			}

			custom := map[string]FillFormFieldType{
				"/x_terms":       FillFormFieldTypeBoolean,
				"/x_employee_id": FillFormFieldTypeString,
				"/x_age":         FillFormFieldTypeInteger,
				"/x_score":       FillFormFieldTypeNumber,
				"/x_plan":        FillFormFieldTypeEnum,
				"/x_country":     FillFormFieldTypeCountryCode,
			}
			fields = nil
			for pointer := range custom {
				fields = append(fields, userProfileField(pointer, false))
			}
			built = build(emptyUser, fields...)
			for pointer, typ := range custom {
				So(built[pointer].Type, ShouldEqual, typ)
			}
		})

		Convey("label falls back to the attribute name", func() {
			built := build(emptyUser, userProfileField("/given_name", true), userProfileField("/x_employee_id", false))
			So(built["/given_name"].Label, ShouldEqual, "Given Name")
			So(built["/given_name"].Required, ShouldBeTrue)
			So(built["/x_employee_id"].Label, ShouldEqual, "X Employee Id")
			So(built["/x_employee_id"].Required, ShouldBeFalse)
		})

		Convey("label uses the translation", func() {
			translations["standard-attribute-label-/given_name"] = "First name"
			translations["custom-attribute-label-/x_employee_id"] = "Employee ID"
			built := build(emptyUser, userProfileField("/given_name", true), userProfileField("/x_employee_id", false))
			So(built["/given_name"].Label, ShouldEqual, "First name")
			So(built["/x_employee_id"].Label, ShouldEqual, "Employee ID")
		})

		Convey("boolean label is sanitized", func() {
			translations["custom-attribute-label-/x_terms"] = `I accept the <a href="https://example.com/t" onclick="x()">Terms</a><script>x</script>`
			built := build(emptyUser, userProfileField("/x_terms", true))
			So(built["/x_terms"].Label, ShouldEqual,
				`I accept the <a href="https://example.com/t" target="_blank" rel="noopener noreferrer">Terms</a>&lt;script&gt;x&lt;/script&gt;`)
		})

		Convey("current value, including false", func() {
			user := &model.User{
				StandardAttributes: map[string]any{
					"given_name": "John",
					"address":    map[string]any{"locality": "London"},
				},
				CustomAttributes: map[string]any{"x_terms": false},
			}
			built := build(user,
				userProfileField("/given_name", false),
				userProfileField("/family_name", false),
				userProfileField("/address", false),
				userProfileField("/x_terms", false),
			)
			So(built["/given_name"].Value, ShouldEqual, "John")
			So(built["/family_name"].Value, ShouldBeNil)
			So(built["/address"].Value, ShouldResemble, map[string]any{"locality": "London"})
			So(built["/x_terms"].Value, ShouldEqual, false)

			// false is kept in JSON; nil is omitted.
			b, err := json.Marshal(built["/x_terms"])
			So(err, ShouldBeNil)
			So(string(b), ShouldContainSubstring, `"value":false`)
			b, err = json.Marshal(built["/family_name"])
			So(err, ShouldBeNil)
			So(string(b), ShouldNotContainSubstring, `"value"`)
		})

		Convey("gender suggestions", func() {
			built := build(emptyUser, userProfileField("/gender", false))
			So(built["/gender"].Suggestions, ShouldResemble, []string{"male", "female"})
		})

		Convey("enum options fall back and use translations", func() {
			translations["custom-attribute-enum-label-/x_plan-free"] = "Free plan"
			built := build(emptyUser, userProfileField("/x_plan", false))
			So(built["/x_plan"].Options, ShouldResemble, []FillFormDataFieldOption{
				{Value: "free", Label: "Free plan"},
				{Value: "pro_plan", Label: "Pro Plan"},
			})
		})

		Convey("locale options are the supported languages", func() {
			translations["language-zh-HK"] = "繁體中文"
			built := build(emptyUser, userProfileField("/locale", false))
			So(built["/locale"].Options, ShouldResemble, []FillFormDataFieldOption{
				{Value: "en", Label: "en"},
				{Value: "zh-HK", Label: "繁體中文"},
			})
		})

		Convey("constraints", func() {
			Convey("string constraints, and pattern_message only with a translation", func() {
				built := build(emptyUser, userProfileField("/x_employee_id", false), userProfileField("/given_name", false))
				So(built["/x_employee_id"].Constraints, ShouldResemble, &FillFormDataFieldConstraints{Pattern: "^E[0-9]+$"})
				two := 2
				So(built["/given_name"].Constraints, ShouldResemble, &FillFormDataFieldConstraints{MinLength: &two})

				translations["custom-attribute-pattern-error-/x_employee_id"] = "Start with E"
				built = build(emptyUser, userProfileField("/x_employee_id", false))
				So(built["/x_employee_id"].Constraints.PatternMessage, ShouldEqual, "Start with E")
			})

			Convey("minimum and maximum", func() {
				built := build(emptyUser, userProfileField("/x_score", false))
				So(*built["/x_score"].Constraints.Minimum, ShouldEqual, 0)
				So(*built["/x_score"].Constraints.Maximum, ShouldEqual, 10)
			})

			Convey("none", func() {
				built := build(emptyUser, userProfileField("/nickname", false))
				So(built["/nickname"].Constraints, ShouldBeNil)
			})
		})

		Convey("the data has its type", func() {
			b, err := json.Marshal(NewFillFormData(FillFormDataForm{Fields: []FillFormDataField{}}))
			So(err, ShouldBeNil)
			So(string(b), ShouldEqualJSON, `{"type": "fill_form_data", "form": {"fields": []}}`)
		})
	})
}
