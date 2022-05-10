package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebVerifyIdentityViaWhatsappHTML = template.RegisterHTML(
	"web/verify_identity_via_whatsapp.html",
	components...,
)

func ConfigureVerifyIdentityViaWhatsappRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/verify_identity_via_whatsapp")
}

type VerifyIdentityViaWhatsappViewModel struct {
	PhoneOTPMode config.AuthenticatorPhoneOTPMode
	WhatsappOTP  string
}

type VerifyIdentityViaWhatsappHandler struct {
	ControllerFactory ControllerFactory
	BaseViewModel     *viewmodels.BaseViewModeler
	Renderer          Renderer
}

func (h *VerifyIdentityViaWhatsappHandler) GetData(r *http.Request, rw http.ResponseWriter, graph *interaction.Graph) (map[string]interface{}, error) {
	data := map[string]interface{}{}
	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	viewModel := VerifyIdentityViaWhatsappViewModel{}
	var n WhatsappOTPNode
	if graph.FindLastNode(&n) {
		viewModel.PhoneOTPMode = n.GetPhoneOTPMode()
		viewModel.WhatsappOTP = n.GetWhatsappOTP()
	}
	viewmodels.Embed(data, baseViewModel)
	viewmodels.Embed(data, viewModel)
	return data, nil
}

func (h *VerifyIdentityViaWhatsappHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.Serve()

	ctrl.Get(func() error {
		graph, err := ctrl.InteractionGet()
		if err != nil {
			return err
		}

		data, err := h.GetData(r, w, graph)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebVerifyIdentityViaWhatsappHTML, data)
		return nil
	})

	ctrl.PostAction("verify", func() error {
		result, err := ctrl.InteractionPost(func() (input interface{}, err error) {
			input = &InputVerifyWhatsappOTP{}
			return
		})
		if err != nil {
			return err
		}

		result.WriteResponse(w, r)
		return nil
	})

	ctrl.PostAction("fallback_sms", func() error {
		result, err := ctrl.InteractionPost(func() (input interface{}, err error) {
			input = &InputVerifyIdentityViaWhatsappFallbackSMS{}
			return
		})
		if err != nil {
			return err
		}

		result.WriteResponse(w, r)
		return nil
	})

}
