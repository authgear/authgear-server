package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/auth/dependency/webapp"
	"github.com/authgear/authgear-server/pkg/httproute"
	"github.com/authgear/authgear-server/pkg/template"
)

const (
	// nolint: gosec
	TemplateItemTypeAuthUIForgotPasswordSuccessHTML config.TemplateItemType = "auth_ui_forgot_password_success.html"
)

var TemplateAuthUIForgotPasswordSuccessHTML = template.Spec{
	Type:        TemplateItemTypeAuthUIForgotPasswordSuccessHTML,
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

<div class="title primary-txt">{{ localize "forgot-password-success-page-title" }}</div>

{{ template "ERROR" . }}

<div class="description primary-txt">{{ localize "forgot-password-success-description" $.GivenLoginID }}</div>

<a class="btn primary-btn align-self-flex-end" href="{{ call .MakeURLWithPathWithoutX "/login" }}">{{ localize "login-button-label--forgot-password-success-page" }}</a>

</div>
{{ template "auth_ui_footer.html" . }}

</div>
</body>
</html>
`,
}

func ConfigureForgotPasswordSuccessRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "GET").
		WithPathPattern("/forgot_password/success")
}

type ForgotPasswordSuccessViewModel struct {
	GivenLoginID string
}

type ForgotPasswordSuccessHandler struct {
	BaseViewModel *BaseViewModeler
	Renderer      Renderer
	WebApp        WebAppService
}

func (h *ForgotPasswordSuccessHandler) GetData(r *http.Request, state *webapp.State, graph *newinteraction.Graph, edges []newinteraction.Edge) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	baseViewModel := h.BaseViewModel.ViewModel(r, state.Error)
	// FIXME(webapp): derive ForgotPasswordSuccessViewModel with graph and edges
	forgotPasswordSuccessViewModel := ForgotPasswordSuccessViewModel{}
	Embed(data, baseViewModel)
	Embed(data, forgotPasswordSuccessViewModel)
	return data, nil
}

func (h *ForgotPasswordSuccessHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

		h.Renderer.Render(w, r, TemplateItemTypeAuthUIForgotPasswordSuccessHTML, data)
		return
	}
}
