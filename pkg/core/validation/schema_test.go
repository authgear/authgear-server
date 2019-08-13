package validation

import (
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestValidateUserConfiguration(t *testing.T) {
	Convey("ValidateUserConfiguration", t, func() {
		cases := []struct {
			input  string
			errStr string
		}{
			// Empty root
			{
				`{}`,
				`#: api_key is required
#: auth is required
#: master_key is required
#: token_store is required
`,
			},
			// Empty auth and empty token_store
			{
				`
{
	"api_key": "api_key",
	"master_key": "master_key",
	"auth": {},
	"token_store": {}
}
				`,
				`#/auth: allowed_realms is required
#/auth: login_id_keys is required
#/token_store: secret is required
`,
			},
			// Empty auth.login_id_keys and auth.allowed_realms
			{
				`
{
	"api_key": "api_key",
	"master_key": "master_key",
	"auth": {
		"login_id_keys": {},
		"allowed_realms": []
	},
	"token_store": {}
}
				`,
				`#/auth/allowed_realms: Array must have at least 1 items
#/auth/login_id_keys: Must have at least 1 properties
#/token_store: secret is required
`,
			},
			// Invalid login id type
			{
				`
{
	"api_key": "api_key",
	"master_key": "master_key",
	"auth": {
		"login_id_keys": {
			"email": {
				"type": "email"
			},
			"phone": {
				"type": "phone"
			},
			"username": {
				"type": "raw"
			},
			"invalid": {
				"type": "invalid"
			}
		},
		"allowed_realms": ["default"]
	},
	"token_store": {}
}
				`,
				`#/auth/login_id_keys/type: auth.login_id_keys.type must be one of the following: "raw", "email", "phone"
#/token_store: secret is required
`,
			},
			// Minimal valid example
			{
				`
{
	"api_key": "api_key",
	"master_key": "master_key",
	"auth": {
		"login_id_keys": {
			"email": {
				"type": "email"
			},
			"phone": {
				"type": "phone"
			},
			"username": {
				"type": "raw"
			}
		},
		"allowed_realms": ["default"]
	},
	"token_store": {
		"secret": "tokensecret"
	}
}
				`,
				``,
			},
			// CORS
			{
				`
{
	"api_key": "api_key",
	"master_key": "master_key",
	"auth": {
		"login_id_keys": {
			"email": {
				"type": "email"
			},
			"phone": {
				"type": "phone"
			},
			"username": {
				"type": "raw"
			}
		},
		"allowed_realms": ["default"]
	},
	"token_store": {
		"secret": "tokensecret"
	},
	"cors": {}
}
				`,
				`#/cors: origin is required
`,
			},
			// User Audit
			{
				`
{
	"api_key": "api_key",
	"master_key": "master_key",
	"auth": {
		"login_id_keys": {
			"email": {
				"type": "email"
			},
			"phone": {
				"type": "phone"
			},
			"username": {
				"type": "raw"
			}
		},
		"allowed_realms": ["default"]
	},
	"token_store": {
		"secret": "tokensecret"
	},
	"user_audit": {
		"password": {
			"min_length": -1,
			"minimum_guessable_level": 5,
			"history_size": -1,
			"history_days": -1,
			"expiry_days": -1
		}
	}
}
				`,
				`#/user_audit/password/expiry_days: Must be greater than or equal to 0/1
#/user_audit/password/history_days: Must be greater than or equal to 0/1
#/user_audit/password/history_size: Must be greater than or equal to 0/1
#/user_audit/password/min_length: Must be greater than or equal to 0/1
#/user_audit/password/minimum_guessable_level: Must be less than or equal to 4/1
`,
			},
			// WelcomeEmailConfiguration
			{
				`
{
	"api_key": "api_key",
	"master_key": "master_key",
	"auth": {
		"login_id_keys": {
			"email": {
				"type": "email"
			},
			"phone": {
				"type": "phone"
			},
			"username": {
				"type": "raw"
			}
		},
		"allowed_realms": ["default"]
	},
	"token_store": {
		"secret": "tokensecret"
	},
	"welcome_email": {
		"destination": "invalid"
	}
}
				`,
				`#/welcome_email/destination: welcome_email.destination must be one of the following: "first", "all"
`,
			},
			// CustomTokenConfiguration
			{
				`
{
	"api_key": "api_key",
	"master_key": "master_key",
	"auth": {
		"login_id_keys": {
			"email": {
				"type": "email"
			},
			"phone": {
				"type": "phone"
			},
			"username": {
				"type": "raw"
			}
		},
		"allowed_realms": ["default"]
	},
	"token_store": {
		"secret": "tokensecret"
	},
	"sso": {
		"custom_token": {
			"enabled": true
		}
	}
}
				`,
				`#/sso/custom_token: Must validate "then" as "if" was valid
#/sso/custom_token: audience is required
#/sso/custom_token: issuer is required
#/sso/custom_token: secret is required
`,
			},
			// OAuth
			{
				`
{
	"api_key": "api_key",
	"master_key": "master_key",
	"auth": {
		"login_id_keys": {
			"email": {
				"type": "email"
			},
			"phone": {
				"type": "phone"
			},
			"username": {
				"type": "raw"
			}
		},
		"allowed_realms": ["default"]
	},
	"token_store": {
		"secret": "tokensecret"
	},
	"sso": {
		"oauth": {
			"providers": [
				{
					"type": "azureadv2"
				},
				{
					"type": "google"
				}
			]
		}
	}
}
				`,
				`#/sso/oauth: Must validate "then" as "if" was valid
#/sso/oauth: allowed_callback_urls is required
#/sso/oauth: state_jwt_secret is required
#/sso/oauth/providers/0: Must validate "then" as "if" was valid
#/sso/oauth/providers/0: client_id is required
#/sso/oauth/providers/0: client_secret is required
#/sso/oauth/providers/0: tenant is required
#/sso/oauth/providers/1: Must validate "else" as "if" was not valid
#/sso/oauth/providers/1: client_id is required
#/sso/oauth/providers/1: client_secret is required
`,
			},
			// UserVerificationConfiguration
			{
				`
{
	"api_key": "api_key",
	"master_key": "master_key",
	"auth": {
		"login_id_keys": {
			"email": {
				"type": "email"
			},
			"phone": {
				"type": "phone"
			},
			"username": {
				"type": "raw"
			}
		},
		"allowed_realms": ["default"]
	},
	"token_store": {
		"secret": "tokensecret"
	},
	"user_verification": {
		"criteria": "invalid",
		"login_id_keys": {
			"email": {
				"code_format": "invalid",
				"provider": "invalid"
			}
		}
	}
}
				`,
				`#/user_verification/criteria: user_verification.criteria must be one of the following: "any", "all"
#/user_verification/login_id_keys/code_format: user_verification.login_id_keys.code_format must be one of the following: "numeric", "complex"
#/user_verification/login_id_keys/provider: user_verification.login_id_keys.provider must be one of the following: "smtp", "twilio", "nexmo"
`,
			},
		}
		for _, c := range cases {
			var value interface{}
			err := json.Unmarshal([]byte(c.input), &value)
			So(err, ShouldBeNil)
			err = ValidateUserConfiguration(value)
			if c.errStr == "" {
				So(err, ShouldBeNil)
			} else {
				So(err, ShouldNotBeNil)
				if err != nil {
					So(err.Error(), ShouldEqual, c.errStr)
				}
			}
		}
	})
}
