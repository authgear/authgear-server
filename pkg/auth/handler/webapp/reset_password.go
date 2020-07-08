package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/config"
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
{{ $.csrfField }}

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

const ResetPasswordRequest = "ResetPasswordRequest"

var ResetPasswordSchema = validation.NewMultipartSchema("").
	Add(ResetPasswordRequest, `
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
	Database *db.Handle
}

func (h *ResetPasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Database.WithTx(func() error {
		// FIXME(webapp): reset_password
		// if r.Method == "GET" {
		// 	writeResponse, err := h.Provider.GetResetPasswordForm(w, r)
		// 	writeResponse(err)
		// 	return err
		// }

		// if r.Method == "POST" {
		// 	writeResponse, err := h.Provider.PostResetPasswordForm(w, r)
		// 	writeResponse(err)
		// 	return err
		// }
		return nil
	})
}
