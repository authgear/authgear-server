package webapp

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator/whatsapp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebWhatsappHTML = template.RegisterHTML(
	"web/whatsapp_otp.html",
	components...,
)

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

type WhatsappOTPHandler struct {
	ControllerFactory    ControllerFactory
	BaseViewModel        *viewmodels.BaseViewModeler
	Renderer             Renderer
	WhatsappCodeProvider WhatsappCodeProvider
}

func (h *WhatsappOTPHandler) GetData(r *http.Request, rw http.ResponseWriter, session *webapp.Session, graph *interaction.Graph) (map[string]interface{}, error) {
	data := map[string]interface{}{}
	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	whatsappViewModel := WhatsappOTPViewModel{
		MethodQuery: getMethodFromQuery(r),
		StateQuery:  getStateFromQuery(r),
	}
	if err := whatsappViewModel.AddData(r, graph); err != nil {
		return nil, err
	}
	viewmodels.Embed(data, baseViewModel)
	viewmodels.Embed(data, whatsappViewModel)
	return data, nil
}

func (h *WhatsappOTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.Serve()

	getPhoneFromGraph := func() (string, error) {
		graph, err := ctrl.InteractionGet()
		if err != nil {
			return "", err
		}
		var phone string
		var n WhatsappOTPNode
		if graph.FindLastNode(&n) {
			phone = n.GetPhone()
		} else {
			panic(fmt.Errorf("webapp: unexpected node for sms fallback: %T", n))
		}

		return phone, nil
	}

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

	ctrl.PostAction("dryrun_verify", func() error {
		webSession := webapp.GetSession(r.Context())
		phone, err := getPhoneFromGraph()
		if err != nil {
			return err
		}

		method := getMethodFromQuery(r)
		var state WhatsappOTPPageQueryState
		_, err = h.WhatsappCodeProvider.VerifyCode(phone, webSession.ID, false)
		if err == nil {
			state = WhatsappOTPPageQueryStateMatched
		} else if errors.Is(err, whatsapp.ErrInvalidCode) {
			state = WhatsappOTPPageQueryStateInvalidCode
		} else if errors.Is(err, whatsapp.ErrInputRequired) {
			state = WhatsappOTPPageQueryStateNoCode
		} else {
			return err
		}

		q := r.URL.Query()
		q.Set(WhatsappOTPPageQueryMethodKey, string(method))
		q.Set(WhatsappOTPPageQueryStateKey, string(state))

		u := url.URL{}
		u.Path = r.URL.Path
		u.RawQuery = q.Encode()
		result := webapp.Result{
			RedirectURI:      u.String(),
			NavigationAction: "replace",
		}
		result.WriteResponse(w, r)
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
		graph, err := ctrl.InteractionGet()
		if err != nil {
			return err
		}

		var phone string
		var n WhatsappOTPNode
		if graph.FindLastNode(&n) {
			phone = n.GetPhone()
		} else {
			panic(fmt.Errorf("webapp: unexpected node for sms fallback: %T", n))
		}

		result, err := ctrl.InteractionPost(func() (input interface{}, err error) {
			input = &InputSetupWhatsappFallbackSMS{
				InputSetupOOB{
					InputType: "phone",
					Target:    phone,
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
