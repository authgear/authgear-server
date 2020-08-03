package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
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

type OOBOTPNode interface {
	GetOOBOTPChannel() string
	GetOOBOTPCodeSendCooldown() int
	GetOOBOTPCodeLength() int
}

func (h *OOBOTPHandler) GetData(r *http.Request, state *webapp.State, graph *newinteraction.Graph, edges []newinteraction.Edge) (map[string]interface{}, error) {
	data := map[string]interface{}{}

	baseViewModel := h.BaseViewModel.ViewModel(r, state.Error)
	oobOTPViewModel := OOBOTPViewModel{
		IdentityDisplayID: graph.MustGetUserLastIdentity().DisplayID(),
	}
	if n, ok := graph.CurrentNode().(OOBOTPNode); ok {
		oobOTPViewModel.OOBOTPCodeSendCooldown = n.GetOOBOTPCodeSendCooldown()
		oobOTPViewModel.OOBOTPCodeLength = n.GetOOBOTPCodeLength()
		oobOTPViewModel.OOBOTPChannel = n.GetOOBOTPChannel()
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
