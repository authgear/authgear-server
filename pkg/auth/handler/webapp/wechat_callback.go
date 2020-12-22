package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

const WechatActionCallback = "callback"

type JSONResponseWriter interface {
	WriteResponse(rw http.ResponseWriter, resp *api.Response)
}

func ConfigureWechatCallbackRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/sso/wechat/callback/:alias")
}

type WechatCallbackHandler struct {
	ControllerFactory ControllerFactory
	JSON              JSONResponseWriter
}

func (h *WechatCallbackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.Serve()

	sessionID := r.Form.Get("state")

	updateWebSession := func() error {
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

		return nil
	}

	handler := func() error {
		err := updateWebSession()
		// serve api
		if httputil.IsJSONContentType(r.Header.Get("content-type")) {
			if err == nil {
				h.JSON.WriteResponse(w, &api.Response{Result: nil})
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
			RedirectURI: "/return",
		}
		result.WriteResponse(w, r)

		return nil
	}
	ctrl.Get(handler)
	ctrl.PostAction("", handler)
}
