package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/config"
	interactionflows "github.com/authgear/authgear-server/pkg/auth/dependency/interaction/flows"
	"github.com/authgear/authgear-server/pkg/db"
	"github.com/authgear/authgear-server/pkg/httproute"
	"github.com/authgear/authgear-server/pkg/template"
	"github.com/authgear/authgear-server/pkg/validation"
)

const (
	// nolint: gosec
	TemplateItemTypeAuthUIEnterPasswordHTML config.TemplateItemType = "auth_ui_enter_password.html"
)

var TemplateAuthUIEnterPasswordHTML = template.Spec{
	Type:        TemplateItemTypeAuthUIEnterPasswordHTML,
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
	<div class="login-id primary-txt">
	<!-- FIXME(webapp): show login ID -->
	{{ if .x_national_number }}
		+{{ .x_calling_code}} {{ .x_national_number }}
	{{ else }}
		{{ .x_login_id }}
	{{ end }}
	</div>
</div>

<div class="title primary-txt">{{ localize "enter-password-page-title" }}</div>

{{ template "ERROR" . }}

<input id="password" class="input text-input primary-txt" type="password" name="x_password" placeholder="{{ localize "password-placeholder" }}">

<button class="btn secondary-btn password-visibility-btn show-password" type="button">{{ localize "show-password" }}</button>
<button class="btn secondary-btn password-visibility-btn hide-password" type="button">{{ localize "hide-password" }}</button>

{{ if $.PasswordAuthenticatorEnabled }}
<a class="link align-self-flex-start" href="{{ call $.MakeURLWithPathWithoutX "/forgot_password" }}">{{ localize "forgot-password-button-label--enter-password-page" }}</a>
{{ end }}

<button class="btn primary-btn align-self-flex-end" type="submit" name="submit" value="">{{ localize "next-button-label" }}</button>

</form>
{{ template "auth_ui_footer.html" . }}

</div>
</body>
</html>
`,
}

const EnterPasswordRequestSchema = "EnterPasswordRequestSchema"

var EnterPasswordSchema = validation.NewMultipartSchema("").
	Add(EnterPasswordRequestSchema, `
		{
			"type": "object",
			"properties": {
				"x_password": { "type": "string" }
			},
			"required": ["x_password"]
		}
	`).Instantiate()

func ConfigureEnterPasswordRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/enter_password")
}

type EnterPasswordInteractions interface {
	EnterSecret(state *interactionflows.State, password string) (*interactionflows.WebAppResult, error)
}

type EnterPasswordHandler struct {
	Database                *db.Handle
	State                   StateService
	BaseViewModel           *BaseViewModeler
	AuthenticationViewModel *AuthenticationViewModeler
	Renderer                Renderer
	Interactions            EnterPasswordInteractions
	Responder               Responder
}

func (h *EnterPasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Method == "GET" {
		state, err := h.State.RestoreState(r, false)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		baseViewModel := h.BaseViewModel.ViewModel(r, state.Error)
		authenticationViewModel := h.AuthenticationViewModel.ViewModel(r)

		data := map[string]interface{}{}

		Embed(data, baseViewModel)
		Embed(data, authenticationViewModel)

		h.Renderer.Render(w, r, TemplateItemTypeAuthUIEnterPasswordHTML, data)
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

			err = EnterPasswordSchema.PartValidator(EnterPasswordRequestSchema).ValidateValue(FormToJSON(r.Form))
			if err != nil {
				return err
			}

			plainPassword := r.Form.Get("x_password")

			result, err = h.Interactions.EnterSecret(state, plainPassword)
			if err != nil {
				return err
			}

			return nil
		})
	}
}
