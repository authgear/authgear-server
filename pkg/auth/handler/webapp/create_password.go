package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/webapp"
	"github.com/authgear/authgear-server/pkg/db"
	"github.com/authgear/authgear-server/pkg/httproute"
	"github.com/authgear/authgear-server/pkg/template"
)

const (
	// nolint: gosec
	TemplateItemTypeAuthUICreatePasswordHTML config.TemplateItemType = "auth_ui_create_password.html"
)

var TemplateAuthUICreatePasswordHTML = template.Spec{
	Type:        TemplateItemTypeAuthUICreatePasswordHTML,
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
	<button class="btn back-btn" type="button" title="{{ "back-button-title" }}"></button>
	<div class="login-id primary-txt">
	<!-- FIXME(webapp): show login ID -->
	{{ if .x_national_number }}
		+{{ .x_calling_code}} {{ .x_national_number }}
	{{ else }}
		{{ .x_login_id }}
	{{ end }}
	</div>
</div>

<div class="title primary-txt">{{ localize "create-password-page-title" }}</div>

{{ template "ERROR" . }}

<input id="password" data-password-policy-password="" class="input text-input primary-txt" type="password" name="x_password" placeholder="{{ localize "password-placeholder" }}">

<button class="btn secondary-btn password-visibility-btn show-password" type="button">{{ localize "show-password" }}</button>
<button class="btn secondary-btn password-visibility-btn hide-password" type="button">{{ localize "hide-password" }}</button>

{{ template "PASSWORD_POLICY" . }}

<button class="btn primary-btn align-self-flex-end" type="submit" name="submit" value="">{{ localize "next-button-label" }}</button>

</form>
{{ template "auth_ui_footer.html" . }}

</div>
</body>
</html>
`,
}

func ConfigureCreatePasswordRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/create_password")
}

type CreatePasswordViewModel struct {
	BaseViewModel
	PasswordPolicyViewModel
}

type CreatePasswordHandler struct {
	Database                *db.Handle
	State                   webapp.StateProvider
	BaseViewModel           *BaseViewModeler
	PasswordPolicyViewModel *PasswordPolicyViewModeler
	Renderer                Renderer
}

func (h *CreatePasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
		passwordPolicyViewModel := h.PasswordPolicyViewModel.ViewModel(state.Error)

		data := map[string]interface{}{}

		EmbedForm(data, r.Form)
		Embed(data, baseViewModel)
		Embed(data, passwordPolicyViewModel)

		h.Renderer.Render(w, r, TemplateItemTypeAuthUICreatePasswordHTML, data)
		return
	}

	// FIXME(webapp): create_password
	// h.Database.WithTx(func() error {

	// 	// if r.Method == "POST" {
	// 	// 	writeResponse, err := h.Provider.EnterSecret(w, r)
	// 	// 	writeResponse(err)
	// 	// 	return err
	// 	// }

	// 	return nil
	// })
}
