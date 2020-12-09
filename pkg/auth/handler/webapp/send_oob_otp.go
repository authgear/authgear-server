package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebSendOOBOTPHTML = template.RegisterHTML(
	"web/send_oob_otp.html",
	components...,
)

type SendOOBOTPViewModel struct {
	AlternativeSteps []viewmodels.AlternativeStep
	OOBOTPTarget     string
	OOBOTPCodeLength int
	OOBOTPChannel    string
}

func ConfigureSendOOBOTPRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/send_oob_otp")
}

type SendOOBOTPHandler struct {
	ControllerFactory ControllerFactory
	BaseViewModel     *viewmodels.BaseViewModeler
	Renderer          Renderer
}

type TriggerOOBOTPEdge interface {
	GetOOBOTPTarget(idx int) string
	GetOOBOTPChannel(idx int) string
}

func (h *SendOOBOTPHandler) GetData(r *http.Request, rw http.ResponseWriter, graph *interaction.Graph) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	baseViewModel := h.BaseViewModel.ViewModel(r, rw)

	// TODO: obtain oob_otp authenticator information for display

	viewmodels.EmbedForm(data, r.Form)
	viewmodels.Embed(data, baseViewModel)
	return data, nil
}

func (h *SendOOBOTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

		h.Renderer.RenderHTML(w, r, TemplateWebSendOOBOTPHTML, data)
		return nil
	})

	ctrl.PostAction("send", func() error {
		result, err := ctrl.InteractionPost(func() (input interface{}, err error) {
			input = &InputTriggerOOB{AuthenticatorIndex: 0}
			return
		})
		if err != nil {
			return err
		}

		result.WriteResponse(w, r)
		return nil
	})
}
