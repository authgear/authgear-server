package webapp

import (
	"net/url"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

func FormToJSON(form url.Values) map[string]interface{} {
	j := make(map[string]interface{})
	// Do not support recurring parameter
	for name := range form {
		value := form.Get(name)
		if value != "" {
			j[name] = value
		}
	}
	return j
}

func JSONPointerFormToMap(form url.Values) map[string]interface{} {
	out := make(map[string]interface{})
	for ptrStr := range form {
		val := form.Get(ptrStr)
		_, err := jsonpointer.Parse(ptrStr)
		if err != nil {
			// ignore this field because it does not seem a valid json pointer.
			continue
		}

		out[ptrStr] = val
	}
	return out
}

type FormPrefiller struct {
	LoginID *config.LoginIDConfig
	UI      *config.UIConfig
}

func (p *FormPrefiller) Prefill(form url.Values) {
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
}
