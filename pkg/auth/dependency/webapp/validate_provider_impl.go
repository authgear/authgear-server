package webapp

import (
	"net/url"

	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

var validator *validation.Validator

func init() {
	validator = validation.NewValidator("https://accounts.skygear.io")
	validator.AddSchemaFragments(
		LoginRequestSchema,
		LoginLoginIDRequestSchema,
		LoginLoginIDPasswordRequestSchema,
		SignupRequestSchema,
		SignupLoginIDRequestSchema,
		SignupLoginIDPasswordRequestSchema,
		SSOCallbackRequestSchema,
	)
}

const LoginRequestSchema = `
{
	"$id": "#WebAppLoginRequest",
	"type": "object",
	"properties": {
		"x_login_id_input_type": { "type": "string", "enum": ["phone", "text"] }
	},
	"required": ["x_login_id_input_type"]
}
`

const LoginLoginIDRequestSchema = `
{
	"$id": "#WebAppLoginLoginIDRequest",
	"type": "object",
	"properties": {
		"x_login_id_input_type": { "type": "string", "enum": ["phone", "text"] },
		"x_calling_code": { "type": "string" },
		"x_national_number": { "type": "string" },
		"x_login_id": { "type": "string" }
	},
	"required": ["x_login_id_input_type"],
	"oneOf": [
		{
			"properties": {
				"x_login_id_input_type": { "type": "string", "const": "phone" }
			},
			"required": ["x_calling_code", "x_national_number"]
		},
		{
			"properties": {
				"x_login_id_input_type": { "type": "string", "const": "text" }
			},
			"required": ["x_login_id"]
		}
	]
}
`

// nolint: gosec
const LoginLoginIDPasswordRequestSchema = `
{
	"$id": "#WebAppLoginLoginIDPasswordRequest",
	"type": "object",
	"properties": {
		"x_login_id_input_type": { "type": "string", "enum": ["phone", "text"] },
		"x_calling_code": { "type": "string" },
		"x_national_number": { "type": "string" },
		"x_login_id": { "type": "string" },
		"x_password": { "type": "string" }
	},
	"required": ["x_login_id_input_type", "x_password"],
	"oneOf": [
		{
			"properties": {
				"x_login_id_input_type": { "type": "string", "const": "phone" }
			},
			"required": ["x_calling_code", "x_national_number"]
		},
		{
			"properties": {
				"x_login_id_input_type": { "type": "string", "const": "text" }
			},
			"required": ["x_login_id"]
		}
	]
}
`

const SignupRequestSchema = `
{
	"$id": "#WebAppSignupRequest",
	"type": "object",
	"properties": {
		"x_login_id_key": { "type": "string" }
	},
	"required": ["x_login_id_key"]
}
`

const SignupLoginIDRequestSchema = `
{
	"$id": "#WebAppSignupLoginIDRequest",
	"type": "object",
	"properties": {
		"x_login_id_key": { "type": "string" },
		"x_login_id_input_type": { "type": "string", "enum": ["phone", "text"] },
		"x_calling_code": { "type": "string" },
		"x_national_number": { "type": "string" },
		"x_login_id": { "type": "string" }
	},
	"required": ["x_login_id_key", "x_login_id_input_type"],
	"oneOf": [
		{
			"properties": {
				"x_login_id_input_type": { "type": "string", "const": "phone" }
			},
			"required": ["x_calling_code", "x_national_number"]
		},
		{
			"properties": {
				"x_login_id_input_type": { "type": "string", "const": "text" }
			},
			"required": ["x_login_id"]
		}
	]
}
`

// nolint: gosec
const SignupLoginIDPasswordRequestSchema = `
{
	"$id": "#WebAppSignupLoginIDPasswordRequest",
	"type": "object",
	"properties": {
		"x_login_id_key": { "type": "string" },
		"x_login_id_input_type": { "type": "string", "enum": ["phone", "text"] },
		"x_calling_code": { "type": "string" },
		"x_national_number": { "type": "string" },
		"x_login_id": { "type": "string" },
		"x_password": { "type": "string" }
	},
	"required": ["x_login_id_key", "x_login_id_input_type", "x_password"],
	"oneOf": [
		{
			"properties": {
				"x_login_id_input_type": { "type": "string", "const": "phone" }
			},
			"required": ["x_calling_code", "x_national_number"]
		},
		{
			"properties": {
				"x_login_id_input_type": { "type": "string", "const": "text" }
			},
			"required": ["x_login_id"]
		}
	]
}
`

const SSOCallbackRequestSchema = `
{
	"$id": "#SSOCallbackRequest",
	"type": "object",
	"properties": {
		"state": { "type": "string" },
		"code": { "type": "string" },
		"scope": { "type": "string" }
	},
	"required": ["state", "code"]
}
`

type ValidateProviderImpl struct {
	Validator                       *validation.Validator
	LoginIDConfiguration            *config.LoginIDConfiguration
	CountryCallingCodeConfiguration *config.AuthUICountryCallingCodeConfiguration
}

var _ ValidateProvider = &ValidateProviderImpl{}

func FormToJSON(form url.Values) map[string]interface{} {
	j := make(map[string]interface{})
	// Do not support recurring parameter
	for name := range form {
		j[name] = form.Get(name)
	}
	return j
}

func (p *ValidateProviderImpl) PrepareValues(form url.Values) {
	// Remove empty values to be compatible with JSON Schema.
	for name := range form {
		if form.Get(name) == "" {
			delete(form, name)
		}
	}

	// Set x_login_id_input_type to the type of the first login ID.
	if _, ok := form["x_login_id_input_type"]; !ok {
		if len(p.LoginIDConfiguration.Keys) > 0 {
			if string(p.LoginIDConfiguration.Keys[0].Type) == "phone" {
				form.Set("x_login_id_input_type", "phone")
			} else {
				form.Set("x_login_id_input_type", "text")
			}
		}
	}

	// Set x_login_id_key to the key of the first login ID.
	if _, ok := form["x_login_id_key"]; !ok {
		if len(p.LoginIDConfiguration.Keys) > 0 {
			form.Set("x_login_id_key", p.LoginIDConfiguration.Keys[0].Key)
		}
	}

	if _, ok := form["x_calling_code"]; !ok {
		form.Set("x_calling_code", p.CountryCallingCodeConfiguration.Default)
	}
}

func (p *ValidateProviderImpl) Validate(schemaID string, form url.Values) (err error) {
	err = p.Validator.ValidateGoValue(schemaID, FormToJSON(form))
	return
}
