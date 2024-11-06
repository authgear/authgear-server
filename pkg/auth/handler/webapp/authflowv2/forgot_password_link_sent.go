package authflowv2

import (
	"context"
	"net/http"

	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow/declarative"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebAuthflowV2ForgotPasswordLinkSentHTML = template.RegisterHTML(
	"web/authflowv2/forgot_password_link_sent.html",
	handlerwebapp.Components...,
)

func ConfigureAuthflowV2ForgotPasswordLinkSentRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "GET", "POST").
		WithPathPattern(AuthflowV2RouteForgotPasswordLinkSent)
}

type AuthflowV2ForgotPasswordLinkSentViewModel struct {
	MaskedDisplayName string
	ResendCooldown    int
}

type AuthflowV2ForgotPasswordLinkSentHandler struct {
	Controller    *handlerwebapp.AuthflowController
	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      handlerwebapp.Renderer
	Clock         clock.Clock
}

func (h *AuthflowV2ForgotPasswordLinkSentHandler) GetData(w http.ResponseWriter, r *http.Request, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	now := h.Clock.NowUTC()

	screenData, ok := screen.StateTokenFlowResponse.Action.
		Data.(declarative.IntentAccountRecoveryFlowStepVerifyAccountRecoveryCodeData)
	if !ok {
		panic("unexpected data type in forgot password link sent screen")
	}

	baseViewModel := h.BaseViewModel.ViewModelForAuthFlow(r, w)
	viewmodels.Embed(data, baseViewModel)

	linkSendViewModel := &AuthflowV2ForgotPasswordLinkSentViewModel{
		MaskedDisplayName: screenData.MaskedDisplayName,
		ResendCooldown:    int(screenData.CanResendAt.Sub(now).Seconds()),
	}
	viewmodels.Embed(data, linkSendViewModel)

	return data, nil
}

func (h *AuthflowV2ForgotPasswordLinkSentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var handlers handlerwebapp.AuthflowControllerHandlers
	handlers.Get(func(ctx context.Context, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {

		data, err := h.GetData(w, r, s, screen)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebAuthflowV2ForgotPasswordLinkSentHTML, data)
		return nil
	})

	handlers.PostAction("resend", func(ctx context.Context, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		result, err := h.Controller.AdvanceWithInput(ctx, r, s, screen, map[string]interface{}{
			"resend": true,
		}, nil)
		if err != nil {
			return err
		}

		result.WriteResponse(w, r)
		return nil
	})

	h.Controller.HandleStep(r.Context(), w, r, &handlers)
}
