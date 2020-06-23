package webapp

import (
	"net/url"

	"github.com/skygeario/skygear-server/pkg/auth/config"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

var validator *validation.Validator

func init() {
	validator = validation.NewValidator("https://accounts.skygear.io")
	validator.AddSchemaFragments(
		EnterLoginIDRequestSchema,
		CreateLoginIDRequestSchema,
		EnterPasswordRequestSchema,
		ForgotPasswordRequestSchema,
		ResetPasswordRequestSchema,
		SSOCallbackRequestSchema,
		AddOrChangeLoginIDRequestSchema,
		RemoveLoginIDRequestSchema,
	)
}

const EnterLoginIDRequestSchema = `
{
	"$id": "#WebAppEnterLoginIDRequest",
	"type": "object",
	"properties": {
		"x_login_id_input_type": { "type": "string", "enum": ["email", "phone", "text"] },
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
				"x_login_id_input_type": { "type": "string", "enum": ["text", "email"] }
			},
			"required": ["x_login_id"]
		}
	]
}
`

// nolint: gosec
const EnterPasswordRequestSchema = `
{
	"$id": "#WebAppEnterPasswordRequest",
	"type": "object",
	"properties": {
		"x_password": { "type": "string" },
		"x_interaction_token": { "type": "string" }
	},
	"required": ["x_password", "x_interaction_token"]
}
`

const CreateLoginIDRequestSchema = `
{
	"$id": "#WebAppCreateLoginIDRequest",
	"type": "object",
	"properties": {
		"x_login_id_key": { "type": "string" },
		"x_login_id_input_type": { "type": "string", "enum": ["email", "phone", "text"] },
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
				"x_login_id_input_type": { "type": "string", "enum": ["text", "email"] }
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
		"error": { "type": "string" },
		"state": { "type": "string" },
		"code": { "type": "string" },
		"scope": { "type": "string" }
	},
	"required": ["state"]
}
`

// nolint: gosec
const ForgotPasswordRequestSchema = `
{
	"$id": "#WebAppForgotPasswordRequest",
	"type": "object",
	"properties": {
		"x_login_id_input_type": { "type": "string", "enum": ["email", "phone", "text"] },
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
				"x_login_id_input_type": { "type": "string", "enum": ["text", "email"] }
			},
			"required": ["x_login_id"]
		}
	]
}
`

// nolint: gosec
const ResetPasswordRequestSchema = `
{
	"$id": "#WebAppResetPasswordRequest",
	"type": "object",
	"properties": {
		"code": { "type": "string" },
		"x_password": { "type": "string" }
	},
	"required": ["code", "x_password"]
}
`

const AddOrChangeLoginIDRequestSchema = `
{
	"$id": "#WebAppAddOrChangeLoginIDRequest",
	"type": "object",
	"properties": {
		"x_login_id_key": { "type": "string" },
		"x_login_id_input_type": { "type": "string", "enum": ["email", "phone", "text"] }
	},
	"required": ["x_login_id_key", "x_login_id_input_type"]
}
`

const RemoveLoginIDRequestSchema = `
{
	"$id": "#WebAppRemoveLoginIDRequest",
	"type": "object",
	"properties": {
		"x_login_id_key": { "type": "string" },
		"x_old_login_id_value": { "type": "string" }
	},
	"required": ["x_login_id_key", "x_old_login_id_value"]
}
`

type ValidateProviderImpl struct {
	Validator *validation.Validator
	LoginID   *config.LoginIDConfig
	UI        *config.UIConfig
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
		if len(p.LoginID.Keys) > 0 {
			if string(p.LoginID.Keys[0].Type) == "phone" {
				form.Set("x_login_id_input_type", "phone")
			} else if string(p.LoginID.Keys[0].Type) == "email" {
				form.Set("x_login_id_input_type", "email")
			} else {
				form.Set("x_login_id_input_type", "text")
			}
		}
	}

	// Set x_login_id_key to the key of the first login ID.
	if _, ok := form["x_login_id_key"]; !ok {
		if len(p.LoginID.Keys) > 0 {
			form.Set("x_login_id_key", p.LoginID.Keys[0].Key)
		}
	}

	if _, ok := form["x_calling_code"]; !ok {
		form.Set("x_calling_code", p.UI.CountryCallingCode.Default)
	}
}

func (p *ValidateProviderImpl) Validate(schemaID string, form url.Values) (err error) {
	err = p.Validator.ValidateGoValue(schemaID, FormToJSON(form))
	return
}
