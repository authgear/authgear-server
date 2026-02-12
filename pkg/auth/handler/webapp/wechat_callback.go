package webapp

import (
	"context"
	"fmt"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/webappoauth"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

const WechatActionCallback = "callback"

func ConfigureWechatCallbackRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/sso/wechat/callback")
}

type WechatCallbackHandlerOAuthStateStore interface {
	PopAndRecoverState(ctx context.Context, stateToken string) (state *webappoauth.WebappOAuthState, err error)
}

// WechatCallbackHandler receives WeChat authorization result (code or error)
// and set it into the web session.
// Refreshing original auth ui WeChat auth page (/sso/wechat/auth/:alias) will
// get and consume the result from the web session.
//
// In web, user will use their WeChat app to scan the qr code in auth ui WeChat
// auth page, then they will complete authorization in their WeChat app and
// redirect to this endpoint through WeChat. This endpoint will set the result
// in web session and instruct user go back to the original auth ui. The
// original auth ui will refresh and proceed.
//
// In native app, the app will receive delegate function call when user click
// login in with WeChat. Developer needs to implement and obtain WeChat
// authorization code through native WeChat SDK. After obtaining the code,
// developer needs to call this endpoint with code through Authgear SDK. At this
// moment, user can click the proceed button in auth ui WeChat auth page to
// continue.
type WechatCallbackHandler struct {
	ControllerFactory ControllerFactory
	BaseViewModel     *viewmodels.BaseViewModeler
	OAuthStateStore   WechatCallbackHandlerOAuthStateStore
}

func (h *WechatCallbackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.ServeWithDBTx(r.Context())

	stateToken := r.Form.Get("state")

	updateWebSession := func(ctx context.Context) error {
		state, err := h.OAuthStateStore.PopAndRecoverState(ctx, stateToken)
		if err != nil {
			return err
		}

		switch state.UIImplementation {
		case config.UIImplementationAuthflowV2:
			session, err := ctrl.GetSession(ctx, state.WebSessionID)
			if err != nil {
				return err
			}

			screen, ok := session.Authflow.AllScreens[state.XStep]
			if !ok {
				return webapp.WebUIInvalidSession.New("x_step does not reference a valid screen")
			}

			screen.WechatCallbackData = &webapp.AuthflowWechatCallbackData{
				State:            stateToken,
				WebappOAuthState: state,
				Query:            r.Form.Encode(),
			}

			err = ctrl.UpdateSession(ctx, session)
			if err != nil {
				return err
			}

			return nil
		default:
			panic(fmt.Errorf("expected ui implementation to be set in state"))
		}
	}

	handler := func(ctx context.Context) error {
		err := updateWebSession(ctx)
		// serve api
		baseViewModel := h.BaseViewModel.ViewModel(r, w)
		if baseViewModel.IsNativePlatform {
			if err == nil {
				httputil.WriteJSONResponse(ctx, w, &api.Response{Result: "OK"})
			} else {
				httputil.WriteJSONResponse(ctx, w, &api.Response{Error: err})
			}
			return nil
		}

		// serve webapp page
		if err != nil {
			return err
		}
		result := &webapp.Result{
			RedirectURI: "/errors/return",
		}
		result.WriteResponse(w, r)

		return nil
	}
	ctrl.Get(handler)
	ctrl.PostAction("", handler)
}
