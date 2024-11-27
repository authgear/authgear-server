package webapp

import (
	"context"
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/interaction/intents"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureReauthRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "GET").
		WithPathPattern("/reauth")
}

type ReauthHandler struct {
	ControllerFactory ControllerFactory
}

func (h *ReauthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	defer ctrl.ServeWithDBTx(r.Context())

	webSession := webapp.GetSession(r.Context())
	userIDHint := ""
	suppressIDPSessionCookie := false
	if webSession != nil {
		userIDHint = webSession.UserIDHint
		suppressIDPSessionCookie = webSession.SuppressIDPSessionCookie
	}

	ctrl.Get(func(ctx context.Context) error {
		opts := webapp.SessionOptions{
			RedirectURI: ctrl.RedirectURI(),
		}
		intent := &intents.IntentReauthenticate{
			UserIDHint:               userIDHint,
			SuppressIDPSessionCookie: suppressIDPSessionCookie,
		}
		result, err := ctrl.EntryPointPost(ctx, opts, intent, func() (input interface{}, err error) {
			return nil, nil
		})
		if err != nil {
			return err
		}
		result.WriteResponse(w, r)
		return nil
	})
}
