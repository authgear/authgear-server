package webapp

import (
	"fmt"
	"mime"
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/lib/authn/mfa"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/interaction/nodes"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebSetupRecoveryCodeHTML = template.RegisterHTML(
	"web/setup_recovery_code.html",
	components...,
)

var TemplateWebDownloadRecoveryCodeTXT = template.RegisterPlainText(
	"web/download_recovery_code.txt",
)

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

func (h *SetupRecoveryCodeHandler) GetData(r *http.Request, rw http.ResponseWriter, graph *interaction.Graph) (map[string]interface{}, error) {
	data := map[string]interface{}{}
	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	viewModel := h.MakeViewModel(graph)

	viewmodels.Embed(data, baseViewModel)
	viewmodels.Embed(data, viewModel)
	return data, nil
}

func (h *SetupRecoveryCodeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if r.Method == "GET" {
		download := r.Form.Get("download") == "true"
		if download {
			err := h.Database.WithTx(func() error {
				_, graph, err := h.WebApp.Get(StateID(r))
				if err != nil {
					return err
				}

				data, err := h.GetData(r, w, graph)
				if err != nil {
					return err
				}

				h.Renderer.Render(w, r, TemplateWebDownloadRecoveryCodeTXT, data, setRecoveryCodeAttachmentHeaders)
				return nil
			})
			if err != nil {
				panic(err)
			}
		} else {
			err := h.Database.WithTx(func() error {
				_, graph, err := h.WebApp.Get(StateID(r))
				if err != nil {
					return err
				}

				data, err := h.GetData(r, w, graph)
				if err != nil {
					return err
				}

				h.Renderer.RenderHTML(w, r, TemplateWebSetupRecoveryCodeHTML, data)
				return nil
			})
			if err != nil {
				panic(err)
			}
		}
	}

	if r.Method == "POST" {
		err := h.Database.WithTx(func() error {
			result, err := h.WebApp.PostInput(StateID(r), func() (input interface{}, err error) {
				input = &SetupRecoveryCodeInput{}
				return
			})
			if err != nil {
				return err
			}
			result.WriteResponse(w, r)
			return nil
		})
		if err != nil {
			panic(err)
		}
	}
}

func formatRecoveryCodes(recoveryCodes []string) []string {
	out := make([]string, len(recoveryCodes))
	for i, code := range recoveryCodes {
		out[i] = mfa.FormatRecoveryCode(code)
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
