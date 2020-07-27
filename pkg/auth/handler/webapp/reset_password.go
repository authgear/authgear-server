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
	TemplateItemTypeAuthUIResetPasswordHTML config.TemplateItemType = "auth_ui_reset_password.html"
)

var TemplateAuthUIResetPasswordHTML = template.Spec{
	Type:        TemplateItemTypeAuthUIResetPasswordHTML,
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

<div class="title primary-txt">{{ localize "reset-password-page-title" }}</div>

{{ template "ERROR" . }}

<div class="description primary-txt">{{ localize "reset-password-description" }}</div>

<input id="password" data-password-policy-password="" class="input text-input primary-txt" type="password" name="x_password" placeholder="{{ localize "password-placeholder" }}">

<button class="btn secondary-btn password-visibility-btn show-password" type="button">{{ localize "show-password" }}</button>
<button class="btn secondary-btn password-visibility-btn hide-password" type="button">{{ localize "hide-password" }}</button>

{{ template "PASSWORD_POLICY" . }}

<button class="btn primary-btn submit-btn align-self-flex-end" type="submit" name="submit" value="">{{ localize "next-button-label" }}</button>

</form>

{{ template "auth_ui_footer.html" . }}

</div>
</body>
</html>
`,
}

const ResetPasswordRequestSchema = "ResetPasswordRequestSchema"

var ResetPasswordSchema = validation.NewMultipartSchema("").
	Add(ResetPasswordRequestSchema, `
		{
			"type": "object",
			"properties": {
				"code": { "type": "string" },
				"x_password": { "type": "string" }
			},
			"required": ["code", "x_password"]
		}
	`).Instantiate()

func ConfigureResetPasswordRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/reset_password")
}

type ResetPasswordHandler struct {
	Database      *db.Handle
	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      Renderer
	WebApp        WebAppService
}

func (h *ResetPasswordHandler) MakeIntent(r *http.Request) *webapp.Intent {
	return &webapp.Intent{
		RedirectURI: "/reset_password/success",
		KeepState:   true,
		// FIXME(webapp): IntentResetPassword
		Intent: intents.NewIntentLogin(),
	}
}

func (h *ResetPasswordHandler) GetData(r *http.Request, state *webapp.State, graph *newinteraction.Graph, edges []newinteraction.Edge) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	var anyError interface{}
	if state != nil {
		anyError = state.Error
	}
	baseViewModel := h.BaseViewModel.ViewModel(r, anyError)
	// FIXME(webapp): derive PasswordPolicyViewModel with graph and edges
	passwordPolicyViewModel := viewmodels.PasswordPolicyViewModel{}
	viewmodels.Embed(data, baseViewModel)
	viewmodels.Embed(data, passwordPolicyViewModel)
	return data, nil
}

// FIXME(webapp): implement input interface
type ResetPasswordInput struct {
	Code     string
	Password string
}

func (h *ResetPasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	intent := h.MakeIntent(r)

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

			h.Renderer.Render(w, r, TemplateItemTypeAuthUIResetPasswordHTML, data)
			return nil
		})
	}

	if r.Method == "POST" {
		h.Database.WithTx(func() error {
			result, err := h.WebApp.PostIntent(intent, func() (input interface{}, err error) {
				err = ResetPasswordSchema.PartValidator(ResetPasswordRequestSchema).ValidateValue(FormToJSON(r.Form))
				if err != nil {
					return
				}

				code := r.Form.Get("code")
				newPassword := r.Form.Get("x_password")

				input = &ResetPasswordInput{
					Code:     code,
					Password: newPassword,
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
