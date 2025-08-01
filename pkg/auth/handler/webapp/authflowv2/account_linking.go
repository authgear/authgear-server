package authflowv2

import (
	"context"
	"fmt"
	"net/http"

	"strconv"

	"github.com/authgear/authgear-server/pkg/api/model"
	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow/declarative"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebAuthflowV2AccountLinkingHTML = template.RegisterHTML(
	"web/authflowv2/account_linking.html",
	handlerwebapp.Components...,
)

var AuthflowV2AccountLinkingIdentifySchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"x_index": { "type": "string" }
		},
		"required": ["x_index"]
	}
`)

func ConfigureAuthflowV2AccountLinkingRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(AuthflowV2RouteAccountLinking)
}

type AuthflowV2AccountLinkingOption struct {
	Identification    model.AuthenticationFlowIdentification
	MaskedDisplayName string
	ProviderType      string
	ProviderStatus    config.OAuthProviderStatus
	Index             int
}

type AuthflowV2AccountLinkingViewModel struct {
	Options []AuthflowV2AccountLinkingOption
	Data    declarative.AccountLinkingIdentifyData
}

type AuthflowV2AccountLinkingHandler struct {
	Controller    *handlerwebapp.AuthflowController
	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      handlerwebapp.Renderer
}

func NewAuthflowV2AccountLinkingViewModel(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) AuthflowV2AccountLinkingViewModel {
	flowResponse := screen.StateTokenFlowResponse
	data := flowResponse.Action.Data.(declarative.AccountLinkingIdentifyData)

	options := []AuthflowV2AccountLinkingOption{}

	for idx, option := range data.Options {
		idx := idx
		option := option

		options = append(options, AuthflowV2AccountLinkingOption{
			Identification:    option.Identifcation,
			MaskedDisplayName: option.MaskedDisplayName,
			ProviderType:      option.ProviderType,
			ProviderStatus:    option.ProviderStatus,
			Index:             idx,
		})
	}

	return AuthflowV2AccountLinkingViewModel{
		Options: options,
		Data:    data,
	}
}

func (h *AuthflowV2AccountLinkingHandler) GetData(w http.ResponseWriter, r *http.Request, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	baseViewModel := h.BaseViewModel.ViewModelForAuthFlow(r, w)
	viewmodels.Embed(data, baseViewModel)

	screenViewModel := NewAuthflowV2AccountLinkingViewModel(s, screen)
	viewmodels.Embed(data, screenViewModel)

	return data, nil
}

func (h *AuthflowV2AccountLinkingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var handlers handlerwebapp.AuthflowControllerHandlers
	handlers.Get(func(ctx context.Context, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		data, err := h.GetData(w, r, s, screen)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebAuthflowV2AccountLinkingHTML, data)
		return nil
	})
	handlers.PostAction("", func(ctx context.Context, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		err := AuthflowV2AccountLinkingIdentifySchema.Validator().ValidateValue(ctx, handlerwebapp.FormToJSON(r.Form))
		if err != nil {
			return err
		}

		index, err := strconv.Atoi(r.Form.Get("x_index"))
		if err != nil {
			return err
		}
		flowResponse := screen.StateTokenFlowResponse
		data := flowResponse.Action.Data.(declarative.AccountLinkingIdentifyData)
		option := data.Options[index]

		var input map[string]interface{}
		switch option.Identifcation {
		case model.AuthenticationFlowIdentificationEmail:
			fallthrough
		case model.AuthenticationFlowIdentificationPhone:
			fallthrough
		case model.AuthenticationFlowIdentificationUsername:
			input = map[string]interface{}{
				"index": index,
			}
		case model.AuthenticationFlowIdentificationOAuth:
			providerAlias := option.Alias
			screenViewModel := NewAuthflowV2AccountLinkingViewModel(s, screen)
			redirectURI, err := h.Controller.GetAccountLinkingSSOCallbackURL(providerAlias, screenViewModel.Data)
			if err != nil {
				return err
			}
			input = map[string]interface{}{
				"index":        index,
				"redirect_uri": redirectURI,
			}
		case model.AuthenticationFlowIdentificationLDAP:
			// TODO(DEV-1672): Support Account Linking for LDAP
			panic(fmt.Errorf("To be implemented identifcation option %v", option.Identifcation))
		default:
			panic(fmt.Errorf("unsupported identifcation option %v", option.Identifcation))
		}

		result, err := h.Controller.AdvanceWithInput(ctx, r, s, screen, input, nil)
		if err != nil {
			return err
		}

		result.WriteResponse(w, r)
		return nil
	})

	h.Controller.HandleStep(r.Context(), w, r, &handlers)
}
