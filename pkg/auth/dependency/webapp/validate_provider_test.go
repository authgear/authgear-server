package webapp

import (
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/core/validation"
)

func TestValidateProvider(t *testing.T) {
	Convey("ValidateProvider", t, func() {
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
}
