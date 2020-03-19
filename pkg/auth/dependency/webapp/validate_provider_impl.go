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
		AuthenticateRequestSchema,
		AuthenticateLoginIDRequestSchema,
	)
}

const AuthenticateRequestSchema = `
{
	"$id": "#WebAppAuthenticateRequest",
	"type": "object",
	"properties": {
		"x_login_id_input_type": { "type": "string", "enum": ["phone", "text"] }
	},
	"required": ["x_login_id_input_type"]
}
`

const AuthenticateLoginIDRequestSchema = `
{
	"$id": "#WebAppAuthenticateLoginIDRequest",
	"type": "object",
	"properties": {
		"x_login_id_input_type": { "type": "string", "enum": ["phone", "text"] },
		"x_step": { "type": "string", "const": "submit_login_id" },
		"x_calling_code": { "type": "string" },
		"x_national_number": { "type": "string" },
		"x_login_id": { "type": "string" }
	},
	"required": ["x_login_id_input_type", "x_step"],
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

type ValidateProviderImpl struct {
	Validator         *validation.Validator
	AuthConfiguration *config.AuthConfiguration
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
		if len(p.AuthConfiguration.LoginIDKeys) > 0 {
			if string(p.AuthConfiguration.LoginIDKeys[0].Type) == "phone" {
				form.Set("x_login_id_input_type", "phone")
			} else {
				form.Set("x_login_id_input_type", "text")
			}
		}
	}
}

func (p *ValidateProviderImpl) Validate(schemaID string, form url.Values) (err error) {
	err = p.Validator.ValidateGoValue(schemaID, FormToJSON(form))
	return
}
