package webapp

import (
	"fmt"
	"mime"
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/template"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/interaction/nodes"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

const (
	TemplateItemTypeAuthUISetupRecoveryCodeHTML   config.TemplateItemType = "auth_ui_setup_recovery_code.html"
	TemplateItemTypeAuthUIDownloadRecoveryCodeTXT config.TemplateItemType = "auth_ui_download_recovery_code.txt"
)

var TemplateAuthUISetupRecoveryCodeHTML = template.Spec{
	Type:        TemplateItemTypeAuthUISetupRecoveryCodeHTML,
	IsHTML:      true,
	Translation: TemplateItemTypeAuthUITranslationJSON,
	Defines:     defines,
	Components:  components,
}

var TemplateAuthUIDownloadRecoveryCodeTXT = template.Spec{
	Type: TemplateItemTypeAuthUIDownloadRecoveryCodeTXT,
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

func (h *SetupRecoveryCodeHandler) MakeViewModel(graph *interaction.Graph) SetupRecoveryCodeViewModel {
	var node SetupRecoveryCodeNode
	if !graph.FindLastNode(&node) {
		panic(fmt.Errorf("setup_recovery_code: expected graph has node implementing SetupRecoveryCodeNode"))
	}

	recoveryCodes := node.GetRecoveryCodes()
	recoveryCodes = formatRecoveryCodes(recoveryCodes)

	return SetupRecoveryCodeViewModel{
		RecoveryCodes: recoveryCodes,
	}
}

func (h *SetupRecoveryCodeHandler) GetData(r *http.Request, state *webapp.State, graph *interaction.Graph) (map[string]interface{}, error) {
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
		download := r.Form.Get("download") == "true"
		if download {
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

				h.Renderer.Render(w, r, TemplateItemTypeAuthUIDownloadRecoveryCodeTXT, data, setRecoveryCodeAttachmentHeaders)
				return nil
			})
		} else {
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

				h.Renderer.RenderHTML(w, r, TemplateItemTypeAuthUISetupRecoveryCodeHTML, data)
				return nil
			})
		}
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

func setRecoveryCodeAttachmentHeaders(w http.ResponseWriter) {
	// No need to use FormatMediaType because the value is constant.
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{
		"filename": "recovery-codes.txt",
	}))
}
