package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/config"
	interactionflows "github.com/authgear/authgear-server/pkg/auth/dependency/interaction/flows"
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

func NewForgotPasswordSuccessViewModel(state *interactionflows.State) ForgotPasswordSuccessViewModel {
	givenLoginID, _ := state.Extra[interactionflows.ExtraGivenLoginID].(string)
	return ForgotPasswordSuccessViewModel{
		GivenLoginID: givenLoginID,
	}
}

type ForgotPasswordSuccessHandler struct {
	State         StateService
	BaseViewModel *BaseViewModeler
	Renderer      Renderer
}

func (h *ForgotPasswordSuccessHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Method == "GET" {
		state, err := h.State.RestoreReadOnlyState(r, false)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		baseViewModel := h.BaseViewModel.ViewModel(r, state.Error)
		forgotPasswordSuccessViewModel := NewForgotPasswordSuccessViewModel(state)

		data := map[string]interface{}{}

		Embed(data, baseViewModel)
		Embed(data, forgotPasswordSuccessViewModel)

		h.Renderer.Render(w, r, TemplateItemTypeAuthUIForgotPasswordSuccessHTML, data)
		return
	}
}
