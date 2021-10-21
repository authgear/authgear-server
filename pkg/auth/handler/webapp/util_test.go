package webapp

import (
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestJSONPointerFormToMap(t *testing.T) {
	Convey("TestJSONPointerFormToMap", t, func() {
		test := func(input map[string]string, expected map[string]interface{}) {
			form := url.Values{}
			for key, value := range input {
				form[key] = []string{value}
			}
			actual := JSONPointerFormToMap(form)
			So(actual, ShouldResemble, expected)
		}

		test(map[string]string{
			"/name":       "John Doe",
			"/given_name": "",
			"x_action":    "save",
		}, map[string]interface{}{
			"/name":       "John Doe",
			"/given_name": "",
		})
	})
}
