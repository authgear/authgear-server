package config

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
				`#: asset is required
#: auth is required
#: hook is required
#: master_key is required
`,
			},
			// Empty auth
			{
				`
{
	"master_key": "master_key",
	"auth": {},
	"hook": {}
}
				`,
				`#: asset is required
#/auth: authentication_session is required
#/hook: secret is required
`,
			},
			// Empty auth.login_id_keys and auth.allowed_realms
			{
				`
{
	"master_key": "master_key",
	"asset": {},
	"auth": {
		"login_id_keys": {},
		"allowed_realms": []
	},
	"hook": {}
}
				`,
				`#/asset: secret is required
#/auth: authentication_session is required
#/auth/allowed_realms: Array must have at least 1 items
#/auth/login_id_keys: Must have at least 1 properties
#/hook: secret is required
`,
			},
			// Invalid login id type
			{
				`
{
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
	"hook": {}
}
				`,
				`#: asset is required
#/auth: authentication_session is required
#/auth/login_id_keys/invalid/type: auth.login_id_keys.invalid.type must be one of the following: "raw", "email", "phone"
#/hook: secret is required
`,
			},
			// Minimal valid example
			{
				`
{
	"master_key": "master_key",
	"asset": {
		"secret": "assetsecret"
	},
	"auth": {
		"authentication_session": {
			"secret": "authnsessionsecret"
		},
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
	"hook": {
		"secret": "hooksecret"
	}
}
				`,
				``,
			},
			// API Clients
			{
				`
{
	"clients": {
		"web-app": {}
	},
	"asset": {
		"secret": "assetsecret"
	},
	"master_key": "master_key",
	"auth": {
		"authentication_session": {
			"secret": "authnsessionsecret"
		},
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
	"hook": {
		"secret": "hooksecret"
	}
}
				`,
				`#/clients/web-app: api_key is required
#/clients/web-app: name is required
#/clients/web-app: session_transport is required
`,
			},
			// CORS
			{
				`
{
	"master_key": "master_key",
	"asset": {
		"secret": "assetsecret"
	},
	"auth": {
		"authentication_session": {
			"secret": "authnsessionsecret"
		},
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
	"hook": {
		"secret": "hooksecret"
	},
	"cors": {}
}
				`,
				`#/cors: origin is required
`,
			},
			// MFA
			{
				`
{
	"master_key": "master_key",
	"asset": {
		"secret": "assetsecret"
	},
	"auth": {
		"authentication_session": {
			"secret": "authnsessionsecret"
		},
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
	"hook": {
		"secret": "hooksecret"
	},
	"mfa": {
		"enforcement": "",
		"maximum": 16,
		"totp": {
			"maximum": 6
		},
		"oob": {
			"sms": {
				"maximum": 6
			},
			"email": {
				"maximum": 6
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
}
				`,
				`#/mfa/bearer_token/expire_in_days: Must be greater than or equal to 1
#/mfa/enforcement: mfa.enforcement must be one of the following: "off", "optional", "required"
#/mfa/maximum: Must be less than or equal to 15
#/mfa/oob/email/maximum: Must be less than or equal to 5
#/mfa/oob/sms/maximum: Must be less than or equal to 5
#/mfa/recovery_code/count: Must be less than or equal to 24
#/mfa/recovery_code/list_enabled: Invalid type. Expected: boolean, given: integer
#/mfa/totp/maximum: Must be less than or equal to 5
`,
			},
			// User Audit
			{
				`
{
	"master_key": "master_key",
	"asset": {
		"secret": "assetsecret"
	},
	"auth": {
		"authentication_session": {
			"secret": "authnsessionsecret"
		},
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
}
				`,
				`#/password_policy/expiry_days: Must be greater than or equal to 0
#/password_policy/history_days: Must be greater than or equal to 0
#/password_policy/history_size: Must be greater than or equal to 0
#/password_policy/min_length: Must be greater than or equal to 0
#/password_policy/minimum_guessable_level: Must be less than or equal to 4
`,
			},
			// WelcomeEmailConfiguration
			{
				`
{
	"master_key": "master_key",
	"asset": {
		"secret": "assetsecret"
	},
	"auth": {
		"authentication_session": {
			"secret": "authnsessionsecret"
		},
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
	"hook": {
		"secret": "hooksecret"
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
	"master_key": "master_key",
	"asset": {
		"secret": "assetsecret"
	},
	"auth": {
		"authentication_session": {
			"secret": "authnsessionsecret"
		},
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
	"hook": {
		"secret": "hooksecret"
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
	"master_key": "master_key",
	"asset": {
		"secret": "assetsecret"
	},
	"auth": {
		"authentication_session": {
			"secret": "authnsessionsecret"
		},
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
	"hook": {
		"secret": "hooksecret"
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
				`#/sso: custom_token is required
#/sso/oauth: Must validate "then" as "if" was valid
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
	"master_key": "master_key",
	"asset": {
		"secret": "assetsecret"
	},
	"auth": {
		"authentication_session": {
			"secret": "authnsessionsecret"
		},
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
	"hook": {
		"secret": "hooksecret"
	},
	"user_verification": {
		"criteria": "invalid",
		"login_id_keys": {
			"email": {
				"code_format": "invalid"
			}
		}
	}
}
				`,
				`#/user_verification/criteria: user_verification.criteria must be one of the following: "any", "all"
#/user_verification/login_id_keys/email/code_format: user_verification.login_id_keys.email.code_format must be one of the following: "numeric", "complex"
`,
			},
			// SMTP config
			{
				`
{
	"master_key": "master_key",
	"asset": {
		"secret": "assetsecret"
	},
	"auth": {
		"authentication_session": {
			"secret": "authnsessionsecret"
		},
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
	"hook": {
		"secret": "hooksecret"
	},
	"smtp": {
		"mode": "invalid"
	}
}
				`,
				`#/smtp/mode: smtp.mode must be one of the following: "normal", "ssl"
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
