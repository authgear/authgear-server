package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/db"
	"github.com/authgear/authgear-server/pkg/httproute"
	"github.com/authgear/authgear-server/pkg/template"
)

const (
	// nolint: gosec
	TemplateItemTypeAuthUIResetPasswordSuccessHTML config.TemplateItemType = "auth_ui_reset_password_success.html"
)

var TemplateAuthUIResetPasswordSuccessHTML = template.Spec{
	Type:        TemplateItemTypeAuthUIResetPasswordSuccessHTML,
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

<div class="title primary-txt">{{ localize "reset-password-success-page-title" }}</div>

{{ template "ERROR" . }}

<!-- FIXME: x_login_id -->

<div class="description primary-txt">{{ localize "reset-password-success-description" .x_login_id }}</div>

</div>
{{ template "auth_ui_footer.html" . }}

</div>
</body>
</html>
`,
}

func ConfigureResetPasswordSuccessRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "GET").
		WithPathPattern("/reset_password/success")
}

type ResetPasswordSuccessHandler struct {
	Database *db.Handle
}

func (h *ResetPasswordSuccessHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Database.WithTx(func() error {
		// FIXME(webapp): reset_password_success
		// if r.Method == "GET" {
		// 	writeResponse, err := h.Provider.GetResetPasswordSuccess(w, r)
		// 	writeResponse(err)
		// 	return err
		// }
		return nil
	})
}
