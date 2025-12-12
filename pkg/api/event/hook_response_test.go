package event

import (
	"context"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestParseHookResponse(t *testing.T) {
	ctx := context.Background()
	s := GetBaseHookResponseSchema()
	s.Add("TestHookResponse", `
{
	"allOf": [
		{ "$ref": "#/$defs/BaseHookResponseSchema" },
		{
			"if": {
				"properties": {
					"is_allowed": { "const": true }
				}
			},
			"then": {
				"type": "object",
				"additionalProperties": false,
				"properties": {
					"is_allowed": {},
					"mutations": {},
					"constraints": {},
					"bot_protection": {},
					"rate_limits": {}
				}
			}
		}
	]
}`)
	s.Instantiate()
	RegisterResponseSchemaValidator("event.test", s.PartValidator("TestHookResponse"))
	Convey("ParseHookResponse", t, func() {
		pass := func(name string, raw string, expected *HookResponse) {
			Convey(name, func() {
				r := strings.NewReader(raw)
				actual, err := ParseHookResponse(ctx, "event.test", r)
				So(err, ShouldBeNil)
				So(actual, ShouldResemble, expected)
			})
		}

		fail := func(name string, raw string) {
			Convey(name, func() {
				r := strings.NewReader(raw)
				_, err := ParseHookResponse(ctx, "event.test", r)
				So(err, ShouldNotBeNil)
			})
		}

		pass("is_allowed true", `{
			"is_allowed": true
		}`, &HookResponse{
			IsAllowed: true,
		})

		pass("is_allowed true, empty mutations", `{
			"is_allowed": true,
			"mutations": {}
		}`, &HookResponse{
			IsAllowed: true,
		})

		pass("is_allowed true, empty user mutations", `{
			"is_allowed": true,
			"mutations": {
				"user": {}
			}
		}`, &HookResponse{
			IsAllowed: true,
		})

		pass("is_allowed true, user mutations with standard and custom attributes", `{
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

		pass("is_allowed true, jwt mutations with payload", `{
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

		pass("is_allowed true, id token mutations with payload", `{
			"is_allowed": true,
			"mutations": {
				"id_token": {
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
				IDToken: IDTokenMutations{
					Payload: map[string]interface{}{
						"https://example.com": map[string]interface{}{
							"foo": "bar",
						},
					},
				},
			},
		})

		pass("is_allowed false", `{
			"is_allowed": false
		}`, &HookResponse{
			IsAllowed: false,
		})

		pass("is_allowed false, with title", `{
			"is_allowed": false,
			"title": "Title"
		}`, &HookResponse{
			IsAllowed: false,
			Title:     "Title",
		})

		pass("is_allowed false, with title and reason", `{
			"is_allowed": false,
			"title": "Title",
			"reason": "Reason"
		}`, &HookResponse{
			IsAllowed: false,
			Title:     "Title",
			Reason:    "Reason",
		})

		pass("is_allowed true, with constraints amr", `{
			"is_allowed": true,
			"constraints": {
				"amr": ["pwd", "otp"]
			}
		}`, &HookResponse{
			IsAllowed: true,
			Constraints: &Constraints{
				AMR: []string{"pwd", "otp"},
			},
		})

		pass("is_allowed true, with user mutations and constraints amr", `{
			"is_allowed": true,
			"mutations": {
				"user": {
					"standard_attributes": {
						"given_name": "johndoe"
					}
				}
			},
			"constraints": {
				"amr": ["pwd"]
			}
		}`, &HookResponse{
			IsAllowed: true,
			Mutations: Mutations{
				User: UserMutations{
					StandardAttributes: map[string]interface{}{
						"given_name": "johndoe",
					},
				},
			},
			Constraints: &Constraints{
				AMR: []string{"pwd"},
			},
		})

		pass("is_allowed true, with bot_protection always", `{
			"is_allowed": true,
			"bot_protection": {
				"mode": "always"
			}
		}`, &HookResponse{
			IsAllowed: true,
			BotProtection: &BotProtectionRequirements{
				Mode: "always",
			},
		})

		pass("is_allowed true, with rate_limits", `{
			"is_allowed": true,
			"rate_limits": {
				"authentication.general": {
					"weight": 1.5
				},
				"authentication.account_enumeration": {
					"weight": 1.5
				}
			}
		}`, &HookResponse{
			IsAllowed: true,
			RateLimits: map[string]RateLimitRequirements{
				"authentication.general": {
					Weight: 1.5,
				},
				"authentication.account_enumeration": {
					Weight: 1.5,
				},
			},
		})

		fail("invalid constraints amr type", `{
			"is_allowed": true,
			"constraints": {
				"amr": "not an array"
			}
		}`)

		fail("invalid user mutations standard_attributes type", `{
			"is_allowed": true,
			"mutations": {
				"user": {
					"standard_attributes": 1,
					"custom_attributes": 2
				}
			}
		}`)

		fail("is_allowed false, with mutations", `{
			"is_allowed": false,
			"mutations": {}
		}`)
	})
}
