package config

var _ = Schema.Add("WorkflowConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"signup_flows": {
			"type": "array",
			"minItems": 1,
			"items": { "$ref": "#/$defs/WorkflowSignupFlow" }
		},
		"login_flows": {
			"type": "array",
			"minItems": 1,
			"items": { "$ref": "#/$defs/WorkflowLoginFlow" }
		},
		"signup_login_flows": {
			"type": "array",
			"minItems": 1,
			"items": { "$ref": "#/$defs/WorkflowSignupLoginFlow" }
		},
		"reauth_flows": {
			"type": "array",
			"minItems": 1,
			"items": { "$ref": "#/$defs/WorkflowReauthFlow" }
		}
	}
}
`)

var _ = Schema.Add("WorkflowObjectID", `
{
	"type": "string",
	"pattern": "^[a-zA-Z_][a-zA-Z0-9_]*$"
}
`)

var _ = Schema.Add("WorkflowIdentificationMethod", `
{
	"type": "string",
	"enum": [
		"email",
		"phone",
		"username",
		"oauth",
		"passkey",
		"siwe"
	]
}
`)

var _ = Schema.Add("WorkflowAuthenticationMethod", `
{
	"type": "string",
	"enum": [
		"primary_password",
		"primary_passkey",
		"primary_oob_otp_email",
		"primary_oob_otp_sms",
		"secondary_password",
		"secondary_totp",
		"secondary_oob_otp_email",
		"secondary_oob_otp_sms"
	]
}
`)

var _ = Schema.Add("WorkflowSignupFlow", `
{
	"type": "object",
	"required": ["id", "steps"],
	"properties": {
		"id": { "$ref": "#/$defs/WorkflowObjectID" },
		"steps": {
			"type": "array",
			"minItems": 1,
			"items": { "$ref": "#/$defs/WorkflowSignupFlowStep" }
		}
	}
}
`)

var _ = Schema.Add("WorkflowSignupFlowStep", `
{
	"type": "object",
	"required": ["type"],
	"properties": {
		"id": { "$ref": "#/$defs/WorkflowObjectID" },
		"type": {
			"type": "string",
			"enum": [
				"identify",
				"authenticate",
				"verify",
				"user_profile"
			]
		}
	},
	"allOf": [
		{
			"if": {
				"required": ["type"],
				"properties": {
					"type": { "const": "identify" }
				}
			},
			"then": {
				"required": ["one_of"],
				"properties": {
					"one_of": {
						"type": "array",
						"items": { "$ref": "#/$defs/WorkflowSignupFlowIdentify" }
					}
				}
			}
		},
		{
			"if": {
				"required": ["type"],
				"properties": {
					"type": { "const": "authenticate" }
				}
			},
			"then": {
				"required": ["one_of"],
				"properties": {
					"one_of": {
						"type": "array",
						"items": { "$ref": "#/$defs/WorkflowSignupFlowAuthenticate" }
					}
				}
			}
		},
		{
			"if": {
				"required": ["type"],
				"properties": {
					"type": { "const": "verify" }
				}
			},
			"then": {
				"required": ["target_step"],
				"properties": {
					"target_step": { "$ref": "#/$defs/WorkflowObjectID" }
				}
			}
		},
		{
			"if": {
				"required": ["type"],
				"properties": {
					"type": { "const": "user_profile" }
				}
			},
			"then": {
				"required": ["user_profile"],
				"properties": {
					"user_profile": {
						"type": "array",
						"items": { "$ref": "#/$defs/WorkflowSignupFlowUserProfile" }
					}
				}
			}
		}
	]
}
`)

var _ = Schema.Add("WorkflowSignupFlowIdentify", `
{
	"type": "object",
	"required": ["identification"],
	"properties": {
		"identification": { "$ref": "#/$defs/WorkflowIdentificationMethod" },
		"steps": {
			"type": "array",
			"items": { "$ref": "#/$defs/WorkflowSignupFlowStep" }
		}
	}
}
`)

var _ = Schema.Add("WorkflowSignupFlowAuthenticate", `
{
	"type": "object",
	"required": ["authentication"],
	"properties": {
		"authentication": { "$ref": "#/$defs/WorkflowAuthenticationMethod" },
		"target_step": { "$ref": "#/$defs/WorkflowObjectID" },
		"steps": {
			"type": "array",
			"items": { "$ref": "#/$defs/WorkflowSignupFlowStep" }
		}
	}
}
`)

var _ = Schema.Add("WorkflowSignupFlowUserProfile", `
{
	"type": "object",
	"required": ["pointer", "required"],
	"properties": {
		"pointer": {
			"type": "string",
			"format": "json-pointer"
		},
		"required": { "type": "boolean" }
	}
}
`)

var _ = Schema.Add("WorkflowLoginFlow", `
{
	"type": "object",
	"required": ["id", "steps"],
	"properties": {
		"id": { "$ref": "#/$defs/WorkflowObjectID" },
		"steps": {
			"type": "array",
			"minItems": 1,
			"items": { "$ref": "#/$defs/WorkflowLoginFlowStep" }
		}
	}
}
`)

var _ = Schema.Add("WorkflowLoginFlowStep", `
{
	"type": "object",
	"required": ["type"],
	"properties": {
		"id": { "$ref": "#/$defs/WorkflowObjectID" },
		"type": {
			"type": "string",
			"enum": [
				"identify",
				"authenticate"
			]
		}
	},
	"allOf": [
		{
			"if": {
				"required": ["type"],
				"properties": {
					"type": { "const": "identify" }
				}
			},
			"then": {
				"required": ["one_of"],
				"properties": {
					"one_of": {
						"type": "array",
						"items": { "$ref": "#/$defs/WorkflowLoginFlowIdentify" }
					}
				}
			}
		},
		{
			"if": {
				"required": ["type"],
				"properties": {
					"type": { "const": "authenticate" }
				}
			},
			"then": {
				"required": ["one_of"],
				"properties": {
					"optional": { "type": "boolean" },
					"one_of": {
						"type": "array",
						"items": { "$ref": "#/$defs/WorkflowSignupFlowAuthenticate" }
					}
				}
			}
		}
	]
}
`)

var _ = Schema.Add("WorkflowLoginFlowIdentify", `
{
	"type": "object",
	"required": ["identification"],
	"properties": {
		"identification": { "$ref": "#/$defs/WorkflowIdentificationMethod" },
		"steps": {
			"type": "array",
			"items": { "$ref": "#/$defs/WorkflowLoginFlowStep" }
		}
	}
}
`)

var _ = Schema.Add("WorkflowSignupFlowAuthenticate", `
{
	"type": "object",
	"required": ["authentication"],
	"properties": {
		"authentication": { "$ref": "#/$defs/WorkflowAuthenticationMethod" },
		"target_step": { "$ref": "#/$defs/WorkflowObjectID" },
		"steps": {
			"type": "array",
			"items": { "$ref": "#/$defs/WorkflowLoginFlowStep" }
		}
	}
}
`)

var _ = Schema.Add("WorkflowSignupLoginFlow", `
{
	"type": "object",
	"required": ["id", "steps"],
	"properties": {
		"id": { "$ref": "#/$defs/WorkflowObjectID" },
		"steps": {
			"type": "array",
			"minItems": 1,
			"items": { "$ref": "#/$defs/WorkflowSignupLoginFlowStep" }
		}
	}
}
`)

var _ = Schema.Add("WorkflowSignupLoginFlowStep", `
{
	"type": "object",
	"required": ["type"],
	"properties": {
		"id": { "$ref": "#/$defs/WorkflowObjectID" },
		"type": {
			"type": "string",
			"enum": ["identify"]
		}
	},
	"allOf": [
		{
			"if": {
				"required": ["type"],
				"properties": {
					"type": { "const": "identify" }
				}
			},
			"then": {
				"required": ["one_of"],
				"properties": {
					"one_of": {
						"type": "array",
						"items": { "$ref": "#/$defs/WorkflowSignupLoginFlowIdentify" }
					}
				}
			}
		}
	]
}
`)

var _ = Schema.Add("WorkflowSignupLoginFlowIdentify", `
{
	"type": "object",
	"required": ["identification", "signup_flow", "login_flow"],
	"properties": {
		"identification": { "$ref": "#/$defs/WorkflowIdentificationMethod" },
		"signup_flow": { "$ref": "#/$defs/WorkflowObjectID" },
		"login_flow": { "$ref": "#/$defs/WorkflowObjectID" }
	}
}
`)

var _ = Schema.Add("WorkflowReauthFlow", `
{
	"type": "object",
	"required": ["id", "steps"],
	"properties": {
		"id": { "$ref": "#/$defs/WorkflowObjectID" },
		"steps": {
			"type": "array",
			"minItems": 1,
			"items": { "$ref": "#/$defs/WorkflowReauthFlowStep" }
		}
	}
}
`)

var _ = Schema.Add("WorkflowReauthFlowStep", `
{
	"type": "object",
	"required": ["type"],
	"properties": {
		"id": { "$ref": "#/$defs/WorkflowObjectID" },
		"type": {
			"type": "string",
			"enum": [
				"authenticate"
			]
		}
	},
	"allOf": [
		{
			"if": {
				"required": ["type"],
				"properties": {
					"type": { "const": "authenticate" }
				}
			},
			"then": {
				"required": ["one_of"],
				"properties": {
					"optional": { "type": "boolean" },
					"one_of": {
						"type": "array",
						"items": { "$ref": "#/$defs/WorkflowReauthFlowAuthenticate" }
					}
				}
			}
		}
	]
}
`)

var _ = Schema.Add("WorkflowReauthFlowAuthenticate", `
{
	"type": "object",
	"required": ["authentication"],
	"properties": {
		"authentication": { "$ref": "#/$defs/WorkflowAuthenticationMethod" },
		"target_step": { "$ref": "#/$defs/WorkflowObjectID" },
		"steps": {
			"type": "array",
			"items": { "$ref": "#/$defs/WorkflowReauthFlowStep" }
		}
	}
}
`)
