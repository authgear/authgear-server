package config

import (
	"encoding/json"
	"testing"

	"github.com/skygeario/skygear-server/pkg/core/validation"

	. "github.com/smartystreets/goconvey/convey"
)

func TestValidateUserConfiguration(t *testing.T) {
	Convey("ValidateUserConfiguration", t, func() {
		test := func(input string, errors ...string) {
			var value interface{}
			err := json.Unmarshal([]byte(input), &value)
			So(err, ShouldBeNil)
			err = ValidateUserConfiguration(value)
			if len(errors) == 0 {
				So(err, ShouldBeNil)
			} else {
				So(err, ShouldNotBeNil)
				So(validation.ErrorCauseStrings(err), ShouldResemble, errors)
			}
		}
		// Empty root
		test(
			`{}`,
			"/asset: Required",
			"/auth: Required",
			"/hook: Required",
			"/master_key: Required",
		)
		// Empty auth
		test(`
			{
				"master_key": "master_key",
				"auth": {},
				"hook": {}
			}`,
			"/asset: Required",
			"/auth/authentication_session: Required",
			"/hook/secret: Required",
		)
		// Empty auth.login_id_keys
		test(`
			{
				"master_key": "master_key",
				"asset": {},
				"auth": {
					"login_id_keys": []
				},
				"hook": {}
			}`,
			"/asset/secret: Required",
			"/auth/authentication_session: Required",
			"/auth/login_id_keys: EntryAmount map[gte:1]",
			"/hook/secret: Required",
		)
		// Invalid login id type
		test(`
			{
				"master_key": "master_key",
				"auth": {
					"login_id_keys": [
						{
							"key": "email",
							"type": "email"
						},
						{
							"key": "phone",
							"type": "phone"
						},
						{
							"key": "username",
							"type": "username"
						},
						{
							"key": "invalid",
							"type": "invalid"
						}
					],
					"login_id_types": {
						"email": {
							"case_sensitive": false,
							"block_plus_sign": false,
							"ignore_dot_sign": false
						},
						"username": {
							"block_reserved_usernames": true,
							"excluded_keywords": [ "skygear" ],
							"ascii_only": false,
							"case_sensitive": false
						},
						"phone": {}
					}
				},
				"hook": {}
			}`,
			"/asset: Required",
			"/auth/authentication_session: Required",
			"/auth/login_id_keys/3/type: Enum map[expected:[raw email phone username]]",
			"/auth/login_id_types/phone: ExtraEntry",
			"/hook/secret: Required",
		)
		// Minimal valid example
		test(`
			{
				"master_key": "master_key",
				"asset": {
					"secret": "assetsecret"
				},
				"auth": {
					"authentication_session": {
						"secret": "authnsessionsecret"
					},
					"login_id_keys": [
						{
							"key": "email",
							"type": "email"
						},
						{
							"key": "phone",
							"type": "phone"
						},
						{
							"key": "username",
							"type": "username"
						}
					]
				},
				"hook": {
					"secret": "hooksecret"
				}
			}`,
		)
		// API Clients
		test(`
			{
				"clients": [
					{
						"key": "web-app"
					}
				],
				"asset": {
					"secret": "assetsecret"
				},
				"master_key": "master_key",
				"auth": {
					"authentication_session": {
						"secret": "authnsessionsecret"
					},
					"login_id_keys": [
						{
							"key": "email",
							"type": "email"
						},
						{
							"key": "phone",
							"type": "phone"
						},
						{
							"key": "username",
							"type": "username"
						}
					]
				},
				"hook": {
					"secret": "hooksecret"
				}
			}`,
			"/clients/0/api_key: Required",
			"/clients/0/id: Required",
			"/clients/0/name: Required",
			"/clients/0/session_transport: Required",
		)
		// MFA
		test(`
			{
				"master_key": "master_key",
				"asset": {
					"secret": "assetsecret"
				},
				"auth": {
					"authentication_session": {
						"secret": "authnsessionsecret"
					},
					"login_id_keys": [
						{
							"key": "email",
							"type": "email"
						},
						{
							"key": "phone",
							"type": "phone"
						},
						{
							"key": "username",
							"type": "username"
						}
					]
				},
				"hook": {
					"secret": "hooksecret"
				},
				"mfa": {
					"enforcement": "",
					"maximum": 1000,
					"totp": {
						"maximum": 1000
					},
					"oob": {
						"sms": {
							"maximum": 1000
						},
						"email": {
							"maximum": 1000
						}
					},
					"bearer_token": {
						"expire_in_days": 0
					},
					"recovery_code": {
						"count": 100,
						"list_enabled": 1
					}
				}
			}`,
			"/mfa/bearer_token/expire_in_days: NumberRange map[gte:1]",
			"/mfa/enforcement: Enum map[expected:[off optional required]]",
			"/mfa/maximum: NumberRange map[lte:999]",
			"/mfa/oob/email/maximum: NumberRange map[lte:999]",
			"/mfa/oob/sms/maximum: NumberRange map[lte:999]",
			"/mfa/recovery_code/count: NumberRange map[lte:24]",
			"/mfa/recovery_code/list_enabled: Type map[expected:boolean]",
			"/mfa/totp/maximum: NumberRange map[lte:999]",
		)
		// User Audit
		test(`
			{
				"master_key": "master_key",
				"asset": {
					"secret": "assetsecret"
				},
				"auth": {
					"authentication_session": {
						"secret": "authnsessionsecret"
					},
					"login_id_keys": [
						{
							"key": "email",
							"type": "email"
						},
						{
							"key": "phone",
							"type": "phone"
						},
						{
							"key": "username",
							"type": "username"
						}
					]
				},
				"hook": {
					"secret": "hooksecret"
				},
				"password_policy": {
					"min_length": -1,
					"minimum_guessable_level": 5,
					"history_size": -1,
					"history_days": -1,
					"expiry_days": -1
				}
			}`,
			"/password_policy/expiry_days: NumberRange map[gte:0]",
			"/password_policy/history_days: NumberRange map[gte:0]",
			"/password_policy/history_size: NumberRange map[gte:0]",
			"/password_policy/min_length: NumberRange map[gte:0]",
			"/password_policy/minimum_guessable_level: NumberRange map[lte:4]",
		)
		// WelcomeEmailConfiguration
		test(`
			{
				"master_key": "master_key",
				"asset": {
					"secret": "assetsecret"
				},
				"auth": {
					"authentication_session": {
						"secret": "authnsessionsecret"
					},
					"login_id_keys": [
						{
							"key": "email",
							"type": "email"
						},
						{
							"key": "phone",
							"type": "phone"
						},
						{
							"key": "username",
							"type": "username"
						}
					]
				},
				"hook": {
					"secret": "hooksecret"
				},
				"welcome_email": {
					"destination": "invalid"
				}
			}`,
			"/welcome_email/destination: Enum map[expected:[first all]]",
		)
		// CustomTokenConfiguration
		test(`
			{
				"master_key": "master_key",
				"asset": {
					"secret": "assetsecret"
				},
				"auth": {
					"authentication_session": {
						"secret": "authnsessionsecret"
					},
					"login_id_keys": [
						{
							"key": "email",
							"type": "email"
						},
						{
							"key": "phone",
							"type": "phone"
						},
						{
							"key": "username",
							"type": "username"
						}
					]
				},
				"hook": {
					"secret": "hooksecret"
				},
				"sso": {
					"custom_token": {
						"enabled": true
					}
				}
			}`,
			"/sso/custom_token/audience: Required",
			"/sso/custom_token/issuer: Required",
			"/sso/custom_token/secret: Required",
		)
		// OAuth
		test(`
			{
				"master_key": "master_key",
				"asset": {
					"secret": "assetsecret"
				},
				"auth": {
					"authentication_session": {
						"secret": "authnsessionsecret"
					},
					"login_id_keys": [
						{
							"key": "email",
							"type": "email"
						},
						{
							"key": "phone",
							"type": "phone"
						},
						{
							"key": "username",
							"type": "username"
						}
					]
				},
				"hook": {
					"secret": "hooksecret"
				},
				"sso": {
					"oauth": {
						"providers": [
							{ "type": "azureadv2" },
							{ "type": "google" },
							{ "type": "apple" }
						]
					}
				}
			}`,
			"/sso/custom_token: Required",
			"/sso/oauth/allowed_callback_urls: Required",
			"/sso/oauth/providers/0/client_id: Required",
			"/sso/oauth/providers/0/client_secret: Required",
			"/sso/oauth/providers/0/tenant: Required",
			"/sso/oauth/providers/1/client_id: Required",
			"/sso/oauth/providers/1/client_secret: Required",
			"/sso/oauth/providers/2/client_id: Required",
			"/sso/oauth/providers/2/client_secret: Required",
			"/sso/oauth/providers/2/key_id: Required",
			"/sso/oauth/providers/2/team_id: Required",
			"/sso/oauth/state_jwt_secret: Required",
		)
		// UserVerificationConfiguration
		test(`
			{
				"master_key": "master_key",
				"asset": {
					"secret": "assetsecret"
				},
				"auth": {
					"authentication_session": {
						"secret": "authnsessionsecret"
					},
					"login_id_keys": [
						{
							"key": "email",
							"type": "email"
						},
						{
							"key": "phone",
							"type": "phone"
						},
						{
							"key": "username",
							"type": "username"
						}
					]
				},
				"hook": {
					"secret": "hooksecret"
				},
				"user_verification": {
					"criteria": "invalid",
					"login_id_keys": [
						{
							"key": "email",
							"code_format": "invalid"
						}
					]
				}
			}`,
			"/user_verification/criteria: Enum map[expected:[any all]]",
			"/user_verification/login_id_keys/0/code_format: Enum map[expected:[numeric complex]]",
		)
		// SMTP config
		test(`
			{
				"master_key": "master_key",
				"asset": {
					"secret": "assetsecret"
				},
				"auth": {
					"authentication_session": {
						"secret": "authnsessionsecret"
					},
					"login_id_keys": [
						{
							"key": "email",
							"type": "email"
						},
						{
							"key": "phone",
							"type": "phone"
						},
						{
							"key": "username",
							"type": "username"
						}
					]
				},
				"hook": {
					"secret": "hooksecret"
				},
				"smtp": {
					"mode": "invalid"
				}
			}`,
			"/smtp/mode: Enum map[expected:[normal ssl]]",
		)
		// Nexmo config
		test(`
			{
				"master_key": "master_key",
				"asset": {
					"secret": "assetsecret"
				},
				"auth": {
					"authentication_session": {
						"secret": "authnsessionsecret"
					},
					"login_id_keys": [
						{
							"key": "email",
							"type": "email"
						},
						{
							"key": "phone",
							"type": "phone"
						},
						{
							"key": "username",
							"type": "username"
						}
					]
				},
				"hook": {
					"secret": "hooksecret"
				},
				"nexmo": {
					"api_secret": 1
				}
			}`,
			"/nexmo/api_secret: Type map[expected:string]",
		)
	})
}
