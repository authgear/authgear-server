package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/auth/dependency/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/db"
	"github.com/authgear/authgear-server/pkg/httproute"
	"github.com/authgear/authgear-server/pkg/template"
	"github.com/authgear/authgear-server/pkg/validation"
)

const (
	TemplateItemTypeAuthUIVerifyUserHTML config.TemplateItemType = "auth_ui_verify_user.html"
)

var TemplateAuthUIVerifyUserHTML = template.Spec{
	Type:        TemplateItemTypeAuthUIVerifyUserHTML,
	IsHTML:      true,
	Translation: TemplateItemTypeAuthUITranslationJSON,
	Defines:     defines,
	Components:  components,
}

const VerifyUserRequestSchema = "VerifyUserRequestSchema"

var VerifyUserSchema = validation.NewMultipartSchema("").
	Add(VerifyUserRequestSchema, `
		{
			"type": "object",
			"properties": {
				"x_password": { "type": "string" }
			},
			"required": ["x_password"]
		}
	`).Instantiate()

func ConfigureVerifyUserRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/verify_user")
}

type VerifyUserViewModel struct {
	VerificationCodeSendCooldown int
	VerificationCodeLength       int
	VerificationCodeChannel      string
	IdentityDisplayID            string
}

type VerifyUserHandler struct {
	Database      *db.Handle
	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      Renderer
	WebApp        WebAppService
}

type VerifyIdentityNode interface {
	GetVerificationIdentity() *identity.Info
	GetVerificationCodeChannel() string
	GetVerificationCodeSendCooldown() int
	GetVerificationCodeLength() int
}

func (h *VerifyUserHandler) GetData(r *http.Request, state *webapp.State, graph *newinteraction.Graph, edges []newinteraction.Edge) (map[string]interface{}, error) {
	data := map[string]interface{}{}

	baseViewModel := h.BaseViewModel.ViewModel(r, state.Error)
	viewModel := VerifyUserViewModel{}
	if n, ok := graph.CurrentNode().(VerifyIdentityNode); ok {
		viewModel.IdentityDisplayID = n.GetVerificationIdentity().DisplayID()
		viewModel.VerificationCodeSendCooldown = n.GetVerificationCodeSendCooldown()
		viewModel.VerificationCodeLength = n.GetVerificationCodeLength()
		viewModel.VerificationCodeChannel = n.GetVerificationCodeChannel()
	}

	viewmodels.Embed(data, baseViewModel)
	viewmodels.Embed(data, viewModel)

	return data, nil
}

type VerificationCodeResend struct{}

func (i *VerificationCodeResend) DoResend() {}

type VerificationCodeInput struct {
	Code string
}

// GetVerificationCode implements InputVerifyIdentityCheckCode.
func (i *VerificationCodeInput) GetVerificationCode() string {
	return i.Code
}

func (h *VerifyUserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

			h.Renderer.Render(w, r, TemplateItemTypeAuthUIVerifyUserHTML, data)
			return nil
		})
	}

	trigger := r.Form.Get("trigger") == "true"

	if r.Method == "POST" && trigger {
		h.Database.WithTx(func() error {
			result, err := h.WebApp.PostInput(StateID(r), func() (input interface{}, err error) {
				input = &VerificationCodeResend{}
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
				err = VerifyUserSchema.PartValidator(VerifyUserRequestSchema).ValidateValue(FormToJSON(r.Form))
				if err != nil {
					return
				}

				code := r.Form.Get("x_password")

				input = &VerificationCodeInput{
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
