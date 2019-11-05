package config

import (
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

const (
	schemaString = `
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
	"UserConfiguration": {
		"$id": "#UserConfiguration",
		"type": "object",
		"properties": {
			"clients": {
				"type": "array",
				"items": { "$ref": "#APIClientConfiguration" }
			},
			"master_key": { "$ref": "#NonEmptyString" },
			"url_prefix": {
				"type": "string",
				"format": "URLFullOnly"
			},
			"cors": { "$ref": "#CORSConfiguration" },
			"auth": { "$ref": "#AuthConfiguration" },
			"mfa": { "$ref": "#MFAConfiguration" },
			"user_audit": { "$ref": "#UserAuditConfiguration" },
			"password_policy": { "$ref": "#PasswordPolicyConfiguration" },
			"forgot_password": { "$ref": "#ForgotPasswordConfiguration" },
			"welcome_email": { "$ref": "#WelcomeEmailConfiguration" },
			"sso": { "$ref": "#SSOConfiguration" },
			"user_verification": { "$ref": "#UserVerificationConfiguration" },
			"hook": { "$ref": "#HookUserConfiguration" },
			"smtp" : { "$ref": "#SMTPConfiguration" },
			"twilio" : { "$ref": "#TwilioConfiguration" },
			"nexmo" : { "$ref": "#NexmoConfiguration" },
			"asset": { "$ref": "#AssetConfiguration" }
		},
		"required": ["master_key", "auth", "hook", "asset"]
	},
	"AssetConfiguration": {
		"$id": "#AssetConfiguration",
		"type": "object",
		"properties": {
			"secret": { "$ref": "#NonEmptyString" }
		},
		"required": ["secret"]
	},
	"APIClientConfiguration": {
		"$id": "#APIClientConfiguration",
		"type": "object",
		"properties": {
			"id": { "$ref": "#NonEmptyString" },
			"name": { "$ref": "#NonEmptyString" },
			"api_key": { "$ref": "#NonEmptyString" },
			"session_transport": {
				"type": "string",
				"enum": ["header", "cookie"]
			},
			"access_token_lifetime": { "type": "integer", "minimum": 0 },
			"session_idle_timeout_enabled": { "type": "boolean" },
			"session_idle_timeout": { "type": "integer", "minimum": 0 },
			"refresh_token_disabled": { "type": "boolean" },
			"refresh_token_lifetime": { "type": "integer", "minimum": 0 }
		},
		"required": ["id", "name", "api_key", "session_transport"]
	},
	"CORSConfiguration": {
		"$id": "#CORSConfiguration",
		"type": "object",
		"properties": {
			"origin": { "$ref": "#NonEmptyString" }
		},
		"required": ["origin"]
	},
	"AuthConfiguration": {
		"$id": "#AuthConfiguration",
		"type": "object",
		"properties": {
			"authentication_session": { "$ref": "#AuthenticationSessionConfiguration" },
			"login_id_keys": {
				"type": "array",
				"minItems": 1,
				"items": { "$ref": "#LoginIDKeyConfiguration" }
			},
			"allowed_realms": {
				"type": "array",
				"minItems": 1,
				"items": { "$ref": "#NonEmptyString" }
			}
		},
		"required": ["authentication_session"]
	},
	"AuthenticationSessionConfiguration": {
		"$id": "#AuthenticationSessionConfiguration",
		"type": "object",
		"properties": {
			"secret": { "$ref": "#NonEmptyString" }
		},
		"required": ["secret"]
	},
	"MFAConfiguration": {
		"$id": "#MFAConfiguration",
		"type": "object",
		"properties": {
			"enforcement": {
				"type": "string",
				"enum": ["off", "optional", "required"]
			},
			"maximum": {
				"type": "integer",
				"minimum": 0,
				"maximum": 15
			},
			"totp": {
				"type": "object",
				"properties": {
					"maximum": {
						"type": "integer",
						"minimum": 0,
						"maximum": 5
					}
				}
			},
			"oob": {
				"type": "object",
				"properties": {
					"sms": {
						"type": "object",
						"properties": {
							"maximum": {
								"type": "integer",
								"minimum": 0,
								"maximum": 5
							}
						}
					},
					"email": {
						"type": "object",
						"properties": {
							"maximum": {
								"type": "integer",
								"minimum": 0,
								"maximum": 5
							}
						}
					}
				}
			},
			"bearer_token": {
				"type": "object",
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
		"properties": {
			"key": { "$ref": "#NonEmptyString" },
			"type": {
				"type": "string",
				"enum": ["raw", "email", "phone"]
			},
			"minimum": { "$ref": "#NonNegativeInteger" },
			"maximum": { "$ref": "#NonNegativeInteger" }
		},
		"required": ["type"]
	},
	"UserAuditConfiguration": {
		"$id": "#UserAuditConfiguration",
		"type": "object",
		"properties": {
			"enabled": { "type": "boolean" },
			"trail_handler_url": { "type": "string" }
		}
	},
	"PasswordPolicyConfiguration": {
		"$id": "#PasswordPolicyConfiguration",
		"type": "object",
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
		"properties": {
			"app_name": { "type": "string" },
			"url_prefix": {
				"type": "string",
				"format": "URLFullOnly"
			},
			"secure_match": { "type": "boolean" },
			"sender": { "type": "string" },
			"reply_to": { "type": "string" },
			"subject": { "type": "string" },
			"reset_url_lifetime": { "$ref": "#NonNegativeInteger" },
			"success_redirect": { "type": "string" },
			"error_redirect": { "type": "string" },
			"email_text_url": { "type": "string" },
			"email_html_url": { "type": "string" },
			"reset_html_url": { "type": "string" },
			"reset_success_html_url": { "type": "string" },
			"reset_error_html_url": { "type": "string" }
		}
	},
	"WelcomeEmailConfiguration": {
		"$id": "#WelcomeEmailConfiguration",
		"type": "object",
		"properties": {
			"enabled": { "type": "boolean" },
			"url_prefix": { "type": "string" },
			"sender": { "type": "string" },
			"reply_to": { "type": "string" },
			"subject": { "type": "string" },
			"text_url": { "type": "string" },
			"html_url": { "type": "string" },
			"destination": {
				"type": "string",
				"enum": ["first", "all"]
			}
		}
	},
	"SSOConfiguration": {
		"$id": "#SSOConfiguration",
		"type": "object",
		"properties": {
			"custom_token": { "$ref": "#CustomTokenConfiguration" },
			"oauth": { "$ref": "#OAuthConfiguration" }
		},
		"required": ["custom_token"]
	},
	"CustomTokenConfiguration": {
		"$id": "#CustomTokenConfiguration",
		"type": "object",
		"properties": {
			"enabled": { "type": "boolean" },
			"issuer": { "$ref": "#NonEmptyString" },
			"audience": { "$ref": "#NonEmptyString" },
			"secret": { "$ref": "#NonEmptyString" },
			"on_user_duplicate_allow_merge": { "type": "boolean" },
			"on_user_duplicate_allow_create": { "type": "boolean" }
		},
		"if": {
			"properties": {
				"enabled": {
					"const": true
				}
			},
			"required": ["enabled"]
		},
		"then": {
			"required": ["issuer", "audience", "secret"]
		},
		"else": {
			"required": ["secret"]
		}
	},
	"OAuthConfiguration": {
		"$id": "#OAuthConfiguration",
		"type": "object",
		"properties": {
			"url_prefix": { "type": "string" },
			"state_jwt_secret": { "type": "string" },
			"allowed_callback_urls": {
				"type": "array",
				"items": { "type": "string" }
			},
			"external_access_token_flow_enabled": { "type": "boolean" },
			"on_user_duplicate_allow_merge": { "type": "boolean" },
			"on_user_duplicate_allow_create": { "type": "boolean" },
			"providers": {
				"type": "array",
				"items": { "$ref": "#OAuthProviderConfiguration" }
			}
		},
		"if": {
			"properties": {
				"providers": {
					"type": "array",
					"minItems": 1
				}
			},
			"required": ["providers"]
		},
		"then": {
			"properties": {
				"allowed_callback_urls": {
					"minItems": 1
				}
			},
			"required": ["state_jwt_secret", "allowed_callback_urls"]
		},
		"else": {
			"required": ["state_jwt_secret"]
		}
	},
	"OAuthProviderConfiguration": {
		"$id": "#OAuthProviderConfiguration",
		"type": "object",
		"properties": {
			"id": { "type": "string" },
			"type": {
				"type": "string",
				"enum": ["google", "facebook", "instagram", "linkedin", "azureadv2"]
			},
			"client_id": { "type": "string" },
			"client_secret": { "type": "string" },
			"scope": { "type": "string" },
			"tenant": { "type": "string" }
		},
		"if": {
			"properties": {
				"type": { "const": "azureadv2" }
			},
			"required": ["type"]
		},
		"then": {
			"required": ["type", "client_id", "client_secret", "tenant"]
		},
		"else": {
			"required": ["type", "client_id", "client_secret"]
		}
	},
	"UserVerificationConfiguration": {
		"$id": "#UserVerificationConfiguration",
		"type": "object",
		"properties": {
			"url_prefix": {
				"type": "string",
				"format": "URLFullOnly"
			},
			"auto_send_on_signup": { "type": "boolean" },
			"criteria": {
				"type": "string",
				"enum": ["any", "all"]
			},
			"error_redirect": { "type": "string" },
			"error_html_url": { "type": "string" },
			"login_id_keys": {
				"type": "array",
				"items": { "$ref": "#UserVerificationKeyConfiguration" }
			}
		}
	},
	"UserVerificationKeyConfiguration": {
		"$id": "#UserVerificationKeyConfiguration",
		"type": "object",
		"properties": {
			"key": { "$ref": "#NonEmptyString" },
			"code_format": {
				"type": "string",
				"enum": ["numeric", "complex"]
			},
			"expiry": { "$ref": "#NonNegativeInteger" },
			"success_redirect": { "type": "string" },
			"success_html_url": { "type": "string" },
			"error_redirect": { "type": "string" },
			"error_html_url": { "type": "string" },
			"provider_config": { "$ref": "#UserVerificationProviderConfiguration" }
		}
	},
	"UserVerificationProviderConfiguration": {
		"$id": "#UserVerificationProviderConfiguration",
		"type": "object",
		"properties": {
			"subject": { "type": "string" },
			"sender": { "type": "string" },
			"reply_to": { "type": "string" },
			"text_url": { "type": "string" },
			"html_url": { "type": "string" }
		}
	},
	"HookUserConfiguration": {
		"$id": "#HookUserConfiguration",
		"type": "object",
		"properties": {
			"secret": { "$ref": "#NonEmptyString" }
		},
		"required": ["secret"]
	},
	"SMTPConfiguration": {
		"$id": "#SMTPConfiguration",
		"type": "object",
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
		"properties": {
			"account_sid": { "type": "string" },
			"auth_token": { "type": "string" },
			"from": { "type": "string" }
		}
	},
	"NexmoConfiguration": {
		"$id": "#NexmoConfiguration",
		"type": "object",
		"properties": {
			"api_key": { "type": "string" },
			"secret": { "type": "string" },
			"from": { "type": "string" }
		}
	}
}
`
)

var (
	userConfigurationValidator *validation.Validator
)

func init() {
	userConfigurationValidator = validation.NewValidator("http://v2.skygear.io")
	userConfigurationValidator.AddSchemaFragments(schemaString)
}

func ValidateUserConfiguration(value interface{}) error {
	return userConfigurationValidator.ValidateGoValue("#UserConfiguration", value)
}
