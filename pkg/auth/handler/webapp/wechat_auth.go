package webapp

import (
	htmltemplate "html/template"
	"image"
	"net/http"
	"net/url"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	coreimage "github.com/authgear/authgear-server/pkg/util/image"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/urlutil"
	"github.com/authgear/authgear-server/pkg/util/wechat"
)

var TemplateWebWechatAuthHandlerHTML = template.RegisterHTML(
	"web/wechat_auth.html",
	components...,
)

func ConfigureWechatAuthRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "GET").
		WithPathPattern("/sso/wechat/auth/:alias")
}

type WeChatAuthViewModel struct {
	ImageURI          htmltemplate.URL
	WeChatRedirectURI htmltemplate.URL
	CurrentURI        string
}

type WechatAuthHandler struct {
	ControllerFactory ControllerFactory
	BaseViewModel     *viewmodels.BaseViewModeler
	Renderer          Renderer
	CSRFCookie        webapp.CSRFCookieDef
	IdentityConfig    *config.IdentityConfig
}

func (h *WechatAuthHandler) GetData(r *http.Request, w http.ResponseWriter, session *webapp.Session, graph *interaction.Graph) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	baseViewModel := h.BaseViewModel.ViewModel(r, w)

	authURL := r.Form.Get("x_auth_url")
	if authURL == "" {
		return nil, apierrors.NewInvalid("missing authorization url")
	}

	img, err := createQRCodeImage(authURL, 512, 512)
	if err != nil {
		return nil, err
	}

	dataURI, err := coreimage.DataURIFromImage(coreimage.CodecPNG, img)
	if err != nil {
		return nil, err
	}

	viewModel := WeChatAuthViewModel{
		// dataURI is generated here and not user generated,
		// so it is safe to use htmltemplate.URL with it.
		// nolint:gosec
		ImageURI:   htmltemplate.URL(dataURI),
		CurrentURI: r.URL.RequestURI(),
	}

	weChatRedirectURIFromCtx := wechat.GetWeChatRedirectURI(r.Context())
	if weChatRedirectURIFromCtx != "" {
		alias := httproute.GetParam(r, "alias")
		u, err := parseWeChatRedirectURI(h.IdentityConfig, alias, weChatRedirectURIFromCtx)
		if err != nil {
			return nil, err
		}
		weChatRedirectURI := urlutil.WithQueryParamsAdded(u, map[string]string{"state": session.ID}).String()
		// weChatRedirectURI is generated here and not user generated,
		// so it is safe to use htmltemplate.URL with it.
		// nolint:gosec
		viewModel.WeChatRedirectURI = htmltemplate.URL(weChatRedirectURI)
	} else {
		if baseViewModel.IsNativePlatform {
			return nil, apierrors.NewInvalid("missing wechat redirect uri")
		}
	}

	viewmodels.Embed(data, baseViewModel)
	viewmodels.Embed(data, viewModel)
	return data, nil
}

func (h *WechatAuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

		if session != nil {
			step := session.CurrentStep()
			action, ok := step.FormData["x_action"].(string)
			if ok && action == WechatActionCallback {
				nonceSource, _ := r.Cookie(h.CSRFCookie.Name)

				data := InputOAuthCallback{
					ProviderAlias:    httproute.GetParam(r, "alias"),
					NonceSource:      nonceSource,
					Code:             step.FormData["x_code"].(string),
					Scope:            step.FormData["x_scope"].(string),
					Error:            step.FormData["x_error"].(string),
					ErrorDescription: step.FormData["x_error_description"].(string),
				}

				result, err := ctrl.InteractionPost(func() (input interface{}, err error) {
					input = &data
					return
				})
				if err != nil {
					return err
				}
				result.WriteResponse(w, r)
				return nil

			}
		}

		data, err := h.GetData(r, w, session, graph)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebWechatAuthHandlerHTML, data)
		return nil
	})
}

func createQRCodeImage(content string, width int, height int) (image.Image, error) {
	b, err := qr.Encode(content, qr.M, qr.Auto)

	if err != nil {
		return nil, err
	}

	b, err = barcode.Scale(b, width, height)

	if err != nil {
		return nil, err
	}

	return b, nil
}

func parseWeChatRedirectURI(identityConfig *config.IdentityConfig, alias string, weChatRedirectURI string) (*url.URL, error) {
	providerConfig, ok := identityConfig.OAuth.GetProviderConfig(alias)
	if !ok {
		return nil, apierrors.NewInvalid("invalid sso alias")
	}

	allowedURIs := providerConfig.WeChatRedirectURIs

	uri, err := url.Parse(weChatRedirectURI)
	if err != nil {
		return nil, apierrors.NewInvalid("invalid wechat redirect URI")
	}

	allowed := false
	for _, u := range allowedURIs {
		if u == weChatRedirectURI {
			allowed = true
			break
		}
	}

	if !allowed {
		return nil, apierrors.NewInvalid("wechat redirect URI is not allowed")
	}

	return uri, nil
}
