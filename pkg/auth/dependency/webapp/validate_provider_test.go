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
		Convey("Prevalidate", func() {
			c := &config.AuthConfiguration{}
			impl := ValidateProviderImpl{AuthConfiguration: c}
			var form url.Values

			Convey("prefill text if first login id type is not phone", func() {
				form = url.Values{}
				c.LoginIDKeys = []config.LoginIDKeyConfiguration{
					{Key: "email", Type: "email"},
				}
				impl.Prevalidate(form)
				So(form.Get("x_login_id_input_type"), ShouldEqual, "text")
			})

			Convey("prefill phone if first login id type is phone", func() {
				form = url.Values{}
				c.LoginIDKeys = []config.LoginIDKeyConfiguration{
					{Key: "phone", Type: "phone"},
				}
				impl.Prevalidate(form)
				So(form.Get("x_login_id_input_type"), ShouldEqual, "phone")
			})

			Convey("do not prefill if already specified", func() {
				form = url.Values{
					"x_login_id_input_type": []string{"text"},
				}
				c.LoginIDKeys = []config.LoginIDKeyConfiguration{
					{Key: "phone", Type: "phone"},
				}
				impl.Prevalidate(form)
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
	})
}
