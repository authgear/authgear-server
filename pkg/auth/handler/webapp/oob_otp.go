package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction/nodes"
	"github.com/authgear/authgear-server/pkg/auth/dependency/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/db"
	"github.com/authgear/authgear-server/pkg/httproute"
	"github.com/authgear/authgear-server/pkg/template"
	"github.com/authgear/authgear-server/pkg/validation"
)

const (
	TemplateItemTypeAuthUIOOBOTPHTML config.TemplateItemType = "auth_ui_oob_otp_html"
)

var TemplateAuthUIOOBOTPHTML = template.Spec{
	Type:        TemplateItemTypeAuthUIOOBOTPHTML,
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

{{ if $.OOBOTPChannel }}
{{ if eq $.OOBOTPChannel "sms" }}
<div class="title primary-txt">{{ localize "oob-otp-page-title--sms" }}</div>
{{ end }}
{{ if eq $.OOBOTPChannel "email" }}
<div class="title primary-txt">{{ localize "oob-otp-page-title--email" }}</div>
{{ end }}
{{ end }}

{{ template "ERROR" . }}

{{ if $.IdentityDisplayID }}
<div class="description primary-txt">{{ localize "oob-otp-description" $.OOBOTPCodeLength $.IdentityDisplayID }}</div>
{{ end }}

<form class="vertical-form form-fields-container" method="post" novalidate>
{{ $.CSRFField }}

<input class="input text-input primary-txt" type="text" inputmode="numeric" pattern="[0-9]*" name="x_password" placeholder="{{ localize "oob-otp-placeholder" }}">
<button class="btn primary-btn align-self-flex-end" type="submit" name="submit" value="">{{ localize "next-button-label" }}</button>
</form>

<form class="link oob-otp-trigger-form" method="post" novalidate>
{{ $.CSRFField }}

<span class="primary-txt">{{ localize "oob-otp-resend-button-hint" }}</span>
<button id="resend-button" class="anchor" type="submit" name="trigger" value="true"
	data-cooldown="{{ $.OOBOTPCodeSendCooldown }}"
	data-label="{{ localize "oob-otp-resend-button-label" }}"
	data-label-unit="{{ localize "oob-otp-resend-button-label--unit" }}">{{ localize "oob-otp-resend-button-label" }}</button>
</form>

</div>
{{ template "auth_ui_footer.html" . }}

</div>
</body>
</html>
`,
}

const OOBOTPRequestSchema = "OOBOTPRequestSchema"

var OOBOTPSchema = validation.NewMultipartSchema("").
	Add(OOBOTPRequestSchema, `
		{
			"type": "object",
			"properties": {
				"x_password": { "type": "string" }
			},
			"required": ["x_password"]
		}
	`).Instantiate()

func ConfigureOOBOTPRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/oob_otp")
}

type OOBOTPViewModel struct {
	OOBOTPCodeSendCooldown int
	OOBOTPCodeLength       int
	OOBOTPChannel          string
	IdentityDisplayID      string
}

type OOBOTPHandler struct {
	Database      *db.Handle
	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      Renderer
	WebApp        WebAppService
}

func (h *OOBOTPHandler) GetData(r *http.Request, state *webapp.State, graph *newinteraction.Graph, edges []newinteraction.Edge) (map[string]interface{}, error) {
	data := map[string]interface{}{}

	baseViewModel := h.BaseViewModel.ViewModel(r, state.Error)
	node := graph.CurrentNode()
	// FIXME(webapp): Fill in OOBOTPCodeSendCooldown, OOBOTPCodeLength from node.
	oobOTPViewModel := OOBOTPViewModel{
		OOBOTPCodeSendCooldown: 60,
		OOBOTPCodeLength:       4,
		IdentityDisplayID:      graph.MustGetUserLastIdentity().DisplayID(),
	}
	if n, ok := node.(*nodes.NodeAuthenticationOOBTrigger); ok {
		oobOTPViewModel.OOBOTPChannel = n.Authenticator.Props[authenticator.AuthenticatorPropOOBOTPChannelType].(string)
	}

	viewmodels.Embed(data, baseViewModel)
	viewmodels.Embed(data, oobOTPViewModel)

	return data, nil
}

type OOBOTPResend struct{}

func (i *OOBOTPResend) DoResend() {}

type OOBOTPInput struct {
	Code string
}

// GetOOBOTP implements InputAuthenticationOOB.
func (i *OOBOTPInput) GetOOBOTP() string {
	return i.Code
}

func (h *OOBOTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Method == "GET" {
		h.Database.WithTx(func() error {
			state, graph, edges, err := h.WebApp.Get(StateID(r))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return err
			}

			data, err := h.GetData(r, state, graph, edges)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return err
			}

			h.Renderer.Render(w, r, TemplateItemTypeAuthUIOOBOTPHTML, data)
			return nil
		})
	}

	trigger := r.Form.Get("trigger") == "true"

	if r.Method == "POST" && trigger {
		h.Database.WithTx(func() error {
			result, err := h.WebApp.PostInput(StateID(r), func() (input interface{}, err error) {
				input = &OOBOTPResend{}
				return
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return err
			}
			result.WriteResponse(w, r)
			return nil
		})
		return
	}

	if r.Method == "POST" {
		h.Database.WithTx(func() error {
			result, err := h.WebApp.PostInput(StateID(r), func() (input interface{}, err error) {
				err = OOBOTPSchema.PartValidator(OOBOTPRequestSchema).ValidateValue(FormToJSON(r.Form))
				if err != nil {
					return
				}

				code := r.Form.Get("x_password")

				input = &OOBOTPInput{
					Code: code,
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
