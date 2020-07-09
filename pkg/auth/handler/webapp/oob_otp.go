package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator/oob"
	interactionflows "github.com/authgear/authgear-server/pkg/auth/dependency/interaction/flows"
	"github.com/authgear/authgear-server/pkg/db"
	"github.com/authgear/authgear-server/pkg/httproute"
	"github.com/authgear/authgear-server/pkg/template"
	"github.com/authgear/authgear-server/pkg/validation"
)

const (
	TemplateItemTypeAuthUIOOBOTPHTML config.TemplateItemType = "auth_ui_oob_otp_html"
)

var TemplateAuthUIOOBOTPHTML = template.Spec{
	Type:        TemplateItemTypeAuthUIOOBOTPHTML,
	IsHTML:      true,
	Translation: TemplateItemTypeAuthUITranslationJSON,
	Defines:     defines,
	Components:  components,
	Default: `<!DOCTYPE html>
<html>
{{ template "auth_ui_html_head.html" . }}
<body class="page">
<div class="content">

{{ template "auth_ui_header.html" . }}

<div class="simple-form vertical-form form-fields-container">

<div class="nav-bar">
	<button class="btn back-btn" type="button" title="{{ localize "back-button-title" }}"></button>
</div>

<!-- FIXME(webapp): x_login_id_input_type, x_calling_code, x_national_number, x_login_id -->
{{ if .x_login_id_input_type }}
{{ if eq .x_login_id_input_type "phone" }}
<div class="title primary-txt">{{ localize "oob-otp-page-title--sms" }}</div>
{{ end }}
{{ if not (eq .x_login_id_input_type "phone") }}
<div class="title primary-txt">{{ localize "oob-otp-page-title--email" }}</div>
{{ end }}
{{ end }}

{{ template "ERROR" . }}

{{ if .x_login_id_input_type }}
{{ if eq .x_login_id_input_type "phone" }}
<!-- FIXME: x_calling_code x_national_number -->
<div class="description primary-txt">{{ localize "oob-otp-description--sms" $.OOBOTPCodeLength "FIXME" "FIXME" }}</div>
{{ end }}
{{ if not (eq .x_login_id_input_type "phone") }}
<!-- FIXME: x_login_id -->
<div class="description primary-txt">{{ localize "oob-otp-description--email" $.OOBOTPCodeLength "FIXME" }}</div>
{{ end }}
{{ end }}

<form class="vertical-form form-fields-container" method="post" novalidate>
{{ $.CSRFField }}

<input class="input text-input primary-txt" type="text" inputmode="numeric" pattern="[0-9]*" name="x_password" placeholder="{{ localize "oob-otp-placeholder" }}">
<button class="btn primary-btn align-self-flex-end" type="submit" name="submit" value="">{{ localize "next-button-label" }}</button>
</form>

<form class="link oob-otp-trigger-form" method="post" novalidate>
{{ $.CSRFField }}

<span class="primary-txt">{{ localize "oob-otp-resend-button-hint" }}</span>
<button id="resend-button" class="anchor" type="submit" name="trigger" value="true"
	data-cooldown="{{ $.OOBOTPCodeSendCooldown }}"
	data-label="{{ localize "oob-otp-resend-button-label" }}"
	data-label-unit="{{ localize "oob-otp-resend-button-label--unit" }}">{{ localize "oob-otp-resend-button-label" }}</button>
</form>

</div>
{{ template "auth_ui_footer.html" . }}

</div>
</body>
</html>
`,
}

const OOBOTPRequestSchema = "OOBOTPRequestSchema"

var OOBOTPSchema = validation.NewMultipartSchema("").
	Add(OOBOTPRequestSchema, `
		{
			"type": "object",
			"properties": {
				"x_password": { "type": "string" }
			},
			"required": ["x_password"]
		}
	`).Instantiate()

func ConfigureOOBOTPRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/oob_otp")
}

type OOBOTPViewModel struct {
	OOBOTPCodeSendCooldown int
	OOBOTPCodeLength       int
}

func NewOOBOTPViewModel() OOBOTPViewModel {
	return OOBOTPViewModel{
		OOBOTPCodeSendCooldown: oob.OOBCodeSendCooldownSeconds,
		OOBOTPCodeLength:       oob.OOBCodeLength,
	}
}

type OOBOTPInteractions interface {
	TriggerOOBOTP(state *interactionflows.State) (*interactionflows.WebAppResult, error)
	EnterSecret(state *interactionflows.State, password string) (*interactionflows.WebAppResult, error)
}

type OOBOTPHandler struct {
	Database      *db.Handle
	State         StateService
	BaseViewModel *BaseViewModeler
	Renderer      Renderer
	Interactions  OOBOTPInteractions
	Responder     Responder
}

func (h *OOBOTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Method == "GET" {
		state, err := h.State.RestoreState(r, true)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		baseViewModel := h.BaseViewModel.ViewModel(r, state.Error)
		oobOTPViewModel := NewOOBOTPViewModel()

		data := map[string]interface{}{}
		Embed(data, baseViewModel)
		Embed(data, oobOTPViewModel)

		h.Renderer.Render(w, r, TemplateItemTypeAuthUIOOBOTPHTML, data)
		return
	}

	trigger := r.Form.Get("trigger") == "true"

	if r.Method == "POST" && trigger {
		h.Database.WithTx(func() error {
			state, err := h.State.RestoreState(r, false)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return err
			}

			var result *interactionflows.WebAppResult
			defer func() {
				h.State.UpdateState(state, result, err)
				h.Responder.Respond(w, r, state, result, err)
			}()

			result, err = h.Interactions.TriggerOOBOTP(state)
			if err != nil {
				return err
			}

			return nil
		})
		return
	}

	if r.Method == "POST" {
		h.Database.WithTx(func() error {
			state, err := h.State.RestoreState(r, false)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return err
			}

			var result *interactionflows.WebAppResult
			defer func() {
				h.State.UpdateState(state, result, err)
				h.Responder.Respond(w, r, state, result, err)
			}()

			err = OOBOTPSchema.PartValidator(OOBOTPRequestSchema).ValidateValue(FormToJSON(r.Form))
			if err != nil {
				return err
			}

			code := r.Form.Get("x_password")

			result, err = h.Interactions.EnterSecret(state, code)
			if err != nil {
				return err
			}

			return nil
		})
	}
}
