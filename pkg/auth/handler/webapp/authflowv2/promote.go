package authflowv2

import (
	"context"
	"net/http"

	"net/url"

	"github.com/authgear/authgear-server/pkg/api/model"
	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var AuthflowPromoteLoginIDSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"q_login_id_key": { "type": "string" },
			"q_login_id": { "type": "string" }
		},
		"required": ["q_login_id_key", "q_login_id"]
	}
`)

func ConfigureAuthflowV2PromoteRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(AuthflowV2RoutePromote)
}

type AuthflowV2PromoteEndpointsProvider interface {
	SSOCallbackURL(alias string) *url.URL
	SharedSSOCallbackURL() *url.URL
}

type AuthflowV2PromoteHandler struct {
	Controller        *handlerwebapp.AuthflowController
	BaseViewModel     *viewmodels.BaseViewModeler
	AuthflowViewModel *viewmodels.AuthflowViewModeler
	Renderer          handlerwebapp.Renderer
}

func (h *AuthflowV2PromoteHandler) getAuthflowViewModel(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse, r *http.Request) viewmodels.AuthflowViewModel {
	return h.AuthflowViewModel.NewWithAuthflow(s, screen.StateTokenFlowResponse, r)
}

func (h *AuthflowV2PromoteHandler) GetData(
	w http.ResponseWriter, r *http.Request,
	s *webapp.Session,
	screen *webapp.AuthflowScreenWithFlowResponse,
) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	baseViewModel := h.BaseViewModel.ViewModelForAuthFlow(r, w)
	viewmodels.Embed(data, baseViewModel)

	authflowViewModel := h.getAuthflowViewModel(s, screen, r)
	viewmodels.Embed(data, authflowViewModel)

	signupViewModel := AuthflowV2SignupViewModel{
		CanSwitchToLogin: false,
		UIVariant:        AuthflowV2SignupUIVariantSignup,
	}
	viewmodels.Embed(data, signupViewModel)

	return data, nil
}

func (h *AuthflowV2PromoteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	opts := webapp.SessionOptions{
		RedirectURI: h.Controller.RedirectURI(r),
	}

	var handlers handlerwebapp.AuthflowControllerHandlers
	handlers.Get(func(ctx context.Context, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		data, err := h.GetData(w, r, s, screen)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebAuthflowV2SignupHTML, data)
		return nil
	})

	handlers.PostAction("oauth", func(ctx context.Context, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		providerAlias := r.Form.Get("x_provider_alias")
		authflowViewModel := h.getAuthflowViewModel(s, screen, r)
		result, err := h.Controller.UseOAuthIdentification(ctx, s, w, r, screen.Screen, providerAlias, authflowViewModel.IdentificationOptions, func(input map[string]interface{}) (result *webapp.Result, err error) {
			err = handlerwebapp.HandleIdentificationBotProtection(ctx, model.AuthenticationFlowIdentificationOAuth, screen.StateTokenFlowResponse, r.Form, input)
			if err != nil {
				return nil, err
			}

			return h.Controller.AdvanceWithInput(ctx, r, s, screen, input, nil)
		})
		if err != nil {
			return err
		}

		result.WriteResponse(w, r)
		return nil
	})

	handlers.PostAction("login_id", func(ctx context.Context, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		err := AuthflowPromoteLoginIDSchema.Validator().ValidateValue(ctx, handlerwebapp.FormToJSON(r.Form))
		if err != nil {
			return err
		}

		loginIDKey := r.Form.Get("q_login_id_key")
		loginID := r.Form.Get("q_login_id")
		identification := loginIDKey
		input := map[string]interface{}{
			"identification": identification,
			"login_id":       loginID,
		}

		err = handlerwebapp.HandleIdentificationBotProtection(ctx, model.AuthenticationFlowIdentification(identification), screen.StateTokenFlowResponse, r.Form, input)
		if err != nil {
			return err
		}

		result, err := h.Controller.AdvanceWithInput(ctx, r, s, screen, input, nil)
		if err != nil {
			return err
		}

		result.WriteResponse(w, r)
		return nil
	})

	h.Controller.HandleStartOfFlow(r.Context(), w, r, opts, authflow.FlowTypePromote, &handlers, nil)
}
