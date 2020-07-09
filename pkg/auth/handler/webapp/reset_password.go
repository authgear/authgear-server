package webapp

import (
	"errors"
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/webapp"
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

type ResetPasswordInteractions interface {
	ResetPassword(code string, newPassword string) error
}

type ResetPasswordHandler struct {
	Database                *db.Handle
	State                   webapp.StateProvider
	BaseViewModel           *BaseViewModeler
	PasswordPolicyViewModel *PasswordPolicyViewModeler
	Renderer                Renderer
	ResetPassword           ResetPasswordInteractions
}

func (h *ResetPasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Method == "GET" {
		state, err := h.State.RestoreState(r, true)
		if errors.Is(err, webapp.ErrStateNotFound) {
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
		passwordPolicyViewModel := h.PasswordPolicyViewModel.ViewModel(anyError)

		data := map[string]interface{}{}

		Embed(data, baseViewModel)
		Embed(data, passwordPolicyViewModel)

		h.Renderer.Render(w, r, TemplateItemTypeAuthUIResetPasswordHTML, data)
		return
	}

	if r.Method == "POST" {
		h.Database.WithTx(func() error {
			var state *webapp.State
			var err error

			defer func() {
				h.State.UpdateState(state, nil, err)
				if err != nil {
					webapp.RedirectToCurrentPath(w, r)
				} else {
					// Remove code from URL
					u := r.URL
					q := u.Query()
					q.Del("code")
					u.RawQuery = q.Encode()
					r.URL = u
					webapp.RedirectToPathWithX(w, r, "/reset_password/success")
				}
			}()
			state = h.State.CreateState(r, nil, nil)

			err = ResetPasswordSchema.PartValidator(ResetPasswordRequestSchema).ValidateValue(FormToJSON(r.Form))
			if err != nil {
				return err
			}

			code := r.Form.Get("code")
			newPassword := r.Form.Get("x_password")

			err = h.ResetPassword.ResetPassword(code, newPassword)
			if err != nil {
				return err
			}

			return nil
		})
	}
}
