package blocking

import (
	"context"
	"strings"
	"testing"

	"github.com/authgear/authgear-server/pkg/api/event"
	. "github.com/smartystreets/goconvey/convey"
)

func TestBlockingHookResponse(t *testing.T) {
	Convey("BlockingHookResponse", t, func() {
		ctx := context.Background()
		pass := func(name string, typ event.Type, raw string) {
			Convey(name, func() {
				r := strings.NewReader(raw)
				_, err := event.ParseHookResponse(ctx, typ, r)
				So(err, ShouldBeNil)
			})
		}

		fail := func(name string, typ event.Type, raw string) {
			Convey(name, func() {
				r := strings.NewReader(raw)
				_, err := event.ParseHookResponse(ctx, typ, r)
				So(err, ShouldNotBeNil)
			})
		}

		Convey(string(AuthenticationPreInitialize), func() {
			pass("is_allowed true", AuthenticationPreInitialize, `{
				"is_allowed": true
			}`)

			pass("constraints supported", AuthenticationPreInitialize, `{
				"is_allowed": true,
				"constraints": {
					"amr": ["mfa"]
				}
			}`)

			pass("bot_protection supported", AuthenticationPreInitialize, `{
				"is_allowed": true,
				"bot_protection": {
					"mode": "always"
				}
			}`)

			pass("rate_limits supported", AuthenticationPreInitialize, `{
				"is_allowed": true,
				"rate_limits": {
					"authentication.general": {
						"weight": 1.5
					}
				}
			}`)

			fail("mutations not supported", AuthenticationPreInitialize, `{
				"is_allowed": true,
				"mutations": {}
			}`)
		})

		Convey(string(AuthenticationPostIdentified), func() {
			pass("is_allowed true", AuthenticationPostIdentified, `{
				"is_allowed": true
			}`)

			pass("constraints supported", AuthenticationPostIdentified, `{
				"is_allowed": true,
				"constraints": {
					"amr": ["mfa"]
				}
			}`)

			pass("bot_protection supported", AuthenticationPostIdentified, `{
				"is_allowed": true,
				"bot_protection": {
					"mode": "always"
				}
			}`)

			pass("rate_limits supported", AuthenticationPostIdentified, `{
				"is_allowed": true,
				"rate_limits": {
					"authentication.general": {
						"weight": 1.5
					}
				}
			}`)

			fail("mutations not supported", AuthenticationPostIdentified, `{
				"is_allowed": true,
				"mutations": {}
			}`)
		})

		Convey(string(AuthenticationPreAuthenticated), func() {
			pass("is_allowed true", AuthenticationPreAuthenticated, `{
				"is_allowed": true
			}`)

			pass("constraints supported", AuthenticationPreAuthenticated, `{
				"is_allowed": true,
				"constraints": {
					"amr": ["mfa"]
				}
			}`)

			pass("rate_limits supported", AuthenticationPreAuthenticated, `{
				"is_allowed": true,
				"rate_limits": {
					"authentication.general": {
						"weight": 1.5
					}
				}
			}`)

			fail("bot_protection not supported", AuthenticationPreAuthenticated, `{
				"is_allowed": true,
				"bot_protection": {
					"mode": "always"
				}
			}`)

			fail("mutations not supported", AuthenticationPreAuthenticated, `{
				"is_allowed": true,
				"mutations": {}
			}`)
		})

		Convey(string(OIDCJWTPreCreate), func() {
			pass("is_allowed true", OIDCJWTPreCreate, `{
				"is_allowed": true
			}`)

			pass("mutations supported", OIDCJWTPreCreate, `{
				"is_allowed": true,
				"mutations": {
					"jwt": {
						"payload": {
							"key": "value"
						}
					}
				}
			}`)

			fail("constraints not supported", OIDCJWTPreCreate, `{
				"is_allowed": true,
				"constraints": {
					"amr": ["mfa"]
				}
			}`)

			fail("bot_protection not supported", OIDCJWTPreCreate, `{
				"is_allowed": true,
				"bot_protection": {
					"mode": "always"
				}
			}`)

			fail("rate_limits not supported", OIDCJWTPreCreate, `{
				"is_allowed": true,
				"rate_limits": {
					"authentication.general": {
						"weight": 1.5
					}
				}
			}`)
		})

		Convey(string(UserPreCreate), func() {
			pass("is_allowed true", UserPreCreate, `{
				"is_allowed": true
			}`)

			pass("mutations supported", UserPreCreate, `{
				"is_allowed": true,
				"mutations": {
					"user": {
						"is_anonymous": true
					}
				}
			}`)

			fail("constraints not supported", UserPreCreate, `{
				"is_allowed": true,
				"constraints": {
					"amr": ["mfa"]
				}
			}`)

			fail("bot_protection not supported", UserPreCreate, `{
				"is_allowed": true,
				"bot_protection": {
					"mode": "always"
				}
			}`)

			fail("rate_limits not supported", UserPreCreate, `{
				"is_allowed": true,
				"rate_limits": {
					"authentication.general": {
						"weight": 1.5
					}
				}
			}`)
		})

		Convey(string(UserPreScheduleAnonymization), func() {
			pass("is_allowed true", UserPreScheduleAnonymization, `{
				"is_allowed": true
			}`)

			pass("mutations supported", UserPreScheduleAnonymization, `{
				"is_allowed": true,
				"mutations": {
					"user": {
						"is_anonymous": true
					}
				}
			}`)

			fail("constraints not supported", UserPreScheduleAnonymization, `{
				"is_allowed": true,
				"constraints": {
					"amr": ["mfa"]
				}
			}`)

			fail("bot_protection not supported", UserPreScheduleAnonymization, `{
				"is_allowed": true,
				"bot_protection": {
					"mode": "always"
				}
			}`)

			fail("rate_limits not supported", UserPreScheduleAnonymization, `{
				"is_allowed": true,
				"rate_limits": {
					"authentication.general": {
						"weight": 1.5
					}
				}
			}`)
		})

		Convey(string(UserPreScheduleDeletion), func() {
			pass("is_allowed true", UserPreScheduleDeletion, `{
				"is_allowed": true
			}`)

			pass("mutations supported", UserPreScheduleDeletion, `{
				"is_allowed": true,
				"mutations": {
					"user": {
						"is_anonymous": true
					}
				}
			}`)

			fail("constraints not supported", UserPreScheduleDeletion, `{
				"is_allowed": true,
				"constraints": {
					"amr": ["mfa"]
				}
			}`)

			fail("bot_protection not supported", UserPreScheduleDeletion, `{
				"is_allowed": true,
				"bot_protection": {
					"mode": "always"
				}
			}`)

			fail("rate_limits not supported", UserPreScheduleDeletion, `{
				"is_allowed": true,
				"rate_limits": {
					"authentication.general": {
						"weight": 1.5
					}
				}
			}`)
		})

		Convey(string(UserProfilePreUpdate), func() {
			pass("is_allowed true", UserProfilePreUpdate, `{
				"is_allowed": true
			}`)

			pass("mutations supported", UserProfilePreUpdate, `{
				"is_allowed": true,
				"mutations": {
					"user": {
						"is_anonymous": true
					}
				}
			}`)

			fail("constraints not supported", UserProfilePreUpdate, `{
				"is_allowed": true,
				"constraints": {
					"amr": ["mfa"]
				}
			}`)

			fail("bot_protection not supported", UserProfilePreUpdate, `{
				"is_allowed": true,
				"bot_protection": {
					"mode": "always"
				}
			}`)

			fail("rate_limits not supported", UserProfilePreUpdate, `{
				"is_allowed": true,
				"rate_limits": {
					"authentication.general": {
						"weight": 1.5
					}
				}
			}`)
		})
	})
}
