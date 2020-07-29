package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction/intents"
	"github.com/authgear/authgear-server/pkg/auth/dependency/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
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

type ForgotPasswordHandler struct {
	Database      *db.Handle
	BaseViewModel *viewmodels.BaseViewModeler
	FormPrefiller *FormPrefiller
	Renderer      Renderer
	WebApp        WebAppService
}

func (h *ForgotPasswordHandler) MakeIntent(r *http.Request) *webapp.Intent {
	return &webapp.Intent{
		RedirectURI: "/forgot_password/success",
		KeepState:   true,
		Intent:      intents.NewIntentForgotPassword(),
	}
}

func (h *ForgotPasswordHandler) GetData(r *http.Request, state *webapp.State, graph *newinteraction.Graph, edges []newinteraction.Edge) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	var anyError interface{}
	if state != nil {
		anyError = state.Error
	}
	baseViewModel := h.BaseViewModel.ViewModel(r, anyError)
	authenticationViewModel := viewmodels.NewAuthenticationViewModel(edges)
	viewmodels.EmbedForm(data, r.Form)
	viewmodels.Embed(data, baseViewModel)
	viewmodels.Embed(data, authenticationViewModel)
	return data, nil
}

type ForgotPasswordLoginID struct {
	LoginID string
}

// GetLoginID implements InputForgotPasswordSelectLoginID.
func (i *ForgotPasswordLoginID) GetLoginID() string {
	return i.LoginID
}

func (h *ForgotPasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	intent := h.MakeIntent(r)

	h.FormPrefiller.Prefill(r.Form)

	if r.Method == "GET" {
		h.Database.WithTx(func() error {
			state, graph, edges, err := h.WebApp.GetIntent(intent, StateID(r))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return err
			}

			data, err := h.GetData(r, state, graph, edges)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return err
			}

			h.Renderer.Render(w, r, TemplateItemTypeAuthUIForgotPasswordHTML, data)
			return nil
		})
	}

	if r.Method == "POST" {
		h.Database.WithTx(func() error {
			result, err := h.WebApp.PostIntent(intent, func() (input interface{}, err error) {
				err = ForgotPasswordSchema.PartValidator(ForgotPasswordRequestSchema).ValidateValue(FormToJSON(r.Form))
				if err != nil {
					return
				}

				loginID, err := FormToLoginID(r.Form)
				if err != nil {
					return
				}

				input = &ForgotPasswordLoginID{
					LoginID: loginID,
				}
				return
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return err
			}
			result.WriteResponse(w, r)
			return nil
		})
	}
}
