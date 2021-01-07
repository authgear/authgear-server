package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

const WechatActionCallback = "callback"

func ConfigureWechatCallbackRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/sso/wechat/callback/:alias")
}

type WechatCallbackHandler struct {
	ControllerFactory ControllerFactory
}

func (h *WechatCallbackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.Serve()

	sessionID := r.Form.Get("state")

	handler := func() error {
		session, err := ctrl.GetSession(sessionID)
		if err != nil {
			return err
		}

		step := session.CurrentStep()
		step.FormData["x_action"] = WechatActionCallback
		step.FormData["x_alias"] = httproute.GetParam(r, "alias")
		step.FormData["x_code"] = r.Form.Get("code")
		step.FormData["x_scope"] = r.Form.Get("scope")
		step.FormData["x_error"] = r.Form.Get("error")
		step.FormData["x_error_description"] = r.Form.Get("error_description")
		session.Steps[len(session.Steps)-1] = step

		err = ctrl.UpdateSession(session)
		if err != nil {
			return err
		}

		result := &webapp.Result{
			RedirectURI: "/return",
		}
		result.WriteResponse(w, r)

		return nil
	}
	ctrl.Get(handler)
	ctrl.PostAction("", handler)
}
