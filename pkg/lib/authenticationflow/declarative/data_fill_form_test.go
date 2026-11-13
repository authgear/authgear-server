package declarative

import (
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func TestNewFillFormDataForm(t *testing.T) {
	Convey("NewFillFormDataForm", t, func() {
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
		deps := &authflow.Dependencies{Config: cfg}

		build := func(user *model.User, fields ...*config.AuthenticationFlowFormField) map[string]FillFormDataField {
			form := NewFillFormDataForm(deps, &config.AuthenticationFlowForm{Name: "names", Fields: fields}, user)
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

		Convey("required", func() {
			built := build(emptyUser, userProfileField("/given_name", true), userProfileField("/x_employee_id", false))
			So(built["/given_name"].Required, ShouldBeTrue)
			So(built["/x_employee_id"].Required, ShouldBeFalse)
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

		Convey("enum options are the enum values", func() {
			built := build(emptyUser, userProfileField("/x_plan", false))
			So(built["/x_plan"].Options, ShouldResemble, []FillFormDataFieldOption{
				{Value: "free"},
				{Value: "pro_plan"},
			})
		})

		Convey("locale options are the supported languages", func() {
			built := build(emptyUser, userProfileField("/locale", false))
			So(built["/locale"].Options, ShouldResemble, []FillFormDataFieldOption{
				{Value: "en"},
				{Value: "zh-HK"},
			})
		})

		Convey("constraints", func() {
			Convey("string constraints", func() {
				built := build(emptyUser, userProfileField("/x_employee_id", false), userProfileField("/given_name", false))
				So(built["/x_employee_id"].Constraints, ShouldResemble, &FillFormDataFieldConstraints{Pattern: "^E[0-9]+$"})
				two := 2
				So(built["/given_name"].Constraints, ShouldResemble, &FillFormDataFieldConstraints{MinLength: &two})
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
