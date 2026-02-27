package webapp

import (
	"image"
	"mime"
	"net/http"
	"net/url"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/secretcode"
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

//nolint:gocognit
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

	// Set q_login_id_key to match q_login_id_input_type
	if inKey := form.Get("q_login_id_key"); inKey == "" {
	Switch:
		switch form.Get("q_login_id_input_type") {
		case "phone":
			for _, k := range p.LoginID.Keys {
				if k.Type == model.LoginIDKeyTypePhone {
					form.Set("q_login_id_key", k.Key)
					break Switch
				}
			}
		case "email":
			for _, k := range p.LoginID.Keys {
				if k.Type == model.LoginIDKeyTypeEmail {
					form.Set("q_login_id_key", k.Key)
					break Switch
				}
			}
		case "text":
			fallthrough
		default:
			for _, k := range p.LoginID.Keys {
				if k.Type != model.LoginIDKeyTypePhone {
					form.Set("q_login_id_key", k.Key)
					break Switch
				}
			}
		}
	}
}

func CreateQRCodeImage(content string, width int, height int, level qr.ErrorCorrectionLevel) (image.Image, error) {
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

func FindLoginIDInPreviousInput(s *webapp.Session, xStep string) (string, bool) {
	if s.Authflow == nil {
		return "", false
	}

	for {
		screen := s.Authflow.AllScreens[xStep]
		if screen == nil {
			return "", false
		}

		if screen.PreviousInput != nil {
			previousInput := screen.PreviousInput
			if loginID, ok := previousInput["login_id"].(string); ok {
				return loginID, true
			}
		}

		if screen.BranchStateToken != nil {
			branchXStep := screen.BranchStateToken.XStep
			branchScreen := s.Authflow.AllScreens[branchXStep]
			if branchScreen != nil {
				previousInput := branchScreen.PreviousInput
				if loginID, ok := previousInput["login_id"].(string); ok {
					return loginID, true
				}
			}
		}

		// Otherwise update xStep and find recursively.
		xStep = screen.PreviousXStep
	}
}

func FormatRecoveryCodes(recoveryCodes []string) []string {
	out := make([]string, len(recoveryCodes))
	for i, code := range recoveryCodes {
		out[i] = secretcode.RecoveryCode.FormatForHuman(code)
	}
	return out
}

func SetRecoveryCodeAttachmentHeaders(w http.ResponseWriter) {
	// No need to use FormatMediaType because the value is constant.
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{
		"filename": "recovery_codes.txt",
	}))
}
