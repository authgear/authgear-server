package webapp

import (
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

func TestValidateProvider(t *testing.T) {
	Convey("ValidateProvider", t, func() {
		Convey("PrepareValues", func() {
			c := &config.AuthConfiguration{}
			impl := ValidateProviderImpl{AuthConfiguration: c}
			var form url.Values

			Convey("remove empty value", func() {
				form = url.Values{
					"a": []string{""},
					"b": []string{"non-empty"},
				}
				impl.PrepareValues(form)
				So(form, ShouldHaveLength, 1)
			})

			Convey("prefill text if first login id type is not phone", func() {
				form = url.Values{}
				c.LoginIDKeys = []config.LoginIDKeyConfiguration{
					{Key: "email", Type: "email"},
				}
				impl.PrepareValues(form)
				So(form.Get("x_login_id_input_type"), ShouldEqual, "text")
			})

			Convey("prefill phone if first login id type is phone", func() {
				form = url.Values{}
				c.LoginIDKeys = []config.LoginIDKeyConfiguration{
					{Key: "phone", Type: "phone"},
				}
				impl.PrepareValues(form)
				So(form.Get("x_login_id_input_type"), ShouldEqual, "phone")
			})

			Convey("do not prefill if already specified", func() {
				form = url.Values{
					"x_login_id_input_type": []string{"text"},
				}
				c.LoginIDKeys = []config.LoginIDKeyConfiguration{
					{Key: "phone", Type: "phone"},
				}
				impl.PrepareValues(form)
				So(form.Get("x_login_id_input_type"), ShouldEqual, "text")
			})
		})

		Convey("Validate", func() {
			validator := validation.NewValidator("http://example.com")
			validator.AddSchemaFragments(`
			{
				"$id": "#A",
				"type": "object",
				"properties": {
					"a": { "type": "string", "const": "42" }
				}
			}
			`)

			var err error
			impl := ValidateProviderImpl{Validator: validator}

			err = impl.Validate("#A", url.Values{
				"a": []string{"24"},
			})
			So(err, ShouldNotBeNil)

			err = impl.Validate("#A", url.Values{
				"a": []string{"42"},
			})
			So(err, ShouldBeNil)
		})

		Convey("WebAppLoginRequest", func() {
			var err error
			impl := ValidateProviderImpl{Validator: validator}

			err = impl.Validate("#WebAppLoginRequest", url.Values{})
			So(err, ShouldNotBeNil)

			err = impl.Validate("#WebAppLoginidRequest", url.Values{
				"x_login_id_input_type": []string{"phone"},
			})
			So(err, ShouldBeNil)
		})

		Convey("WebAppLoginLoginIDRequest", func() {
			var err error
			impl := ValidateProviderImpl{Validator: validator}

			err = impl.Validate("#WebAppLoginLoginIDRequest", url.Values{
				"x_step":                []string{"login:submit_login_id"},
				"x_login_id_input_type": []string{"phone"},
			})
			So(err, ShouldNotBeNil)

			err = impl.Validate("#WebAppLoginLoginIDRequest", url.Values{
				"x_step":                []string{"login:submit_login_id"},
				"x_login_id_input_type": []string{"phone"},
				"x_calling_code":        []string{"852"},
				"x_national_number":     []string{"99887766"},
			})
			So(err, ShouldBeNil)

			err = impl.Validate("#WebAppLoginLoginIDRequest", url.Values{
				"x_step":                []string{"login:submit_login_id"},
				"x_login_id_input_type": []string{"text"},
			})
			So(err, ShouldNotBeNil)

			err = impl.Validate("#WebAppLoginLoginIDRequest", url.Values{
				"x_step":                []string{"login:submit_login_id"},
				"x_login_id_input_type": []string{"text"},
				"x_login_id":            []string{"john.doe"},
			})
			So(err, ShouldBeNil)
		})

		Convey("WebAppLoginLoginIDPasswordRequest", func() {
			var err error
			impl := ValidateProviderImpl{Validator: validator}

			err = impl.Validate("#WebAppLoginLoginIDPasswordRequest", url.Values{
				"x_step":                []string{"login:submit_password"},
				"x_login_id_input_type": []string{"phone"},
			})
			So(err, ShouldNotBeNil)

			err = impl.Validate("#WebAppLoginLoginIDPasswordRequest", url.Values{
				"x_step":                []string{"login:submit_password"},
				"x_login_id_input_type": []string{"phone"},
				"x_calling_code":        []string{"852"},
				"x_national_number":     []string{"99887766"},
			})
			So(err, ShouldNotBeNil)

			err = impl.Validate("#WebAppLoginLoginIDPasswordRequest", url.Values{
				"x_step":                []string{"login:submit_password"},
				"x_login_id_input_type": []string{"text"},
			})
			So(err, ShouldNotBeNil)

			err = impl.Validate("#WebAppLoginLoginIDPasswordRequest", url.Values{
				"x_step":                []string{"login:submit_password"},
				"x_login_id_input_type": []string{"text"},
				"x_login_id":            []string{"john.doe"},
			})
			So(err, ShouldNotBeNil)

			err = impl.Validate("#WebAppLoginLoginIDPasswordRequest", url.Values{
				"x_step":                []string{"login:submit_password"},
				"x_login_id_input_type": []string{"text"},
				"x_login_id":            []string{"john.doe"},
				"x_password":            []string{"123456"},
			})
			So(err, ShouldBeNil)

			err = impl.Validate("#WebAppLoginLoginIDPasswordRequest", url.Values{
				"x_step":                []string{"login:submit_password"},
				"x_login_id_input_type": []string{"phone"},
				"x_calling_code":        []string{"852"},
				"x_national_number":     []string{"99887766"},
				"x_password":            []string{"123456"},
			})
			So(err, ShouldBeNil)
		})
	})
}
