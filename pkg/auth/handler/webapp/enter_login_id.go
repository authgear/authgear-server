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
	TemplateItemTypeAuthUIEnterLoginIDHTML config.TemplateItemType = "auth_ui_enter_login_id.html"
)

var TemplateAuthUIEnterLoginIDHTML = template.Spec{
	Type:        TemplateItemTypeAuthUIEnterLoginIDHTML,
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

<!-- FIXME: x_old_login_id_value, x_login_id_key, x_login_id_type, x_login_id_input_type -->

<div class="title primary-txt">
	{{ if .x_old_login_id_value }}
	{{ localize "enter-login-id-page-title--change" .x_login_id_key }}
	{{ else }}
	{{ localize "enter-login-id-page-title--add" .x_login_id_key }}
	{{ end }}
</div>

{{ template "ERROR" . }}

<form class="vertical-form form-fields-container" method="post" novalidate>

{{ $.csrfField }}
<input type="hidden" name="x_login_id_key" value="{{ .x_login_id_key }}">
<input type="hidden" name="x_login_id_type" value="{{ .x_login_id_type }}">
<input type="hidden" name="x_login_id_input_type" value="{{ .x_login_id_input_type }}">
<input type="hidden" name="x_old_login_id_value" value="{{ .x_old_login_id_value }}">

{{ if eq .x_login_id_input_type "phone" }}
<div class="phone-input">
	<select class="input select primary-txt" name="x_calling_code">
		{{ range .x_calling_codes }}
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
{{ else }}
<input class="input text-input primary-txt" type="{{ .x_login_id_input_type }}" name="x_login_id" placeholder="{{ localize "login-id-placeholder" .x_login_id_type }}">
{{ end }}

<button class="btn primary-btn align-self-flex-end" type="submit" name="submit" value="">{{ localize "next-button-label" }}</button>

</form>

{{ if .x_old_login_id_value }}
<form class="enter-login-id-remove-form" method="post" novalidate>
{{ $.csrfField }}
<input type="hidden" name="x_login_id_key" value="{{ .x_login_id_key }}">
<input type="hidden" name="x_login_id_type" value="{{ .x_login_id_type }}">
<input type="hidden" name="x_login_id_input_type" value="{{ .x_login_id_input_type }}">
<input type="hidden" name="x_old_login_id_value" value="{{ .x_old_login_id_value }}">
<button class="anchor" type="submit" name="x_action" value="remove">{{ localize "disconnect-button-label" }}</button>
{{ end }}
</form>

</div>
{{ template "auth_ui_footer.html" . }}

</div>
</body>
</html>
`,
}

const RemoveLoginIDRequest = "RemoveLoginIDRequest"

var EnterLoginIDSchema = validation.NewMultipartSchema("").
	Add(RemoveLoginIDRequest, `
		{
			"type": "object",
			"properties": {
				"x_login_id_key": { "type": "string" },
				"x_old_login_id_value": { "type": "string" }
			},
			"required": ["x_login_id_key", "x_old_login_id_value"]
		}
	`).Instantiate()

func ConfigureEnterLoginIDRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/enter_login_id")
}

type EnterLoginIDHandler struct {
	Database *db.Handle
}

func (h *EnterLoginIDHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Database.WithTx(func() error {
		// FIXME(webapp): enter_login_id
		// if r.Method == "GET" {
		// 	writeResponse, err := h.Provider.GetEnterLoginIDForm(w, r)
		// 	writeResponse(err)
		// 	return err
		// }

		// if r.Method == "POST" {
		// 	if r.Form.Get("x_action") == "remove" {
		// 		writeResponse, err := h.Provider.RemoveLoginID(w, r)
		// 		writeResponse(err)
		// 		return err
		// 	}

		// 	writeResponse, err := h.Provider.EnterLoginID(w, r)
		// 	writeResponse(err)
		// 	return err
		// }

		return nil
	})
}
