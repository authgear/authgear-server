package authflowv2

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/authenticationflow/declarative"
)

type fakeTranslations map[string]string

func (t fakeTranslations) HasKey(key string) (bool, error) {
	_, ok := t[key]
	return ok, nil
}

func (t fakeTranslations) RenderText(key string, args any) (string, error) {
	return t[key], nil
}

func TestFormFieldLabel(t *testing.T) {
	Convey("formFieldLabel", t, func() {
		Convey("form field label wins over the attribute label", func() {
			tr := fakeTranslations{
				"standard-attribute-label-/given_name":            "First name",
				"v2.page.fill-form.names.field-label-/given_name": "Legal first name",
			}
			label, err := formFieldLabel(tr, "names", "/given_name")
			So(err, ShouldBeNil)
			So(label, ShouldEqual, "Legal first name")
		})

		Convey("form field label of another form is not used", func() {
			tr := fakeTranslations{"v2.page.fill-form.other.field-label-/given_name": "Other"}
			for _, name := range []string{"names", ""} {
				label, err := formFieldLabel(tr, name, "/given_name")
				So(err, ShouldBeNil)
				So(label, ShouldEqual, "Given Name")
			}
		})

		Convey("attribute labels", func() {
			tr := fakeTranslations{
				"standard-attribute-label-/given_name":  "First name",
				"custom-attribute-label-/x_employee_id": "Employee ID",
				// A standard key for a custom attribute is not used.
				"standard-attribute-label-/x_terms": "Wrong",
			}
			label, err := formFieldLabel(tr, "", "/given_name")
			So(err, ShouldBeNil)
			So(label, ShouldEqual, "First name")
			label, err = formFieldLabel(tr, "", "/x_employee_id")
			So(err, ShouldBeNil)
			So(label, ShouldEqual, "Employee ID")
			label, err = formFieldLabel(tr, "", "/x_terms")
			So(err, ShouldBeNil)
			So(label, ShouldEqual, "X Terms")
		})

		Convey("falls back to the attribute name", func() {
			label, err := formFieldLabel(fakeTranslations{}, "names", "/x_employee_id")
			So(err, ShouldBeNil)
			So(label, ShouldEqual, "X Employee Id")
		})
	})
}

func TestFormFieldOptions(t *testing.T) {
	Convey("formFieldOptions", t, func() {
		Convey("locale options use the language names", func() {
			tr := fakeTranslations{"language-zh-HK": "繁體中文"}
			options, err := formFieldOptions(tr, "/locale", []declarative.FillFormDataFieldOption{{Value: "en"}, {Value: "zh-HK"}})
			So(err, ShouldBeNil)
			So(options, ShouldResemble, []FormFieldOptionViewModel{
				{Value: "en", Label: "en"},
				{Value: "zh-HK", Label: "繁體中文"},
			})
		})

		Convey("enum options use the enum labels", func() {
			tr := fakeTranslations{"custom-attribute-enum-label-/x_plan-free": "Free plan"}
			options, err := formFieldOptions(tr, "/x_plan", []declarative.FillFormDataFieldOption{{Value: "free"}, {Value: "pro_plan"}})
			So(err, ShouldBeNil)
			So(options, ShouldResemble, []FormFieldOptionViewModel{
				{Value: "free", Label: "Free plan"},
				{Value: "pro_plan", Label: "Pro Plan"},
			})
		})

		Convey("no options", func() {
			options, err := formFieldOptions(fakeTranslations{}, "/given_name", nil)
			So(err, ShouldBeNil)
			So(options, ShouldBeNil)
		})
	})
}
