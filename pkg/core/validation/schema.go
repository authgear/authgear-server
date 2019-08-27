package validation

import (
	"github.com/louischan-oursky/gojsonschema"
)

const (
	schemaString = `
{
	"$id": "http://v2.skygear.io",
	"definitions": {
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
					"type": "object",
					"additionalProperties": {
						"$ref": "#APIClientConfiguration"
					}
				},
				"master_key": { "$ref": "#NonEmptyString" },
				"url_prefix": {
					"type": "string",
					"format": "URLFullOnly"
				},
				"cors": { "$ref": "#CORSConfiguration" },
				"auth": { "$ref": "#AuthConfiguration" },
				"user_audit": { "$ref": "#UserAuditConfiguration" },
				"forgot_password": { "$ref": "#ForgotPasswordConfiguration" },
				"welcome_email": { "$ref": "#WelcomeEmailConfiguration" },
				"sso": { "$ref": "#SSOConfiguration" },
				"user_verification": { "$ref": "#UserVerificationConfiguration" },
				"hook": { "$ref": "#HookUserConfiguration" }
			},
			"required": ["master_key", "auth", "hook"]
		},
		"APIClientConfiguration": {
			"$id": "#APIClientConfiguration",
			"type": "object",
			"properties": {
				"name": { "$ref": "#NonEmptyString" },
				"disabled": { "type": "boolean" },
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
			"required": ["name", "api_key", "session_transport"]
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
				"login_id_keys": {
					"type": "object",
					"minProperties": 1,
					"additionalProperties": {
						"$ref": "#LoginIDKeyConfiguration"
					}
				},
				"allowed_realms": {
					"type": "array",
					"minItems": 1,
					"items": { "$ref": "#NonEmptyString" }
				}
			}
		},
		"LoginIDKeyConfiguration": {
			"$id": "#LoginIDKeyConfiguration",
			"type": "object",
			"properties": {
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
				"trail_handler_url": { "type": "string" },
				"password": { "$ref": "#PasswordConfiguration" }
			}
		},
		"PasswordConfiguration": {
			"$id": "#PasswordConfiguration",
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
					"type": "object",
					"additionalProperties": { "$ref": "#UserVerificationKeyConfiguration" }
				}
			}
		},
		"UserVerificationKeyConfiguration": {
			"$id": "#UserVerificationKeyConfiguration",
			"type": "object",
			"properties": {
				"code_format": {
					"type": "string",
					"enum": ["numeric", "complex"]
				},
				"expiry": { "$ref": "#NonNegativeInteger" },
				"success_redirect": { "type": "string" },
				"success_html_url": { "type": "string" },
				"error_redirect": { "type": "string" },
				"error_html_url": { "type": "string" },
				"provider": {
					"type": "string",
					"enum": ["smtp", "twilio", "nexmo"]
				},
				"provider_config": { "$ref": "#UserVerificationProviderConfiguration" }
			},
			"required": ["provider"]
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
		}
	}
}
`
)

var (
	schemaLoader     *gojsonschema.SchemaLoader
	userConfigSchema *gojsonschema.Schema
)

func init() {
	schemaLoader = gojsonschema.NewSchemaLoader()
	err := schemaLoader.AddSchemas(gojsonschema.NewStringLoader(schemaString))
	if err != nil {
		panic(err)
	}
	userConfigSchema, err = schemaLoader.Compile(gojsonschema.NewReferenceLoader("http://v2.skygear.io#UserConfiguration"))
	if err != nil {
		panic(err)
	}
}

func ValidateUserConfiguration(value interface{}) error {
	result, err := userConfigSchema.Validate(gojsonschema.NewGoLoader(value))
	if err != nil {
		return err
	}
	if !result.Valid() {
		return ConvertErrors(result.Errors())
	}
	return nil
}
