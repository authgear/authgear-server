package webapp

import (
	"errors"
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/config"
	interactionflows "github.com/authgear/authgear-server/pkg/auth/dependency/interaction/flows"
	"github.com/authgear/authgear-server/pkg/httproute"
	"github.com/authgear/authgear-server/pkg/template"
)

const (
	TemplateItemTypeAuthUIRegisterTOTPHTML config.TemplateItemType = "auth_ui_register_totp.html"
)

var TemplateAuthUIRegisterTOTPHTML = template.Spec{
	Type:        TemplateItemTypeAuthUIRegisterTOTPHTML,
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

<div class="nav-bar">
	<button class="btn back-btn" type="button" title="{{ localize "back-button-title" }}"></button>
</div>

<div class="title primary-txt">{{ localize "register-totp-page-title" }}</div>

{{ template "ERROR" . }}

<div class="description primary-txt">{{ localize "register-totp-page-description" }}</div>

<!-- TODO(mfa): Use the real QR code image -->
<img class="img-qr-code align-self-center" src="https://via.placeholder.com/256">

<form class="vertical-form form-fields-container" method="post" novalidate>

{{ $.CSRFField }}

<input class="input text-input primary-txt" type="text" inputmode="numeric" pattern="[0-9]*" name="x_password" placeholder="{{ localize "register-totp-placeholder" }}">

<button class="btn primary-btn align-self-flex-end" type="submit" name="submit" value="">{{ localize "next-button-label" }}</button>

</form>

{{ template "auth_ui_footer.html" . }}

</div>
</body>
</html>
`,
}

func ConfigureRegisterTOTPRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/register_totp")
}

type RegisterTOTPHandler struct {
	State         StateService
	BaseViewModel *BaseViewModeler
	Renderer      Renderer
}

func (h *RegisterTOTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Method == "GET" {
		// TODO(mfa): Make state required
		state, err := h.State.RestoreReadOnlyState(r, true)
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

		data := map[string]interface{}{}
		Embed(data, baseViewModel)

		h.Renderer.Render(w, r, TemplateItemTypeAuthUIRegisterTOTPHTML, data)
		return
	}
}
