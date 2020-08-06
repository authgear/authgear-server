package webapp

import (
	"fmt"
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction/nodes"
	"github.com/authgear/authgear-server/pkg/auth/dependency/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/db"
	"github.com/authgear/authgear-server/pkg/httproute"
	"github.com/authgear/authgear-server/pkg/template"
)

const (
	TemplateItemTypeAuthUISetupRecoveryCodeHTML config.TemplateItemType = "auth_ui_setup_recovery_code.html"
)

var TemplateAuthUISetupRecoveryCodeHTML = template.Spec{
	Type:        TemplateItemTypeAuthUISetupRecoveryCodeHTML,
	IsHTML:      true,
	Translation: TemplateItemTypeAuthUITranslationJSON,
	Defines:     defines,
	Components:  components,
}

func ConfigureSetupRecoveryCodeRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/setup_recovery_code")
}

type SetupRecoveryCodeViewModel struct {
	RecoveryCodes []string
}

type SetupRecoveryCodeNode interface {
	GetRecoveryCodes() []string
}

type SetupRecoveryCodeInput struct{}

var _ nodes.InputGenerateRecoveryCodeEnd = &SetupRecoveryCodeInput{}

// ViewedSetupRecoveryCodes implements InputGenerateRecoveryCodeEnd.
func (i *SetupRecoveryCodeInput) ViewedRecoveryCodes() {}

type SetupRecoveryCodeHandler struct {
	Database      *db.Handle
	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      Renderer
	WebApp        WebAppService
}

func (h *SetupRecoveryCodeHandler) MakeViewModel(graph *newinteraction.Graph) SetupRecoveryCodeViewModel {
	node, ok := graph.CurrentNode().(SetupRecoveryCodeNode)
	if !ok {
		panic(fmt.Errorf("setup_recovery_code: expected current node to implement SetupRecoveryCodeNode: %T", graph.CurrentNode()))
	}

	recoveryCodes := node.GetRecoveryCodes()
	recoveryCodes = formatRecoveryCodes(recoveryCodes)

	return SetupRecoveryCodeViewModel{
		RecoveryCodes: recoveryCodes,
	}
}

func (h *SetupRecoveryCodeHandler) GetData(r *http.Request, state *webapp.State, graph *newinteraction.Graph, edges []newinteraction.Edge) (map[string]interface{}, error) {
	data := map[string]interface{}{}

	var anyError interface{}
	if state != nil {
		anyError = state.Error
	}

	baseViewModel := h.BaseViewModel.ViewModel(r, anyError)
	viewModel := h.MakeViewModel(graph)

	viewmodels.Embed(data, baseViewModel)
	viewmodels.Embed(data, viewModel)
	return data, nil
}

func (h *SetupRecoveryCodeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

			h.Renderer.Render(w, r, TemplateItemTypeAuthUISetupRecoveryCodeHTML, data)
			return nil
		})
	}

	if r.Method == "POST" {
		h.Database.WithTx(func() error {
			result, err := h.WebApp.PostInput(StateID(r), func() (input interface{}, err error) {
				input = &SetupRecoveryCodeInput{}
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

func formatRecoveryCodes(recoveryCodes []string) []string {
	out := make([]string, len(recoveryCodes))
	for i, code := range recoveryCodes {
		halfLength := len(code) / 2
		formattedCode := fmt.Sprintf("%s %s", code[:halfLength], code[halfLength:])
		out[i] = formattedCode
	}
	return out
}
