package webapp

import (
	"fmt"
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/phone"
	"github.com/authgear/authgear-server/pkg/util/template"
)

const DefaultAuthenticatorIndex = 0

type AuthenticationBeginNode interface {
	GetAuthenticationEdges() ([]interaction.Edge, error)
}

var TemplateWebSendOOBOTPHTML = template.RegisterHTML(
	"web/send_oob_otp.html",
	components...,
)

type SendOOBOTPViewModel struct {
	OOBOTPTarget     string
	OOBOTPCodeLength int
	OOBOTPChannel    authn.AuthenticatorOOBChannel
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
	GetOOBOTPChannel(idx int) authn.AuthenticatorOOBChannel
}

func (h *SendOOBOTPHandler) GetData(r *http.Request, rw http.ResponseWriter, graph *interaction.Graph) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	viewModel := SendOOBOTPViewModel{}

	var node AuthenticationBeginNode
	if !graph.FindLastNode(&node) {
		panic("send_oob_otp: expected graph has node implementing AuthenticationBeginNode")
	}

	edges, err := node.GetAuthenticationEdges()
	if err != nil {
		return nil, err
	}

	if edge, ok := edges[0].(TriggerOOBOTPEdge); ok {
		viewModel.OOBOTPChannel = edge.GetOOBOTPChannel(DefaultAuthenticatorIndex)
		switch viewModel.OOBOTPChannel {
		case authn.AuthenticatorOOBChannelEmail:
			viewModel.OOBOTPTarget = mail.MaskAddress(edge.GetOOBOTPTarget(DefaultAuthenticatorIndex))
		case authn.AuthenticatorOOBChannelSMS:
			viewModel.OOBOTPTarget = phone.Mask(edge.GetOOBOTPTarget(DefaultAuthenticatorIndex))
		}
	} else {
		panic(fmt.Errorf("send_oob_otp: unexpected edge: %T", edges[0]))
	}

	alternatives := viewmodels.AlternativeStepsViewModel{}
	err = alternatives.AddAuthenticationAlternatives(graph, webapp.SessionStepEnterOOBOTPAuthn)
	if err != nil {
		return nil, err
	}

	viewmodels.Embed(data, baseViewModel)
	viewmodels.Embed(data, viewModel)
	viewmodels.Embed(data, alternatives)
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
			input = &InputTriggerOOB{AuthenticatorIndex: DefaultAuthenticatorIndex}
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
