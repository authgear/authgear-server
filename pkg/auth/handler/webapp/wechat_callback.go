package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/webappoauth"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

const WechatActionCallback = "callback"

type JSONResponseWriter interface {
	WriteResponse(rw http.ResponseWriter, resp *api.Response)
}

func ConfigureWechatCallbackRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/sso/wechat/callback")
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
	JSON              JSONResponseWriter
}

func (h *WechatCallbackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.Serve()

	state := r.Form.Get("state")
	code := r.Form.Get("code")
	error_ := r.Form.Get("error")
	errorDescription := r.Form.Get("error_description")

	updateWebSession := func() error {
		if authflowOAuthState, err := webappoauth.DecodeWebappOAuthState(state); err == nil {
			session, err := ctrl.GetSession(authflowOAuthState.WebSessionID)
			if err != nil {
				return err
			}

			screen, ok := session.Authflow.AllScreens[authflowOAuthState.XStep]
			if !ok {
				return webapp.WebUIInvalidSession.New("x_step does not reference a valid screen")
			}

			screen.WechatCallbackData = &webapp.AuthflowWechatCallbackData{
				State:            state,
				Code:             code,
				Error:            error_,
				ErrorDescription: errorDescription,
			}

			err = ctrl.UpdateSession(session)
			if err != nil {
				return err
			}

			return nil
		}

		// interaction
		webSessionID := state
		session, err := ctrl.GetSession(webSessionID)
		if err != nil {
			return err
		}

		step := session.CurrentStep()
		step.FormData["x_action"] = WechatActionCallback
		step.FormData["x_code"] = code
		step.FormData["x_error"] = error_
		step.FormData["x_error_description"] = errorDescription
		session.Steps[len(session.Steps)-1] = step

		err = ctrl.UpdateSession(session)
		if err != nil {
			return err
		}

		return nil
	}

	handler := func() error {
		err := updateWebSession()
		// serve api
		baseViewModel := h.BaseViewModel.ViewModel(r, w)
		if baseViewModel.IsNativePlatform {
			if err == nil {
				h.JSON.WriteResponse(w, &api.Response{Result: "OK"})
			} else {
				h.JSON.WriteResponse(w, &api.Response{Error: err})
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
