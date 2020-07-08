package webapp

import (
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/validation"
)

func TestValidateProvider(t *testing.T) {
	Convey("ValidateProvider", t, func() {
		s := WebAppSchema
		defer func() { WebAppSchema = s }()

		Convey("PrepareValues", func() {
			c := &config.LoginIDConfig{}
			impl := ValidateProviderImpl{
				LoginID: c,
				UI: &config.UIConfig{
					CountryCallingCode: &config.UICountryCallingCodeConfig{
						Values:  []string{"852"},
						Default: "852",
					},
				},
			}
			var form url.Values

			Convey("remove empty value", func() {
				form = url.Values{
					"a": []string{""},
					"b": []string{"non-empty"},
				}
				impl.PrepareValues(form)
				_, ok := form["a"]
				So(ok, ShouldBeFalse)
			})

			Convey("prefill email if first login id type is email", func() {
				form = url.Values{}
				c.Keys = []config.LoginIDKeyConfig{
					{Key: "email", Type: "email"},
				}
				impl.PrepareValues(form)
				So(form.Get("x_login_id_input_type"), ShouldEqual, "email")
			})

			Convey("prefill phone if first login id type is phone", func() {
				form = url.Values{}
				c.Keys = []config.LoginIDKeyConfig{
					{Key: "phone", Type: "phone"},
				}
				impl.PrepareValues(form)
				So(form.Get("x_login_id_input_type"), ShouldEqual, "phone")
			})

			Convey("prefill text if first login id type is other", func() {
				form = url.Values{}
				c.Keys = []config.LoginIDKeyConfig{
					{Key: "username", Type: "username"},
				}
				impl.PrepareValues(form)
				So(form.Get("x_login_id_input_type"), ShouldEqual, "text")
			})

			Convey("do not prefill if already specified", func() {
				form = url.Values{
					"x_login_id_input_type": []string{"text"},
				}
				c.Keys = []config.LoginIDKeyConfig{
					{Key: "phone", Type: "phone"},
				}
				impl.PrepareValues(form)
				So(form.Get("x_login_id_input_type"), ShouldEqual, "text")
			})

			Convey("prefill country calling code", func() {
				form = url.Values{}
				impl.PrepareValues(form)
				So(form.Get("x_calling_code"), ShouldEqual, "852")
			})
		})

		Convey("Validate", func() {
			WebAppSchema = validation.NewMultipartSchema("").
				Add("A", `
					{
						"type": "object",
						"properties": {
							"a": { "type": "string", "const": "42" }
						}
					}
				`).
				Instantiate()

			var err error
			impl := ValidateProviderImpl{}

			err = impl.Validate("A", url.Values{
				"a": []string{"24"},
			})
			So(err, ShouldNotBeNil)

			err = impl.Validate("A", url.Values{
				"a": []string{"42"},
			})
			So(err, ShouldBeNil)
		})

		Convey("WebAppEnterLoginIDRequest", func() {
			var err error
			impl := ValidateProviderImpl{}

			err = impl.Validate(WebAppSchemaIDEnterLoginIDRequest, url.Values{
				"x_login_id_input_type": []string{"phone"},
			})
			So(err, ShouldNotBeNil)

			err = impl.Validate(WebAppSchemaIDEnterLoginIDRequest, url.Values{
				"x_login_id_input_type": []string{"phone"},
				"x_calling_code":        []string{"852"},
				"x_national_number":     []string{"99887766"},
			})
			So(err, ShouldBeNil)

			err = impl.Validate(WebAppSchemaIDEnterLoginIDRequest, url.Values{
				"x_login_id_input_type": []string{"text"},
			})
			So(err, ShouldNotBeNil)

			err = impl.Validate(WebAppSchemaIDEnterLoginIDRequest, url.Values{
				"x_login_id_input_type": []string{"text"},
				"x_login_id":            []string{"john.doe"},
			})
			So(err, ShouldBeNil)
		})

		Convey("WebAppEnterPasswordRequest", func() {
			var err error
			impl := ValidateProviderImpl{}

			err = impl.Validate(WebAppSchemaIDEnterPasswordRequest, url.Values{
				"x_password": []string{"123456"},
			})
			So(err, ShouldBeNil)
		})
	})
}
