package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebMissingWeb3WalletHTML = template.RegisterHTML(
	"web/missing_web3_wallet.html",
	Components...,
)

func ConfigureMissingWeb3WalletRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "GET").
		WithPathPattern("/errors/missing_web3_wallet")
}

type MissingWeb3WalletViewModel struct {
	Provider string
}

type MissingWeb3WalletHandler struct {
	ControllerFactory    ControllerFactory
	BaseViewModel        *viewmodels.BaseViewModeler
	Renderer             Renderer
	AuthenticationConfig *config.AuthenticationConfig
}

func (h *MissingWeb3WalletHandler) GetData(r *http.Request, rw http.ResponseWriter) (map[string]interface{}, error) {
	data := map[string]interface{}{}

	baseViewModel := h.BaseViewModel.ViewModel(r, rw)

	provider := ""
	if p := r.Form.Get("provider"); p == "" {
		provider = "metamask"
	} else {
		provider = p
	}

	missingWeb3WalletViewModel := MissingWeb3WalletViewModel{
		Provider: provider,
	}

	viewmodels.Embed(data, missingWeb3WalletViewModel)
	viewmodels.Embed(data, baseViewModel)

	return data, nil
}

func (h *MissingWeb3WalletHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	identities := h.AuthenticationConfig.Identities
	isSIWEEnabled := false
	for _, i := range identities {
		if i == model.IdentityTypeSIWE {
			isSIWEEnabled = true
			break
		}
	}

	if !isSIWEEnabled {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	defer ctrl.ServeWithDBTx()

	ctrl.Get(func() error {
		data, err := h.GetData(r, w)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebMissingWeb3WalletHTML, data)
		return nil
	})
}
