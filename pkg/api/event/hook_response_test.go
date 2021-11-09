package event

import (
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestParseHookResponse(t *testing.T) {
	Convey("ParseHookResponse", t, func() {
		pass := func(raw string, expected *HookResponse) {
			r := strings.NewReader(raw)
			actual, err := ParseHookResponse(r)
			So(err, ShouldBeNil)
			So(actual, ShouldResemble, expected)
		}

		fail := func(raw string) {
			r := strings.NewReader(raw)
			_, err := ParseHookResponse(r)
			So(err, ShouldNotBeNil)
		}

		pass(`{
			"is_allowed": true
		}`, &HookResponse{
			IsAllowed: true,
		})

		pass(`{
			"is_allowed": true,
			"mutations": {}
		}`, &HookResponse{
			IsAllowed: true,
		})

		pass(`{
			"is_allowed": true,
			"mutations": {
				"user": {}
			}
		}`, &HookResponse{
			IsAllowed: true,
		})

		pass(`{
			"is_allowed": true,
			"mutations": {
				"user": {
					"standard_attributes": {}
				}
			}
		}`, &HookResponse{
			IsAllowed: true,
			Mutations: Mutations{
				User: UserMutations{
					StandardAttributes: map[string]interface{}{},
				},
			},
		})

		pass(`{
			"is_allowed": false
		}`, &HookResponse{
			IsAllowed: false,
		})

		pass(`{
			"is_allowed": false,
			"title": "Title"
		}`, &HookResponse{
			IsAllowed: false,
			Title:     "Title",
		})

		pass(`{
			"is_allowed": false,
			"title": "Title",
			"reason": "Reason"
		}`, &HookResponse{
			IsAllowed: false,
			Title:     "Title",
			Reason:    "Reason",
		})

		fail(`{
			"is_allowed": false,
			"mutations": {}
		}`)
	})
}
