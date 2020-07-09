package webapp

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/auth/config"
	interactionflows "github.com/authgear/authgear-server/pkg/auth/dependency/interaction/flows"
	"github.com/authgear/authgear-server/pkg/auth/dependency/webapp"
	"github.com/authgear/authgear-server/pkg/core/phone"
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
				{{ range $.IdentityCandidates }}
				{{ if eq .type "oauth" }}
				<form class="authorize-idp-form" method="post" novalidate>
				{{ $.CSRFField }}
				<button class="btn sso-btn {{ .provider_type }}" type="submit" name="x_provider_alias" value="{{ .provider_alias }}" data-form-xhr="false">
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
			{{ range $.IdentityCandidates }}
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
				{{ $.CSRFField }}

				{{ if $.x_login_id_input_type }}{{ if eq $.x_login_id_input_type "phone" }}{{ if $.LoginPageLoginIDHasPhone }}
				<div class="phone-input">
					<select class="input select primary-txt" name="x_calling_code">
						{{ range $.CountryCallingCodes }}
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

				{{ if $.x_login_id_input_type }}{{ if not (eq $.x_login_id_input_type "phone") }}{{ if (not (eq $.LoginPageTextLoginIDVariant "none")) }}
				<input class="input text-input primary-txt" type="{{ $.LoginPageTextLoginIDInputType }}" name="x_login_id" placeholder="{{ localize "login-id-placeholder" $.LoginPageTextLoginIDVariant }}">
				{{ end }}{{ end }}{{ end }}

				{{ if $.x_login_id_input_type }}{{ if eq $.x_login_id_input_type "phone" }}{{ if (not (eq $.LoginPageTextLoginIDVariant "none")) }}
				<a class="link align-self-flex-start" href="{{ call $.MakeURLWithQuery "x_login_id_input_type" $.LoginPageTextLoginIDInputType }}">{{ localize "use-text-login-id-description" $.LoginPageTextLoginIDVariant }}</a>
				{{ end }}{{ end }}{{ end }}

				{{ if $.x_login_id_input_type }}{{ if not (eq $.x_login_id_input_type "phone") }}{{ if $.LoginPageLoginIDHasPhone }}
				<a class="link align-self-flex-start" href="{{ call $.MakeURLWithQuery "x_login_id_input_type" "phone" }}">{{ localize "use-phone-login-id-description" }}</a>
				{{ end }}{{ end }}{{ end }}

				<div class="link">
					<span class="primary-text">{{ localize "signup-button-hint" }}</span>
					<a href="{{ call $.MakeURLWithPathWithoutX "/signup" }}">{{ localize "signup-button-label" }}</a>
				</div>

				{{ if $.PasswordAuthenticatorEnabled }}
				<a class="link align-self-flex-start" href="{{ call $.MakeURLWithPathWithoutX "/forgot_password" }}">{{ localize "forgot-password-button-label" }}</a>
				{{ end }}

				{{ if or $.LoginPageLoginIDHasPhone (not (eq $.LoginPageTextLoginIDVariant "none")) }}
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

const LoginWithLoginIDRequestSchema = "LoginWithLoginIDRequestSchema"

var LoginSchema = validation.NewMultipartSchema("").
	Add(LoginWithLoginIDRequestSchema, `
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

type LoginOAuthService interface {
	LoginOAuthProvider(w http.ResponseWriter, r *http.Request, providerAlias string, state *webapp.State) (*interactionflows.WebAppResult, error)
}

type LoginInteractions interface {
	LoginWithLoginID(loginID string) (*interactionflows.WebAppResult, error)
}

type LoginHandler struct {
	Database                *db.Handle
	State                   webapp.StateProvider
	BaseViewModel           *BaseViewModeler
	AuthenticationViewModel *AuthenticationViewModeler
	FormPrefiller           *FormPrefiller
	Renderer                Renderer
	OAuth                   LoginOAuthService
	Interactions            LoginInteractions
	Responder               Responder
}

func (h *LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.FormPrefiller.Prefill(r.Form)

	if r.Method == "GET" {
		state, err := h.State.RestoreState(r, true)
		if errors.Is(err, webapp.ErrStateNotFound) {
			err = nil
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var anyError interface{}
		if state != nil {
			anyError = state.Error
		}

		baseViewModel := h.BaseViewModel.ViewModel(r, anyError)
		authenticationViewModel := h.AuthenticationViewModel.ViewModel(r)

		data := map[string]interface{}{}

		EmbedForm(data, r.Form)
		Embed(data, baseViewModel)
		Embed(data, authenticationViewModel)

		h.Renderer.Render(w, r, TemplateItemTypeAuthUILoginHTML, data)
		return
	}

	providerAlias := r.Form.Get("x_provider_alias")

	if r.Method == "POST" && providerAlias != "" {
		h.Database.WithTx(func() error {
			var state *webapp.State
			var result *interactionflows.WebAppResult
			var err error

			defer func() {
				h.State.UpdateState(state, result, err)
				h.Responder.Respond(w, r, state, result, err)
			}()
			state = h.State.CreateState(r, nil, nil)

			result, err = h.OAuth.LoginOAuthProvider(w, r, providerAlias, state)
			if err != nil {
				return err
			}

			return nil
		})
		return
	}

	if r.Method == "POST" {
		h.Database.WithTx(func() error {
			var state *webapp.State
			var result *interactionflows.WebAppResult
			var err error

			defer func() {
				h.State.UpdateState(state, result, err)
				h.Responder.Respond(w, r, state, result, err)
			}()
			state = h.State.CreateState(r, nil, nil)

			err = LoginSchema.PartValidator(LoginWithLoginIDRequestSchema).ValidateValue(FormToJSON(r.Form))
			if err != nil {
				return err
			}

			loginID, err := FormToLoginID(r.Form)
			if err != nil {
				return err
			}

			result, err = h.Interactions.LoginWithLoginID(loginID)
			if err != nil {
				return err
			}

			return nil
		})
	}

	return
}

// FormToLoginID returns the raw login ID or the parsed phone number.
func FormToLoginID(form url.Values) (loginID string, err error) {
	if form.Get("x_login_id_input_type") == "phone" {
		nationalNumber := form.Get("x_national_number")
		countryCallingCode := form.Get("x_calling_code")
		var e164 string
		e164, err = phone.Parse(nationalNumber, countryCallingCode)
		if err != nil {
			err = &validation.AggregatedError{
				Errors: []validation.Error{{
					Keyword:  "format",
					Location: "/x_national_number",
					Info: map[string]interface{}{
						"format": "phone",
					},
				}},
			}
			return
		}
		loginID = e164
		return
	}

	loginID = form.Get("x_login_id")
	return
}
