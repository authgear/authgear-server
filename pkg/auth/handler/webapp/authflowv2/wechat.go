package authflowv2

import (
	"context"
	htmltemplate "html/template"
	"net/http"

	"net/url"

	"github.com/boombuler/barcode/qr"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow/declarative"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/webappoauth"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	coreimage "github.com/authgear/authgear-server/pkg/util/image"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/urlutil"
	"github.com/authgear/authgear-server/pkg/util/wechat"
)

var TemplateWebAuthflowWechatHTML = template.RegisterHTML(
	"web/authflowv2/wechat.html",
	handlerwebapp.Components...,
)

func ConfigureAuthflowV2WechatRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "GET", "POST").
		WithPathPattern(AuthflowV2RouteWechat)
}

type AuthflowWechatViewModel struct {
	ImageURI          htmltemplate.URL
	WechatRedirectURI htmltemplate.URL
}

type AuthflowV2WechatHandlerOAuthStateStore interface {
	GenerateState(ctx context.Context, state *webappoauth.WebappOAuthState) (stateToken string, err error)
}

type AuthflowV2WechatHandler struct {
	AppID           config.AppID
	Controller      *handlerwebapp.AuthflowController
	BaseViewModel   *viewmodels.BaseViewModeler
	Renderer        handlerwebapp.Renderer
	OAuthStateStore AuthflowV2WechatHandlerOAuthStateStore
}

func (h *AuthflowV2WechatHandler) GetData(ctx context.Context, w http.ResponseWriter, r *http.Request, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	baseViewModel := h.BaseViewModel.ViewModelForAuthFlow(r, w)
	viewmodels.Embed(data, baseViewModel)

	screenData := screen.StateTokenFlowResponse.Action.Data.(declarative.OAuthData)
	state := &webappoauth.WebappOAuthState{
		AppID:            string(h.AppID),
		WebSessionID:     s.ID,
		UIImplementation: config.UIImplementationAuthflowV2,
		XStep:            screen.Screen.StateToken.XStep,
		ErrorRedirectURI: (&url.URL{
			Path:     r.URL.Path,
			RawQuery: r.URL.Query().Encode(),
		}).String(),
		ProviderAlias: screenData.Alias,
	}
	stateToken, err := h.OAuthStateStore.GenerateState(ctx, state)
	if err != nil {
		return nil, err
	}

	authorizationURL, err := url.Parse(screenData.OAuthAuthorizationURL)
	if err != nil {
		return nil, err
	}
	authorizationURL = urlutil.WithQueryParamsAdded(authorizationURL, map[string]string{"state": stateToken})

	img, err := handlerwebapp.CreateQRCodeImage(authorizationURL.String(), 512, 512, qr.M)
	if err != nil {
		return nil, err
	}
	dataURI, err := coreimage.DataURIFromImage(coreimage.CodecPNG, img)
	if err != nil {
		return nil, err
	}

	screenViewModel := AuthflowWechatViewModel{
		// nolint: gosec
		ImageURI: htmltemplate.URL(dataURI),
	}
	wechatRedirectURI := wechat.GetWeChatRedirectURI(ctx)
	if wechatRedirectURI != "" {
		u, err := url.Parse(wechatRedirectURI)
		if err != nil {
			return nil, err
		}
		q := u.Query()
		q.Set("state", stateToken)
		u.RawQuery = q.Encode()
		// nolint: gosec
		screenViewModel.WechatRedirectURI = htmltemplate.URL(u.String())
	} else {
		if baseViewModel.IsNativePlatform {
			return nil, apierrors.NewInvalid("missing wechat redirect uri")
		}
	}
	viewmodels.Embed(data, screenViewModel)

	return data, nil
}

func (h *AuthflowV2WechatHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var handlers handlerwebapp.AuthflowControllerHandlers

	submit := func(ctx context.Context, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		data := screen.Screen.WechatCallbackData
		h.Controller.HandleOAuthCallback(ctx, w, r, handlerwebapp.AuthflowOAuthCallbackResponse{
			State: data.WebappOAuthState,
			Query: data.Query,
		})
		return nil
	}

	handlers.Get(func(ctx context.Context, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		if screen.Screen.WechatCallbackData != nil {
			return submit(ctx, s, screen)
		}

		// Otherwise render the page.
		data, err := h.GetData(ctx, w, r, s, screen)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebAuthflowWechatHTML, data)
		return nil
	})
	handlers.PostAction("", func(ctx context.Context, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		if screen.Screen.WechatCallbackData != nil {
			return submit(ctx, s, screen)
		}

		// Otherwise redirect to the same page.
		redirectURI := &url.URL{
			Path:     r.URL.Path,
			RawQuery: r.URL.Query().Encode(),
		}
		result := &webapp.Result{
			NavigationAction: webapp.NavigationActionReplace,
			RedirectURI:      redirectURI.String(),
		}
		result.WriteResponse(w, r)
		return nil
	})
	h.Controller.HandleStep(r.Context(), w, r, &handlers)
}
