package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/auth"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity/loginid"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/auth/dependency/webapp"
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

<div class="title primary-txt">
	{{ if $.OldLoginIDValue }}
	{{ localize "enter-login-id-page-title--change" $.LoginIDKey }}
	{{ else }}
	{{ localize "enter-login-id-page-title--add" $.LoginIDKey }}
	{{ end }}
</div>

{{ template "ERROR" . }}

<form class="vertical-form form-fields-container" method="post" novalidate>

{{ $.CSRFField }}

{{ if eq .LoginIDInputType "phone" }}
<div class="phone-input">
	<select class="input select primary-txt" name="x_calling_code">
		{{ range $.CountryCallingCodes }}
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
<input class="input text-input primary-txt" type="{{ .LoginIDInputType }}" name="x_login_id" placeholder="{{ localize "login-id-placeholder" .LoginIDType }}">
{{ end }}

<button class="btn primary-btn align-self-flex-end" type="submit" name="x_action" value="add_or_update">{{ localize "next-button-label" }}</button>

</form>

{{ if .OldLoginIDValue }}
<form class="enter-login-id-remove-form" method="post" novalidate>
{{ $.CSRFField }}
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

type EnterLoginIDViewModel struct {
	LoginIDKey       string
	LoginIDType      string
	OldLoginIDValue  string
	LoginIDInputType string
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
	Database      *db.Handle
	BaseViewModel *BaseViewModeler
	Renderer      Renderer
	WebApp        WebAppService
}

func (h *EnterLoginIDHandler) GetData(r *http.Request, state *webapp.State, graph *newinteraction.Graph, edges []newinteraction.Edge) (map[string]interface{}, error) {
	data := map[string]interface{}{}
	baseViewModel := h.BaseViewModel.ViewModel(r, state.Error)
	// FIXME(webapp): derive EnterLoginIDViewModel with graph and edges
	enterLoginIDViewModel := EnterLoginIDViewModel{}

	Embed(data, baseViewModel)
	Embed(data, enterLoginIDViewModel)
	return data, nil
}

// FIXME(webapp): implement input interface
type EnterLoginIDRemoveLoginID struct {
	UserID  string
	LoginID loginid.LoginID
}

// FIXME(webapp): implement input interface
type EnterLoginIDUpdateLoginID struct {
	UserID string
	Old    loginid.LoginID
	New    loginid.LoginID
}

// FIXME(webapp): implement input interface
type EnterLoginIDAddLoginID struct {
	UserID  string
	LoginID loginid.LoginID
}

func (h *EnterLoginIDHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	userID := auth.GetSession(r.Context()).AuthnAttrs().UserID

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

		h.Renderer.Render(w, r, TemplateItemTypeAuthUIEnterLoginIDHTML, data)
		return
	}

	if r.Method == "POST" && r.Form.Get("x_action") == "remove" {
		h.Database.WithTx(func() error {
			_, _, _, err := h.WebApp.Get(StateID(r))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return err
			}

			// FIXME(webapp): derive EnterLoginIDViewModel with graph and edges
			enterLoginIDViewModel := EnterLoginIDViewModel{}

			result, err := h.WebApp.PostInput(StateID(r), func() (input interface{}, err error) {
				input = &EnterLoginIDRemoveLoginID{
					UserID: userID,
					LoginID: loginid.LoginID{
						Key:   enterLoginIDViewModel.LoginIDKey,
						Value: enterLoginIDViewModel.OldLoginIDValue,
					},
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

	if r.Method == "POST" && r.Form.Get("x_action") == "add_or_update" {
		h.Database.WithTx(func() error {
			_, _, _, err := h.WebApp.Get(StateID(r))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return err
			}

			// FIXME(webapp): derive EnterLoginIDViewModel with graph and edges
			enterLoginIDViewModel := EnterLoginIDViewModel{}

			newLoginID := r.Form.Get("x_login_id")

			result, err := h.WebApp.PostInput(StateID(r), func() (input interface{}, err error) {
				if enterLoginIDViewModel.OldLoginIDValue != "" {
					input = &EnterLoginIDUpdateLoginID{
						UserID: userID,
						Old: loginid.LoginID{
							Key:   enterLoginIDViewModel.LoginIDKey,
							Value: enterLoginIDViewModel.OldLoginIDValue,
						},
						New: loginid.LoginID{
							Key:   enterLoginIDViewModel.LoginIDKey,
							Value: newLoginID,
						},
					}
				} else {
					input = &EnterLoginIDAddLoginID{
						UserID: userID,
						LoginID: loginid.LoginID{
							Key:   enterLoginIDViewModel.LoginIDKey,
							Value: newLoginID,
						},
					}
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
