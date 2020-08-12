package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction/nodes"
	"github.com/authgear/authgear-server/pkg/auth/dependency/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/template"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

const (
	TemplateItemTypeAuthUIEnterRecoveryCodeHTML config.TemplateItemType = "auth_ui_enter_recovery_code.html"
)

var TemplateAuthUIEnterRecoveryCodeHTML = template.Spec{
	Type:        TemplateItemTypeAuthUIEnterRecoveryCodeHTML,
	IsHTML:      true,
	Translation: TemplateItemTypeAuthUITranslationJSON,
	Defines:     defines,
	Components:  components,
}

const EnterRecoveryCodeRequestSchema = "EnterRecoveryCodeRequestSchema"

var EnterRecoveryCodeSchema = validation.NewMultipartSchema("").
	Add(EnterRecoveryCodeRequestSchema, `
		{
			"type": "object",
			"properties": {
				"x_code": { "type": "string" }
			},
			"required": ["x_code"]
		}
	`).Instantiate()

func ConfigureEnterRecoveryCodeRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/enter_recovery_code")
}

type EnterRecoveryCodeViewModel struct {
	Alternatives []AuthenticationAlternative
}

type EnterRecoveryCodeHandler struct {
	Database      *db.Handle
	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      Renderer
	WebApp        WebAppService
}

func (h *EnterRecoveryCodeHandler) GetData(r *http.Request, state *webapp.State, graph *newinteraction.Graph) (map[string]interface{}, error) {
	data := map[string]interface{}{}

	baseViewModel := h.BaseViewModel.ViewModel(r, state.Error)
	alternatives := DeriveAuthenticationAlternatives(
		// Use current state ID because the current node should be NodeAuthenticationBegin.
		state.ID,
		graph,
		AuthenticationTypeRecoveryCode,
		"",
	)

	viewModel := EnterRecoveryCodeViewModel{
		Alternatives: alternatives,
	}

	viewmodels.Embed(data, baseViewModel)
	viewmodels.Embed(data, viewModel)

	return data, nil
}

type EnterRecoveryCodeInput struct {
	Code string
}

var _ nodes.InputConsumeRecoveryCode = &EnterRecoveryCodeInput{}

// GetRecoveryCode implements InputConsumeRecoveryCode.
func (i *EnterRecoveryCodeInput) GetRecoveryCode() string {
	return i.Code
}

func (h *EnterRecoveryCodeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Method == "GET" {
		h.Database.WithTx(func() error {
			state, graph, err := h.WebApp.Get(StateID(r))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return err
			}

			data, err := h.GetData(r, state, graph)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return err
			}

			h.Renderer.RenderHTML(w, r, TemplateItemTypeAuthUIEnterRecoveryCodeHTML, data)
			return nil
		})
	}

	if r.Method == "POST" {
		h.Database.WithTx(func() error {
			result, err := h.WebApp.PostInput(StateID(r), func() (input interface{}, err error) {
				err = EnterRecoveryCodeSchema.PartValidator(EnterRecoveryCodeRequestSchema).ValidateValue(FormToJSON(r.Form))
				if err != nil {
					return
				}

				code := r.Form.Get("x_code")

				input = &EnterRecoveryCodeInput{
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
