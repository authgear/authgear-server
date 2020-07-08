package webapp

import (
	"net/url"
)

func FormToJSON(form url.Values) map[string]interface{} {
	j := make(map[string]interface{})
	// Do not support recurring parameter
	for name := range form {
		j[name] = form.Get(name)
	}
	return j
}

// type ValidateProviderImpl struct {
// 	LoginID *config.LoginIDConfig
// 	UI      *config.UIConfig
// }
//
// func (p *ValidateProviderImpl) PrepareValues(form url.Values) {
// 	// Remove empty values to be compatible with JSON Schema.
// 	for name := range form {
// 		if form.Get(name) == "" {
// 			delete(form, name)
// 		}
// 	}
//
// 	// Set x_login_id_input_type to the type of the first login ID.
// 	if _, ok := form["x_login_id_input_type"]; !ok {
// 		if len(p.LoginID.Keys) > 0 {
// 			if string(p.LoginID.Keys[0].Type) == "phone" {
// 				form.Set("x_login_id_input_type", "phone")
// 			} else if string(p.LoginID.Keys[0].Type) == "email" {
// 				form.Set("x_login_id_input_type", "email")
// 			} else {
// 				form.Set("x_login_id_input_type", "text")
// 			}
// 		}
// 	}
//
// 	// Set x_login_id_key to the key of the first login ID.
// 	if _, ok := form["x_login_id_key"]; !ok {
// 		if len(p.LoginID.Keys) > 0 {
// 			form.Set("x_login_id_key", p.LoginID.Keys[0].Key)
// 		}
// 	}
//
// 	if _, ok := form["x_calling_code"]; !ok {
// 		form.Set("x_calling_code", p.UI.CountryCallingCode.Default)
// 	}
// }
// func (p *ValidateProviderImpl) Validate(partID string, form url.Values) (err error) {
// 	err = WebAppSchema.PartValidator(partID).ValidateValue(FormToJSON(form))
// 	return
// }
