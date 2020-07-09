package webapp

import (
	"errors"
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/config"
	interactionflows "github.com/authgear/authgear-server/pkg/auth/dependency/interaction/flows"
	"github.com/authgear/authgear-server/pkg/auth/dependency/webapp"
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
{{ $.CSRFField }}

<div class="nav-bar">
	<button class="btn back-btn" type="button" title="{{ localize "back-button-title" }}"></button>
</div>

<div class="title primary-txt">{{ localize "forgot-password-page-title" }}</div>

{{ template "ERROR" . }}

{{ if $.x_login_id_input_type }}{{ if eq $.x_login_id_input_type "phone" }}{{ if $.LoginPageLoginIDHasPhone }}
<div class="description primary-txt">{{ localize "forgot-password-phone-description" }}</div>
<div class="phone-input">
	<select class="input select primary-txt" name="x_calling_code">
		{{ range .CountryCallingCodes }}
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

{{ if $.x_login_id_input_type }}{{ if (not (eq $.x_login_id_input_type "phone")) }}{{ if or (eq $.LoginPageTextLoginIDVariant "email") (eq $.LoginPageTextLoginIDVariant "email_or_username") }}
<div class="description primary-txt">{{ localize "forgot-password-email-description" }}</div>
<input class="input text-input primary-txt" type="{{ $.x_login_id_input_type }}" name="x_login_id" placeholder="{{ localize "email-placeholder" }}">
{{ end }}{{ end }}{{ end }}

{{ if $.x_login_id_input_type }}{{ if eq $.x_login_id_input_type "phone" }}{{ if or (eq $.LoginPageTextLoginIDVariant "email") (eq $.LoginPageTextLoginIDVariant "email_or_username") }}
<a class="link align-self-flex-start" href="{{ call $.MakeURLWithQuery "x_login_id_input_type" "email" }}">{{ localize "use-email-login-id-description" }}</a>
{{ end }}{{ end }}{{ end }}

{{ if $.x_login_id_input_type }}{{ if eq $.x_login_id_input_type "email" }}{{ if $.LoginPageLoginIDHasPhone }}
<a class="link align-self-flex-start" href="{{ call $.MakeURLWithQuery "x_login_id_input_type" "phone" }}">{{ localize "use-phone-login-id-description" }}</a>
{{ end }}{{ end }}{{ end }}

{{ if or $.LoginPageLoginIDHasPhone (not (eq $.LoginPageTextLoginIDVariant "none")) }}
<button class="btn primary-btn submit-btn align-self-flex-end" type="submit" name="submit" value="">{{ localize "next-button-label" }}</button>
{{ end }}

</form>
{{ template "auth_ui_footer.html" . }}

</div>
</body>
</html>
`,
}

const ForgotPasswordRequestSchema = "ForgotPasswordRequestSchema"

var ForgotPasswordSchema = validation.NewMultipartSchema("").
	Add(ForgotPasswordRequestSchema, `
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

type ForgotPasswordInteractions interface {
	SendCode(loginID string) error
}

type ForgotPasswordHandler struct {
	Database                *db.Handle
	State                   StateService
	BaseViewModel           *BaseViewModeler
	AuthenticationViewModel *AuthenticationViewModeler
	FormPrefiller           *FormPrefiller
	Renderer                Renderer
	ForgotPassword          ForgotPasswordInteractions
}

func (h *ForgotPasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.FormPrefiller.Prefill(r.Form)

	if r.Method == "GET" {
		state, err := h.State.RestoreState(r, true)
		if errors.Is(err, interactionflows.ErrStateNotFound) {
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

		h.Renderer.Render(w, r, TemplateItemTypeAuthUIForgotPasswordHTML, data)
		return
	}

	if r.Method == "POST" {
		h.Database.WithTx(func() error {
			var state *interactionflows.State
			var err error

			defer func() {
				h.State.UpdateState(state, nil, err)
				if err != nil {
					webapp.RedirectToCurrentPath(w, r)
				} else {
					webapp.RedirectToPathWithX(w, r, "/forgot_password/success")
				}
			}()
			state = h.State.CreateState(r, nil, nil)

			err = ForgotPasswordSchema.PartValidator(ForgotPasswordRequestSchema).ValidateValue(FormToJSON(r.Form))
			if err != nil {
				return err
			}

			loginID, err := FormToLoginID(r.Form)
			if err != nil {
				return err
			}

			err = h.ForgotPassword.SendCode(loginID)
			if err != nil {
				return err
			}

			return nil
		})
	}
}
