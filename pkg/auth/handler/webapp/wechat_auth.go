package webapp

import (
	htmltemplate "html/template"
	"net/http"
	"net/url"

	"github.com/boombuler/barcode/qr"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	coreimage "github.com/authgear/authgear-server/pkg/util/image"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/urlutil"
	"github.com/authgear/authgear-server/pkg/util/wechat"
)

var TemplateWebWechatAuthHandlerHTML = template.RegisterHTML(
	"web/wechat_auth.html",
	Components...,
)

func ConfigureWechatAuthRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "GET", "POST").
		WithPathPattern("/sso/wechat/auth/:alias")
}

type WeChatAuthViewModel struct {
	ImageURI          htmltemplate.URL
	WeChatRedirectURI htmltemplate.URL
}

type WechatAuthHandler struct {
	ControllerFactory ControllerFactory
	BaseViewModel     *viewmodels.BaseViewModeler
	Renderer          Renderer
}

func (h *WechatAuthHandler) GetData(r *http.Request, w http.ResponseWriter, session *webapp.Session, graph *interaction.Graph) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	baseViewModel := h.BaseViewModel.ViewModel(r, w)

	authURL := r.Form.Get("x_auth_url")
	if authURL == "" {
		return nil, apierrors.NewInvalid("missing authorization url")
	}

	img, err := CreateQRCodeImage(authURL, 512, 512, qr.M)
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
		ImageURI: htmltemplate.URL(dataURI),
	}

	weChatRedirectURIFromCtx := wechat.GetWeChatRedirectURI(r.Context())
	if weChatRedirectURIFromCtx != "" {
		u, err := url.Parse(weChatRedirectURIFromCtx)
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

//nolint:gocognit
func (h *WechatAuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.ServeWithDBTx()

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
				query := url.Values{}
				query.Set("code", step.FormData["x_code"].(string))
				query.Set("error", step.FormData["x_error"].(string))
				query.Set("error_description", step.FormData["x_error_description"].(string))

				data := InputOAuthCallback{
					ProviderAlias: httproute.GetParam(r, "alias"),
					Query:         query.Encode(),
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

	ctrl.PostAction("", func() error {
		session, err := ctrl.InteractionSession()
		if err != nil {
			return err
		}

		if session != nil {
			step := session.CurrentStep()
			action, ok := step.FormData["x_action"].(string)
			if ok && action == WechatActionCallback {
				query := url.Values{}
				query.Set("code", step.FormData["x_code"].(string))
				query.Set("error", step.FormData["x_error"].(string))
				query.Set("error_description", step.FormData["x_error_description"].(string))

				data := InputOAuthCallback{
					ProviderAlias: httproute.GetParam(r, "alias"),
					Query:         query.Encode(),
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

		// Otherwise redirect to the current page.
		redirectURI := httputil.HostRelative(r.URL).String()
		result := webapp.Result{
			NavigationAction: "replace",
			RedirectURI:      redirectURI,
		}
		result.WriteResponse(w, r)
		return nil
	})
}
