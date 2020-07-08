package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator/oob"
	"github.com/authgear/authgear-server/pkg/db"
	"github.com/authgear/authgear-server/pkg/httproute"
	"github.com/authgear/authgear-server/pkg/template"
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

<!-- FIXME: x_login_id_input_type, x_calling_code, x_national_number, x_login_id -->

{{ if eq .x_login_id_input_type "phone" }}
<div class="title primary-txt">{{ localize "oob-otp-page-title--sms" }}</div>
{{ end }}
{{ if not (eq .x_login_id_input_type "phone") }}
<div class="title primary-txt">{{ localize "oob-otp-page-title--email" }}</div>
{{ end }}

{{ template "ERROR" . }}

{{ if eq .x_login_id_input_type "phone" }}
<!-- FIXME: x_calling_code x_national_number -->
<div class="description primary-txt">{{ localize "oob-otp-description--sms" .x_oob_otp_code_length "FIXME" "FIXME" }}</div>
{{ end }}
{{ if not (eq .x_login_id_input_type "phone") }}
<!-- FIXME: x_login_id -->
<div class="description primary-txt">{{ localize "oob-otp-description--email" .x_oob_otp_code_length "FIXME" }}</div>
{{ end }}

<form class="vertical-form form-fields-container" method="post" novalidate>
{{ $.csrfField }}

<input class="input text-input primary-txt" type="text" inputmode="numeric" pattern="[0-9]*" name="x_password" placeholder="{{ localize "oob-otp-placeholder" }}">
<button class="btn primary-btn align-self-flex-end" type="submit" name="submit" value="">{{ localize "next-button-label" }}</button>
</form>

<form class="link oob-otp-trigger-form" method="post" novalidate>
{{ $.csrfField }}

<span class="primary-txt">{{ localize "oob-otp-resend-button-hint" }}</span>
<button id="resend-button" class="anchor" type="submit" name="trigger" value="true"
	data-cooldown="{{ .x_oob_otp_code_send_cooldown }}"
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

type OOBOTPHandler struct {
	Database *db.Handle
}

func (h *OOBOTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Database.WithTx(func() error {
		// FIXME(webapp): oob_otp
		// if r.Method == "GET" {
		// 	writeResponse, err := h.Provider.GetOOBOTPForm(w, r)
		// 	writeResponse(err)
		// 	return err
		// }

		// if r.Method == "POST" {
		// 	if r.Form.Get("trigger") == "true" {
		// 		r.Form.Del("trigger")
		// 		writeResponse, err := h.Provider.TriggerOOBOTP(w, r)
		// 		writeResponse(err)
		// 		return err
		// 	}

		// 	writeResponse, err := h.Provider.EnterSecret(w, r)
		// 	writeResponse(err)
		// 	return err
		// }

		return nil
	})
}
