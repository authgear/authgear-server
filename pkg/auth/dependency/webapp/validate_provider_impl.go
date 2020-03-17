package webapp

import (
	"net/url"

	"github.com/skygeario/skygear-server/pkg/core/validation"
)

var validator *validation.Validator

func init() {
	validator = validation.NewValidator("https://accounts.skygear.io")
	validator.AddSchemaFragments(AuthenticateRequestSchema)
}

const AuthenticateRequestSchema = `
{
	"$id": "#WebAppAuthenticateRequest",
	"type": "object",
	"properties": {
		"x_login_id_input_type": { "type": "string", "enum": ["phone", "text"] }
	}
}
`

type ValidateProviderImpl struct {
	Validator *validation.Validator
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

func (p *ValidateProviderImpl) Validate(schemaID string, form url.Values) (err error) {
	err = p.Validator.ValidateGoValue(schemaID, FormToJSON(form))
	return
}
