package config

import (
	"fmt"
	"io"

	"github.com/skygeario/skygear-server/pkg/core/apiversion"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

var (
	// Schemas is public so that the schemas can be composed in other contexts.
	Schemas = fmt.Sprintf(`
{
	"NonEmptyString": {
		"$id": "#NonEmptyString",
		"type": "string",
		"minLength": 1
	},
	"NonNegativeInteger": {
		"$id": "#NonNegativeInteger",
		"type": "integer",
		"minimum": 0
	},
	"TenantConfiguration": {
		"$id": "#TenantConfiguration",
		"type": "object",
		"properties": {
			"api_version": { "enum": %s },
			"app_id": { "$ref": "#NonEmptyString" },
			"app_name": { "$ref": "#NonEmptyString" },
			"database_config": { "$ref": "#DatabaseConfiguration" },
			"hook": { "$ref": "#HookTenantConfiguration" },
			"app_config": { "$ref": "#AppConfiguration" }
		},
		"required": ["api_version", "app_id", "app_name", "database_config", "app_config"]
	},
	"DatabaseConfiguration": {
		"$id": "#DatabaseConfiguration",
		"type": "object",
		"properties": {
			"database_url": { "$ref": "#NonEmptyString" },
			"database_schema": { "$ref": "#NonEmptyString" }
		}
	},
	"HookTenantConfiguration": {
		"$id": "#HookTenantConfiguration",
		"type": "object",
		"properties": {
			"sync_hook_timeout_second": { "type": "integer" },
			"sync_hook_total_timeout_second": { "type": "integer" }
		}
	},
	"AppConfiguration": {
		"$id": "#AppConfiguration",
		"type": "object",
		"additionalProperties": false,
		"properties": {
			"api_version": { "enum": %s },
			"display_app_name": { "type": "string" },
			"clients": {
				"type": "array",
				"items": { "$ref": "#APIClientConfiguration" }
			},
			"master_key": { "$ref": "#NonEmptyString" },
			"session": { "$ref": "#SessionConfiguration" },
			"cors": { "$ref": "#CORSConfiguration" },
			"oidc": { "$ref": "#OIDCConfiguration" },
			"auth": { "$ref": "#AuthConfiguration" },
			"auth_ui": { "$ref": "#AuthUIConfiguration" },
			"mfa": { "$ref": "#MFAConfiguration" },
			"password_policy": { "$ref": "#PasswordPolicyConfiguration" },
			"forgot_password": { "$ref": "#ForgotPasswordConfiguration" },
			"welcome_email": { "$ref": "#WelcomeEmailConfiguration" },
			"sso": { "$ref": "#SSOConfiguration" },
			"user_verification": { "$ref": "#UserVerificationConfiguration" },
			"hook": { "$ref": "#HookAppConfiguration" },
			"smtp" : { "$ref": "#SMTPConfiguration" },
			"twilio" : { "$ref": "#TwilioConfiguration" },
			"nexmo" : { "$ref": "#NexmoConfiguration" },
			"asset": { "$ref": "#AssetConfiguration" }
		},
		"required": ["api_version", "master_key", "auth", "hook", "asset"]
	},
	"AssetConfiguration": {
		"$id": "#AssetConfiguration",
		"type": "object",
		"additionalProperties": false,
		"properties": {
			"secret": { "$ref": "#NonEmptyString" }
		},
		"required": ["secret"]
	},
	"SessionConfiguration": {
		"$id": "#SessionConfiguration",
		"type": "object",
		"additionalProperties": false,
		"properties": {
			"lifetime": { "type": "integer", "minimum": 0 },
			"idle_timeout_enabled": { "type": "boolean" },
			"idle_timeout": { "type": "integer", "minimum": 0 },
			"cookie_domain": { "type": "string" },
			"cookie_non_persistent": { "type": "boolean" }
		}
	},
	"APIClientConfiguration": {
		"$id": "#APIClientConfiguration",
		"type": "object",
		"additionalProperties": false,
		"properties": {
			"client_name": { "$ref": "#NonEmptyString" },
			"client_id": { "$ref": "#NonEmptyString" },
			"redirect_uris": {
				"type": "array",
				"items": { "type": "string" }
			},
			"auth_api_use_cookie": { "type": "boolean" },
			"access_token_lifetime": { "type": "integer", "minimum": 0 },
			"refresh_token_lifetime": { "type": "integer", "minimum": 0 },
			"grant_types": {
				"type": "array",
				"items": { "type": "string" }
			}
		},
		"required": ["client_name", "client_id"]
	},
	"CORSConfiguration": {
		"$id": "#CORSConfiguration",
		"type": "object",
		"additionalProperties": false,
		"properties": {
			"origin": { "type": "string" }
		}
	},
	"OIDCConfiguration": {
		"$id": "#OIDCConfiguration",
		"type": "object",
		"additionalProperties": false,
		"properties": {
			"keys": {
				"type": "array",
				"items": {
					"type": "object",
					"properties": {
						"kid": { "type": "string" },
						"public_key": { "type": "string" },
						"private_key": { "type": "string" }
					},
					"required": ["kid", "public_key", "private_key"]
				}
			}
		}
	},
	"AuthUIConfiguration": {
		"$id": "#AuthUIConfiguration",
		"type": "object",
		"additionalProperties": false,
		"properties": {
			"css": { "type": "string" }
		}
	},
	"AuthConfiguration": {
		"$id": "#AuthConfiguration",
		"type": "object",
		"additionalProperties": false,
		"properties": {
			"enable_api": { "type": "boolean" },
			"authentication_session": { "$ref": "#AuthenticationSessionConfiguration" },
			"login_id_keys": {
				"type": "array",
				"minItems": 1,
				"items": { "$ref": "#LoginIDKeyConfiguration" }
			},
			"login_id_types": { "$ref": "#LoginIDTypesConfiguration" },
			"on_user_duplicate_allow_create": { "type": "boolean" }
		},
		"required": ["authentication_session"]
	},
	"AuthenticationSessionConfiguration": {
		"$id": "#AuthenticationSessionConfiguration",
		"type": "object",
		"additionalProperties": false,
		"properties": {
			"secret": { "$ref": "#NonEmptyString" }
		},
		"required": ["secret"]
	},
	"MFAConfiguration": {
		"$id": "#MFAConfiguration",
		"type": "object",
		"additionalProperties": false,
		"properties": {
			"enabled": { "type": "boolean" },
			"enforcement": {
				"type": "string",
				"enum": ["off", "optional", "required"]
			},
			"maximum": {
				"type": "integer",
				"minimum": 0,
				"maximum": 999
			},
			"totp": {
				"type": "object",
				"additionalProperties": false,
				"properties": {
					"maximum": {
						"type": "integer",
						"minimum": 0,
						"maximum": 999
					}
				}
			},
			"oob": {
				"type": "object",
				"additionalProperties": false,
				"properties": {
					"app_name": { "type": "string" },
					"sender": { "type": "string", "format": "NameEmailAddr" },
					"subject": { "type": "string" },
					"reply_to": { "type": "string", "format": "NameEmailAddr" },
					"sms": {
						"type": "object",
						"additionalProperties": false,
						"properties": {
							"maximum": {
								"type": "integer",
								"minimum": 0,
								"maximum": 999
							}
						}
					},
					"email": {
						"type": "object",
						"additionalProperties": false,
						"properties": {
							"maximum": {
								"type": "integer",
								"minimum": 0,
								"maximum": 999
							}
						}
					}
				}
			},
			"bearer_token": {
				"type": "object",
				"additionalProperties": false,
				"properties": {
					"expire_in_days": {
						"type": "integer",
						"minimum": 1,
						"maximum": 3650
					}
				}
			},
			"recovery_code": {
				"type": "object",
				"additionalProperties": false,
				"properties": {
					"count": {
						"type": "integer",
						"minimum": 8,
						"maximum": 24
					},
					"list_enabled": {
						"type": "boolean"
					}
				}
			}
		}
	},
	"LoginIDKeyConfiguration": {
		"$id": "#LoginIDKeyConfiguration",
		"type": "object",
		"additionalProperties": false,
		"properties": {
			"key": { "$ref": "#NonEmptyString" },
			"type": {
				"type": "string",
				"enum": ["raw", "email", "phone", "username"]
			},
			"maximum": { "$ref": "#NonNegativeInteger" }
		},
		"required": ["type"]
	},
	"LoginIDTypesConfiguration": {
		"$id": "#LoginIDTypesConfiguration",
		"type": "object",
		"additionalProperties": false,
		"properties": {
			"email": { "$ref": "#LoginIDTypeEmailConfiguration" },
			"username": { "$ref": "#LoginIDTypeUsernameConfiguration" }
		}
	},
	"LoginIDTypeEmailConfiguration": {
		"$id": "#LoginIDTypeEmailConfiguration",
		"type": "object",
		"additionalProperties": false,
		"properties": {
			"case_sensitive": { "type": "boolean" },
			"block_plus_sign": { "type": "boolean" },
			"ignore_dot_sign": { "type": "boolean" }
		}
	},
	"LoginIDTypeUsernameConfiguration": {
		"$id": "#LoginIDTypeUsernameConfiguration",
		"type": "object",
		"additionalProperties": false,
		"properties": {
			"block_reserved_usernames": { "type": "boolean" },
			"excluded_keywords": {
				"type": "array",
				"items": { "type": "string" }
			},
			"ascii_only": { "type": "boolean" },
			"case_sensitive": { "type": "boolean" }
		}
	},
	"PasswordPolicyConfiguration": {
		"$id": "#PasswordPolicyConfiguration",
		"type": "object",
		"additionalProperties": false,
		"properties": {
			"min_length": { "$ref": "#NonNegativeInteger" },
			"uppercase_required": { "type": "boolean" },
			"lowercase_required": { "type": "boolean" },
			"digit_required": { "type": "boolean" },
			"symbol_required": { "type": "boolean" },
			"minimum_guessable_level": {
				"type": "integer",
				"minimum": 0,
				"maximum": 4
			},
			"excluded_keywords": {
				"type": "array",
				"items": { "type": "string" }
			},
			"history_size": { "$ref": "#NonNegativeInteger" },
			"history_days": { "$ref": "#NonNegativeInteger" },
			"expiry_days": { "$ref": "#NonNegativeInteger" }
		}
	},
	"ForgotPasswordConfiguration": {
		"$id": "#ForgotPasswordConfiguration",
		"type": "object",
		"additionalProperties": false,
		"properties": {
			"app_name": { "type": "string" },
			"secure_match": { "type": "boolean" },
			"sender": { "type": "string", "format": "NameEmailAddr" },
			"reply_to": { "type": "string", "format": "NameEmailAddr" },
			"subject": { "type": "string" },
			"reset_url_lifetime": { "$ref": "#NonNegativeInteger" },
			"success_redirect": { "type": "string" },
			"error_redirect": { "type": "string" }
		}
	},
	"WelcomeEmailConfiguration": {
		"$id": "#WelcomeEmailConfiguration",
		"type": "object",
		"additionalProperties": false,
		"properties": {
			"enabled": { "type": "boolean" },
			"sender": { "type": "string", "format": "NameEmailAddr" },
			"reply_to": { "type": "string", "format": "NameEmailAddr" },
			"subject": { "type": "string" },
			"destination": {
				"type": "string",
				"enum": ["first", "all"]
			}
		}
	},
	"SSOConfiguration": {
		"$id": "#SSOConfiguration",
		"type": "object",
		"additionalProperties": false,
		"properties": {
			"oauth": { "$ref": "#OAuthConfiguration" }
		}
	},
	"OAuthConfiguration": {
		"$id": "#OAuthConfiguration",
		"type": "object",
		"additionalProperties": false,
		"properties": {
			"state_jwt_secret": { "type": "string" },
			"external_access_token_flow_enabled": { "type": "boolean" },
			"on_user_duplicate_allow_merge": { "type": "boolean" },
			"on_user_duplicate_allow_create": { "type": "boolean" },
			"providers": {
				"type": "array",
				"items": { "$ref": "#OAuthProviderConfiguration" }
			}
		},
		"required": ["state_jwt_secret"]
	},
	"OAuthProviderConfiguration": {
		"$id": "#OAuthProviderConfiguration",
		"type": "object",
		"additionalProperties": false,
		"properties": {
			"id": { "type": "string" },
			"type": {
				"type": "string",
				"enum": ["google", "facebook", "instagram", "linkedin", "azureadv2", "apple"]
			},
			"client_id": { "type": "string" },
			"client_secret": { "type": "string" },
			"scope": { "type": "string" },
			"tenant": { "type": "string" },
			"key_id": { "type": "string" },
			"team_id": { "type": "string" }
		},
		"allOf": [
			{
				"if": {
					"properties": { "type": { "const": "azureadv2" } }
				},
				"then": {
					"required": ["client_id", "client_secret", "tenant"]
				}
			},
			{
				"if": {
					"properties": { "type": { "const": "apple" } }
				},
				"then": {
					"required": ["client_id", "client_secret", "key_id", "team_id"]
				}
			},
			{
				"if": {
					"properties": { "type": { "enum": ["google", "facebook", "instagram", "linkedin"] } }
				},
				"then": {
					"required": ["client_id", "client_secret"]
				}
			}
		]
	},
	"UserVerificationConfiguration": {
		"$id": "#UserVerificationConfiguration",
		"type": "object",
		"additionalProperties": false,
		"properties": {
			"auto_send_on_signup": { "type": "boolean" },
			"criteria": {
				"type": "string",
				"enum": ["any", "all"]
			},
			"error_redirect": { "type": "string" },
			"login_id_keys": {
				"type": "array",
				"items": { "$ref": "#UserVerificationKeyConfiguration" }
			}
		}
	},
	"UserVerificationKeyConfiguration": {
		"$id": "#UserVerificationKeyConfiguration",
		"type": "object",
		"additionalProperties": false,
		"properties": {
			"key": { "$ref": "#NonEmptyString" },
			"code_format": {
				"type": "string",
				"enum": ["numeric", "complex"]
			},
			"expiry": { "$ref": "#NonNegativeInteger" },
			"success_redirect": { "type": "string" },
			"error_redirect": { "type": "string" },
			"subject": { "type": "string" },
			"sender": { "type": "string", "format": "NameEmailAddr" },
			"reply_to": { "type": "string", "format": "NameEmailAddr" }
		}
	},
	"HookAppConfiguration": {
		"$id": "#HookAppConfiguration",
		"type": "object",
		"additionalProperties": false,
		"properties": {
			"secret": { "$ref": "#NonEmptyString" }
		},
		"required": ["secret"]
	},
	"SMTPConfiguration": {
		"$id": "#SMTPConfiguration",
		"type": "object",
		"additionalProperties": false,
		"properties": {
			"host": { "type": "string" },
			"port": { "$ref": "#NonNegativeInteger" },
			"mode": {
				"type": "string",
				"enum": ["normal", "ssl"]
			},
			"login": { "type": "string" },
			"password": { "type": "string" }
		}
	},
	"TwilioConfiguration": {
		"$id": "#TwilioConfiguration",
		"type": "object",
		"additionalProperties": false,
		"properties": {
			"account_sid": { "type": "string" },
			"auth_token": { "type": "string" },
			"from": { "type": "string" }
		}
	},
	"NexmoConfiguration": {
		"$id": "#NexmoConfiguration",
		"type": "object",
		"additionalProperties": false,
		"properties": {
			"api_key": { "type": "string" },
			"api_secret": { "type": "string" },
			"from": { "type": "string" }
		}
	}
}
`, apiversion.SupportedVersionsJSON, apiversion.SupportedVersionsJSON)
)

var (
	tenantConfigurationValidator *validation.Validator
)

func init() {
	tenantConfigurationValidator = validation.NewValidator("http://v2.skygear.io")
	tenantConfigurationValidator.AddSchemaFragments(Schemas)
}

func ParseTenantConfiguration(r io.Reader) (*TenantConfiguration, error) {
	config := TenantConfiguration{}
	err := tenantConfigurationValidator.ParseReader("#TenantConfiguration", r, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func ParseAppConfiguration(r io.Reader) (*AppConfiguration, error) {
	config := AppConfiguration{}
	err := tenantConfigurationValidator.ParseReader("#AppConfiguration", r, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
