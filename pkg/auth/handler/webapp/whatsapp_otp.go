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
		WithPathPattern("/flows/whatsapp_otp")
}

type WhatsappOTPNode interface {
	GetPhoneOTPMode() config.AuthenticatorPhoneOTPMode
	GetWhatsappOTP() string
	GetPhone() string
}

type WhatsappOTPAuthnNode interface {
	GetAuthenticatorIndex() int
}

type WhatsappOTPHandler struct {
	ControllerFactory         ControllerFactory
	BaseViewModel             *viewmodels.BaseViewModeler
	AlternativeStepsViewModel *viewmodels.AlternativeStepsViewModeler
	Renderer                  Renderer
	WhatsappCodeProvider      WhatsappCodeProvider
}

func (h *WhatsappOTPHandler) GetData(r *http.Request, rw http.ResponseWriter, session *webapp.Session, graph *interaction.Graph) (map[string]interface{}, error) {
	data := map[string]interface{}{}
	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	whatsappViewModel := WhatsappOTPViewModel{
		StateQuery: getStateFromQuery(r),
	}
	if err := whatsappViewModel.AddData(r, graph, h.WhatsappCodeProvider, baseViewModel.Translations); err != nil {
		return nil, err
	}
	currentStepKind := session.CurrentStep().Kind
	phoneOTPAlternatives := viewmodels.PhoneOTPAlternativeStepsViewModel{}
	if err := phoneOTPAlternatives.AddAlternatives(graph, currentStepKind); err != nil {
		return nil, err
	}
	// alternatives
	var alternatives *viewmodels.AlternativeStepsViewModel
	var node1 CreateAuthenticatorBeginNode
	var node2 AuthenticationBeginNode
	nodesInf := []interface{}{
		&node1,
		&node2,
	}
	node := graph.FindLastNodeFromList(nodesInf)
	switch node.(type) {
	case *CreateAuthenticatorBeginNode:
		var err error
		alternatives, err = h.AlternativeStepsViewModel.CreateAuthenticatorAlternatives(graph, currentStepKind)
		if err != nil {
			return nil, err
		}
	case *AuthenticationBeginNode:
		var err error
		alternatives, err = h.AlternativeStepsViewModel.AuthenticationAlternatives(graph, currentStepKind)
		if err != nil {
			return nil, err
		}
	default:
		// identity verification
		// alternatives are provided in PhoneOTPAlternativeStepsViewModel
		alternatives = &viewmodels.AlternativeStepsViewModel{}
	}

	viewmodels.Embed(data, baseViewModel)
	viewmodels.Embed(data, whatsappViewModel)
	viewmodels.Embed(data, phoneOTPAlternatives)
	viewmodels.Embed(data, *alternatives)
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

		deviceToken := r.Form.Get("x_device_token") == "true"
		q := r.URL.Query()
		q.Set(WhatsappOTPPageQueryStateKey, string(state))
		if deviceToken {
			q.Set(WhatsappOTPPageQueryXDeviceTokenKey, "true")
		} else {
			q.Del(WhatsappOTPPageQueryXDeviceTokenKey)
		}

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
		deviceToken := r.Form.Get("x_device_token") == "true"
		result, err := ctrl.InteractionPost(func() (input interface{}, err error) {
			input = &InputVerifyWhatsappOTP{
				DeviceToken: deviceToken,
			}
			return
		})
		if err != nil {
			return err
		}

		result.WriteResponse(w, r)
		return nil
	})

	handleAlternativeSteps(ctrl)
}
