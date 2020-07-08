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
	TemplateItemTypeAuthUILoginHTML config.TemplateItemType = "auth_ui_login.html"
)

var TemplateAuthUILoginHTML = template.Spec{
	Type:        TemplateItemTypeAuthUILoginHTML,
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
		<div class="authorize-form">
			<div class="authorize-idp-section">
				{{ range .x_identity_candidates }}
				{{ if eq .type "oauth" }}
				<form class="authorize-idp-form" method="post" novalidate>
				{{ $.csrfField }}
				<button class="btn sso-btn {{ .provider_type }}" type="submit" name="x_idp_id" value="{{ .provider_alias }}" data-form-xhr="false">
					{{- if eq .provider_type "apple" -}}
					{{ localize "sign-in-apple" }}
					{{- end -}}
					{{- if eq .provider_type "google" -}}
					{{ localize "sign-in-google" }}
					{{- end -}}
					{{- if eq .provider_type "facebook" -}}
					{{ localize "sign-in-facebook" }}
					{{- end -}}
					{{- if eq .provider_type "linkedin" -}}
					{{ localize "sign-in-linkedin" }}
					{{- end -}}
					{{- if eq .provider_type "azureadv2" -}}
					{{ localize "sign-in-azureadv2" }}
					{{- end -}}
				</button>
				</form>
				{{ end }}
				{{ end }}
			</div>

			{{ $has_oauth := false }}
			{{ $has_login_id := false }}
			{{ range .x_identity_candidates }}
				{{ if eq .type "oauth" }}
				{{ $has_oauth = true }}
				{{ end }}
				{{ if eq .type "login_id" }}
				{{ $has_login_id = true }}
				{{ end }}
			{{ end }}
			{{ if $has_oauth }}{{ if $has_login_id }}
			<div class="primary-txt sso-loginid-separator">{{ localize "sso-login-id-separator" }}</div>
			{{ end }}{{ end }}

			{{ template "ERROR" . }}

			<form class="authorize-loginid-form" method="post" novalidate>
				{{ $.csrfField }}

				{{ if .x_login_id_input_type }}{{ if eq .x_login_id_input_type "phone" }}{{ if .x_login_page_login_id_has_phone }}
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

				{{ if .x_login_id_input_type }}{{ if not (eq .x_login_id_input_type "phone") }}{{ if (not (eq .x_login_page_text_login_id_variant "none")) }}
				<input class="input text-input primary-txt" type="{{ .x_login_page_text_login_id_input_type }}" name="x_login_id" placeholder="{{ localize "login-id-placeholder" .x_login_page_text_login_id_variant }}">
				{{ end }}{{ end }}{{ end }}

				{{ if .x_login_id_input_type }}{{ if eq .x_login_id_input_type "phone" }}{{ if (not (eq .x_login_page_text_login_id_variant "none")) }}
				<a class="link align-self-flex-start" href="{{ call .MakeURLWithQuery "x_login_id_input_type" .x_login_page_text_login_id_input_type }}">{{ localize "use-text-login-id-description" .x_login_page_text_login_id_variant }}</a>
				{{ end }}{{ end }}{{ end }}

				{{ if .x_login_id_input_type }}{{ if not (eq .x_login_id_input_type "phone") }}{{ if .x_login_page_login_id_has_phone }}
				<a class="link align-self-flex-start" href="{{ call .MakeURLWithQuery "x_login_id_input_type" "phone" }}">{{ localize "use-phone-login-id-description" }}</a>
				{{ end }}{{ end }}{{ end }}

				<div class="link">
					<span class="primary-text">{{ localize "signup-button-hint" }}</span>
					<a href="{{ call .MakeURLWithPathWithoutX "/signup" }}">{{ localize "signup-button-label" }}</a>
				</div>

				{{ if .x_password_authenticator_enabled }}
				<a class="link align-self-flex-start" href="{{ call .MakeURLWithPathWithoutX "/forgot_password" }}">{{ localize "forgot-password-button-label" }}</a>
				{{ end }}

				{{ if or .x_login_page_login_id_has_phone (not (eq .x_login_page_text_login_id_variant "none")) }}
				<button class="btn primary-btn align-self-flex-end" type="submit" name="submit" value="">{{ localize "next-button-label" }}</button>
				{{ end }}
			</form>
		</div>
		{{ template "auth_ui_footer.html" . }}
	</div>
</body>
</html>
`,
}

const LoginWithLoginIDRequest = "LoginWithLoginIDRequest"

var LoginSchema = validation.NewMultipartSchema("").
	Add(LoginWithLoginIDRequest, `
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

func ConfigureLoginRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/login")
}

type LoginHandler struct {
	Database *db.Handle
}

func (h *LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Database.WithTx(func() error {
		// FIXME(webapp): login
		// if r.Method == "GET" {
		// 	writeResponse, err := h.Provider.GetLoginForm(w, r)
		// 	writeResponse(err)
		// 	return err
		// }

		// if r.Method == "POST" {
		// 	if r.Form.Get("x_idp_id") != "" {
		// 		writeResponse, err := h.Provider.LoginIdentityProvider(w, r, r.Form.Get("x_idp_id"))
		// 		writeResponse(err)
		// 		return err
		// 	}

		// 	writeResponse, err := h.Provider.LoginWithLoginID(w, r)
		// 	writeResponse(err)
		// 	return err
		// }

		return nil
	})

	return
}
