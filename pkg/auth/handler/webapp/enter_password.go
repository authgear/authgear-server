package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/auth/dependency/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
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
	{{ .GivenLoginID }}
	</div>
</div>

<div class="title primary-txt">{{ localize "enter-password-page-title" }}</div>

{{ template "ERROR" . }}

<input id="password" class="input text-input primary-txt" type="password" name="x_password" placeholder="{{ localize "password-placeholder" }}">

<button class="btn secondary-btn password-visibility-btn show-password" type="button">{{ localize "show-password" }}</button>
<button class="btn secondary-btn password-visibility-btn hide-password" type="button">{{ localize "hide-password" }}</button>

<!-- This page for entering password. So if the user reaches this page normally, forgot password link should be provided -->
<a class="link align-self-flex-start" href="{{ call $.MakeURLWithPathWithoutX "/forgot_password" }}">{{ localize "forgot-password-button-label--enter-password-page" }}</a>

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

type EnterPasswordViewModel struct {
	GivenLoginID string
}

type EnterPasswordHandler struct {
	Database      *db.Handle
	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      Renderer
	WebApp        WebAppService
}

func (h *EnterPasswordHandler) GetData(r *http.Request, state *webapp.State, graph *newinteraction.Graph, edges []newinteraction.Edge) (map[string]interface{}, error) {
	data := map[string]interface{}{}

	baseViewModel := h.BaseViewModel.ViewModel(r, state.Error)
	// FIXME(webapp): derive AuthenticationViewModel with graph and edges
	authenticationViewModel := viewmodels.AuthenticationViewModel{}
	// FIXME(webapp): derive EnterPasswordViewModel with graph and edges
	enterPasswordViewModel := EnterPasswordViewModel{}

	viewmodels.Embed(data, baseViewModel)
	viewmodels.Embed(data, authenticationViewModel)
	viewmodels.Embed(data, enterPasswordViewModel)

	return data, nil
}

// FIXME(webapp): implement input interface
type EnterPasswordInput struct {
	Password string
}

func (h *EnterPasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Method == "GET" {
		state, graph, edges, err := h.WebApp.Get(StateID(r))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data, err := h.GetData(r, state, graph, edges)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		h.Renderer.Render(w, r, TemplateItemTypeAuthUIEnterPasswordHTML, data)
		return

	}

	if r.Method == "POST" {
		h.Database.WithTx(func() error {
			result, err := h.WebApp.PostInput(StateID(r), func() (input interface{}, err error) {
				err = EnterPasswordSchema.PartValidator(EnterPasswordRequestSchema).ValidateValue(FormToJSON(r.Form))
				if err != nil {
					return
				}

				plainPassword := r.Form.Get("x_password")

				input = &EnterPasswordInput{
					Password: plainPassword,
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
