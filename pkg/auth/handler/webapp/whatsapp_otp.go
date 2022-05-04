package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebWhatsappHTML = template.RegisterHTML(
	"web/whatsapp_otp.html",
	components...,
)

var FallbackSMSPasswordSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"x_e164": { "type": "string" }
		},
		"required": ["x_e164"]
	}
`)

func ConfigureWhatsappOTPRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/whatsapp_otp")
}

type WhatsappOTPNode interface {
	GetPhoneOTPMode() config.AuthenticatorPhoneOTPMode
	GetWhatsappOTP() string
	GetPhone() string
}

type WhatsappOTPViewModel struct {
	PhoneOTPMode config.AuthenticatorPhoneOTPMode
	WhatsappOTP  string
	Phone        string
}

type WhatsappOTPHandler struct {
	ControllerFactory ControllerFactory
	BaseViewModel     *viewmodels.BaseViewModeler
	Renderer          Renderer
}

func (h *WhatsappOTPHandler) GetData(r *http.Request, rw http.ResponseWriter, session *webapp.Session, graph *interaction.Graph) (map[string]interface{}, error) {
	data := map[string]interface{}{}
	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	viewModel := WhatsappOTPViewModel{}
	var n WhatsappOTPNode
	if graph.FindLastNode(&n) {
		viewModel.PhoneOTPMode = n.GetPhoneOTPMode()
		viewModel.WhatsappOTP = n.GetWhatsappOTP()
		viewModel.Phone = n.GetPhone()
	}
	viewmodels.Embed(data, baseViewModel)
	viewmodels.Embed(data, viewModel)
	return data, nil
}

func (h *WhatsappOTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.Serve()

	ctrl.Get(func() error {
		session, err := ctrl.InteractionSession()
		if err != nil {
			return err
		}

		graph, err := ctrl.InteractionGet()
		if err != nil {
			return err
		}

		data, err := h.GetData(r, w, session, graph)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebWhatsappHTML, data)
		return nil
	})

	ctrl.PostAction("verify", func() error {
		result, err := ctrl.InteractionPost(func() (input interface{}, err error) {
			// fixme(whatsapp): get the otp from redis which store by the webhook
			input = &InputWhatsappOTP{
				WhatsappOTP: "secret",
			}
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
			err = FallbackSMSPasswordSchema.Validator().ValidateValue(FormToJSON(r.Form))
			if err != nil {
				return
			}

			input = &InputSetupWhatsappFallbackSMS{
				InputSetupOOB{
					InputType: "phone",
					Target:    r.Form.Get("x_e164"),
				},
			}
			return
		})
		if err != nil {
			return err
		}

		result.WriteResponse(w, r)
		return nil
	})
}
