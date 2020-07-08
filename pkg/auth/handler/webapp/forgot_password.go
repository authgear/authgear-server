package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/db"
	"github.com/authgear/authgear-server/pkg/httproute"
	"github.com/authgear/authgear-server/pkg/template"
	"github.com/authgear/authgear-server/pkg/validation"
)

const (
	// nolint: gosec
	TemplateItemTypeAuthUIForgotPasswordHTML config.TemplateItemType = "auth_ui_forgot_password.html"
)

var TemplateAuthUIForgotPasswordHTML = template.Spec{
	Type:        TemplateItemTypeAuthUIForgotPasswordHTML,
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

<form class="simple-form vertical-form form-fields-container" method="post" novalidate>
{{ $.csrfField }}

<div class="nav-bar">
	<button class="btn back-btn" type="button" title="{{ localize "back-button-title" }}"></button>
</div>

<div class="title primary-txt">{{ localize "forgot-password-page-title" }}</div>

{{ template "ERROR" . }}

{{ if .x_login_id_input_type }}{{ if eq .x_login_id_input_type "phone" }}{{ if .x_login_page_login_id_has_phone }}
<div class="description primary-txt">{{ localize "forgot-password-phone-description" }}</div>
<div class="phone-input">
	<select class="input select primary-txt" name="x_calling_code">
		{{ range .x_calling_codes }}
		<option
			value="{{ . }}"
			{{ if $.x_calling_code }}{{ if eq $.x_calling_code . }}
			selected
			{{ end }}{{ end }}
			>
			+{{ . }}
		</option>
		{{ end }}
	</select>
	<input class="input text-input primary-txt" type="text" inputmode="numeric" pattern="[0-9]*" name="x_national_number" placeholder="{{ localize "phone-number-placeholder" }}">
</div>
{{ end }}{{ end }}{{ end }}

{{ if .x_login_id_input_type }}{{ if (not (eq .x_login_id_input_type "phone")) }}{{ if or (eq .x_login_page_text_login_id_variant "email") (eq .x_login_page_text_login_id_variant "email_or_username") }}
<div class="description primary-txt">{{ localize "forgot-password-email-description" }}</div>
<input class="input text-input primary-txt" type="{{ .x_login_id_input_type }}" name="x_login_id" placeholder="{{ localize "email-placeholder" }}">
{{ end }}{{ end }}{{ end }}

{{ if .x_login_id_input_type }}{{ if eq .x_login_id_input_type "phone" }}{{ if or (eq .x_login_page_text_login_id_variant "email") (eq .x_login_page_text_login_id_variant "email_or_username") }}
<a class="link align-self-flex-start" href="{{ call .MakeURLWithQuery "x_login_id_input_type" "email" }}">{{ localize "use-email-login-id-description" }}</a>
{{ end }}{{ end }}{{ end }}

{{ if .x_login_id_input_type }}{{ if eq .x_login_id_input_type "email" }}{{ if .x_login_page_login_id_has_phone }}
<a class="link align-self-flex-start" href="{{ call .MakeURLWithQuery "x_login_id_input_type" "phone" }}">{{ localize "use-phone-login-id-description" }}</a>
{{ end }}{{ end }}{{ end }}

{{ if or .x_login_page_login_id_has_phone (not (eq .x_login_page_text_login_id_variant "none")) }}
<button class="btn primary-btn submit-btn align-self-flex-end" type="submit" name="submit" value="">{{ localize "next-button-label" }}</button>
{{ end }}

</form>
{{ template "auth_ui_footer.html" . }}

</div>
</body>
</html>
`,
}

const ForgotPasswordRequest = "ForgotPasswordRequest"

var ForgotPasswordSchema = validation.NewMultipartSchema("").
	Add(ForgotPasswordRequest, `
		{
			"type": "object",
			"properties": {
				"x_login_id_input_type": { "type": "string", "enum": ["email", "phone", "text"] },
				"x_calling_code": { "type": "string" },
				"x_national_number": { "type": "string" },
				"x_login_id": { "type": "string" }
			},
			"required": ["x_login_id_input_type"],
			"allOf": [
				{
					"if": {
						"properties": {
							"x_login_id_input_type": { "type": "string", "const": "phone" }
						}
					},
					"then": {
						"required": ["x_calling_code", "x_national_number"]
					}
				},
				{
					"if": {
						"properties": {
							"x_login_id_input_type": { "type": "string", "enum": ["text", "email"] }
						}
					},
					"then": {
						"required": ["x_login_id"]
					}
				}
			]
		}
	`).Instantiate()

func ConfigureForgotPasswordRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/forgot_password")
}

type ForgotPasswordHandler struct {
	Database *db.Handle
}

func (h *ForgotPasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Database.WithTx(func() error {
		// FIXME(webapp): forgot_password
		// if r.Method == "GET" {
		// 	writeResponse, err := h.Provider.GetForgotPasswordForm(w, r)
		// 	writeResponse(err)
		// 	return err
		// }

		// if r.Method == "POST" {
		// 	writeResponse, err := h.Provider.PostForgotPasswordForm(w, r)
		// 	writeResponse(err)
		// 	return err
		// }

		return nil
	})
}
