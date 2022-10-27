package webapp

import (
	"image"
	"net/url"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/model"
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

func JSONPointerFormToMap(form url.Values) map[string]string {
	out := make(map[string]string)
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
	hasEmail := false
	hasUsername := false

	for _, k := range p.LoginID.Keys {
		switch k.Type {
		case model.LoginIDKeyTypeEmail:
			hasEmail = true
		case model.LoginIDKeyTypeUsername:
			hasUsername = true
		}
	}

	nonPhoneLoginIDInputType := "text"
	if hasEmail && !hasUsername {
		nonPhoneLoginIDInputType = "email"
	}

	// Set q_login_id_input_type to the type of the first login ID.
	if _, ok := form["q_login_id_input_type"]; !ok {
		// When SIWE is enabled, keys will be empty.
		// but q_login_id_input_type should always be there.
		if len(p.LoginID.Keys) > 0 {
			if string(p.LoginID.Keys[0].Type) == "phone" {
				form.Set("q_login_id_input_type", "phone")
			} else {
				form.Set("q_login_id_input_type", nonPhoneLoginIDInputType)
			}
		} else {
			form.Set("q_login_id_input_type", "text")
		}
	}

	// Set x_login_id_key to match q_login_id_input_type
	if inKey := form.Get("x_login_id_key"); inKey == "" {
	Switch:
		switch form.Get("q_login_id_input_type") {
		case "phone":
			for _, k := range p.LoginID.Keys {
				if k.Type == model.LoginIDKeyTypePhone {
					form.Set("x_login_id_key", k.Key)
					break Switch
				}
			}
		case "email":
			for _, k := range p.LoginID.Keys {
				if k.Type == model.LoginIDKeyTypeEmail {
					form.Set("x_login_id_key", k.Key)
					break Switch
				}
			}
		case "text":
			fallthrough
		default:
			for _, k := range p.LoginID.Keys {
				if k.Type != model.LoginIDKeyTypePhone {
					form.Set("x_login_id_key", k.Key)
					break Switch
				}
			}
		}
	}
}

func createQRCodeImage(content string, width int, height int, level qr.ErrorCorrectionLevel) (image.Image, error) {
	b, err := qr.Encode(content, level, qr.Auto)

	if err != nil {
		return nil, err
	}

	b, err = barcode.Scale(b, width, height)

	if err != nil {
		return nil, err
	}

	return b, nil
}
