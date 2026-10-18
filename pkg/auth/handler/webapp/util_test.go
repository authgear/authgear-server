package webapp

import (
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestJSONPointerFormToMap(t *testing.T) {
	Convey("TestJSONPointerFormToMap", t, func() {
		test := func(input map[string]string, expected map[string]string) {
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
		}, map[string]string{
			"/name":       "John Doe",
			"/given_name": "",
		})

		// A checkbox is followed by a hidden input of "false".
		// Checked, both are submitted, and the checkbox wins because it comes first.
		Convey("takes the first value", func() {
			form := url.Values{"/accepted": []string{"true", "false"}}
			So(JSONPointerFormToMap(form), ShouldResemble, map[string]string{"/accepted": "true"})
			form = url.Values{"/accepted": []string{"false"}}
			So(JSONPointerFormToMap(form), ShouldResemble, map[string]string{"/accepted": "false"})
		})
	})
}
