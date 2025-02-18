package event

import (
	"context"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestParseHookResponse(t *testing.T) {
	ctx := context.Background()
	Convey("ParseHookResponse", t, func() {
		pass := func(raw string, expected *HookResponse) {
			r := strings.NewReader(raw)
			actual, err := ParseHookResponse(ctx, r)
			So(err, ShouldBeNil)
			So(actual, ShouldResemble, expected)
		}

		fail := func(raw string) {
			r := strings.NewReader(raw)
			_, err := ParseHookResponse(ctx, r)
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
					"standard_attributes": {
						"given_name": "johndoe"
					},
					"custom_attributes": {
						"foobar": "42"
					}
				}
			}
		}`, &HookResponse{
			IsAllowed: true,
			Mutations: Mutations{
				User: UserMutations{
					StandardAttributes: map[string]interface{}{
						"given_name": "johndoe",
					},
					CustomAttributes: map[string]interface{}{
						"foobar": "42",
					},
				},
			},
		})

		pass(`{
			"is_allowed": true,
			"mutations": {
				"jwt": {
					"payload": {
						"https://example.com": {
							"foo": "bar"
						}
					}
				}
			}
		}`, &HookResponse{
			IsAllowed: true,
			Mutations: Mutations{
				JWT: JWTMutations{
					Payload: map[string]interface{}{
						"https://example.com": map[string]interface{}{
							"foo": "bar",
						},
					},
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
			"is_allowed": true,
			"mutations": {
				"user": {
					"standard_attributes": 1,
					"custom_attributes": 2
				}
			}
		}`)

		fail(`{
			"is_allowed": false,
			"mutations": {}
		}`)
	})
}
