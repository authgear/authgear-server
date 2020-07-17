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
	{{ .GivenLoginID }}
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

const CreatePasswordRequestSchema = "CreatePasswordRequestSchema"

var CreatePasswordSchema = validation.NewMultipartSchema("").
	Add(CreatePasswordRequestSchema, `
		{
			"type": "object",
			"properties": {
				"x_password": { "type": "string" }
			},
			"required": ["x_password"]
		}
	`).Instantiate()

func ConfigureCreatePasswordRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/create_password")
}

type CreatePasswordInteractions interface {
	EnterSecret(state *interactionflows.State, password string) (*interactionflows.WebAppResult, error)
}

type CreatePasswordViewModel struct {
	GivenLoginID string
}

func NewCreatePasswordViewModel(state *interactionflows.State) CreatePasswordViewModel {
	givenLoginID, _ := state.Extra[interactionflows.ExtraGivenLoginID].(string)
	return CreatePasswordViewModel{
		GivenLoginID: givenLoginID,
	}
}

type CreatePasswordHandler struct {
	Database                *db.Handle
	State                   StateService
	BaseViewModel           *BaseViewModeler
	PasswordPolicyViewModel *PasswordPolicyViewModeler
	Renderer                Renderer
	Interactions            CreatePasswordInteractions
	Responder               Responder
}

func (h *CreatePasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
		passwordPolicyViewModel := h.PasswordPolicyViewModel.ViewModel(state.Error)
		createPasswordViewModel := NewCreatePasswordViewModel(state)

		data := map[string]interface{}{}

		EmbedForm(data, r.Form)
		Embed(data, baseViewModel)
		Embed(data, passwordPolicyViewModel)
		Embed(data, createPasswordViewModel)

		h.Renderer.Render(w, r, TemplateItemTypeAuthUICreatePasswordHTML, data)
		return
	}

	if r.Method == "POST" {
		h.Database.WithTx(func() error {
			state, err := h.State.CloneState(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return err
			}

			var result *interactionflows.WebAppResult
			defer func() {
				h.State.UpdateState(state, result, err)
				h.Responder.Respond(w, r, state, result, err)
			}()

			err = CreatePasswordSchema.PartValidator(CreatePasswordRequestSchema).ValidateValue(FormToJSON(r.Form))
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
