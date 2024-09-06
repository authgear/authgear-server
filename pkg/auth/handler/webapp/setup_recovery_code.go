package webapp

import (
	"fmt"
	"mime"
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/secretcode"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebSetupRecoveryCodeHTML = template.RegisterHTML(
	"web/setup_recovery_code.html",
	Components...,
)

var TemplateWebDownloadRecoveryCodeTXT = template.RegisterPlainText(
	"web/download_recovery_code.txt",
	plainTextComponents...,
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

type SetupRecoveryCodeHandler struct {
	ControllerFactory ControllerFactory
	BaseViewModel     *viewmodels.BaseViewModeler
	Renderer          Renderer
}

func (h *SetupRecoveryCodeHandler) MakeViewModel(graph *interaction.Graph) SetupRecoveryCodeViewModel {
	var node SetupRecoveryCodeNode
	if !graph.FindLastNode(&node) {
		panic(fmt.Errorf("setup_recovery_code: expected graph has node implementing SetupRecoveryCodeNode"))
	}

	recoveryCodes := node.GetRecoveryCodes()
	recoveryCodes = FormatRecoveryCodes(recoveryCodes)

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
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.ServeWithDBTx()

	ctrl.Get(func() error {
		graph, err := ctrl.InteractionGet()
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

	ctrl.PostAction("download", func() error {
		graph, err := ctrl.InteractionGet()
		if err != nil {
			return err
		}

		data, err := h.GetData(r, w, graph)
		if err != nil {
			return err
		}

		SetRecoveryCodeAttachmentHeaders(w)
		h.Renderer.Render(w, r, TemplateWebDownloadRecoveryCodeTXT, data)
		return nil
	})

	ctrl.PostAction("proceed", func() error {
		result, err := ctrl.InteractionPost(func() (input interface{}, err error) {
			input = &InputSetupRecoveryCode{}
			return
		})
		if err != nil {
			return err
		}

		result.WriteResponse(w, r)
		return nil
	})
}

func FormatRecoveryCodes(recoveryCodes []string) []string {
	out := make([]string, len(recoveryCodes))
	for i, code := range recoveryCodes {
		out[i] = secretcode.RecoveryCode.FormatForHuman(code)
	}
	return out
}

func SetRecoveryCodeAttachmentHeaders(w http.ResponseWriter) {
	// No need to use FormatMediaType because the value is constant.
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{
		"filename": "recovery-codes.txt",
	}))
}
