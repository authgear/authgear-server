package webapp

import (
	"context"
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/interaction/nodes"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebAccountStatusHTML = template.RegisterHTML(
	"web/account_status.html",
	Components...,
)

func ConfigureAccountStatusRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "GET").
		WithPathPattern("/flows/account_status")
}

type AccountStatusHandler struct {
	ControllerFactory ControllerFactory
	BaseViewModel     *viewmodels.BaseViewModeler
	Renderer          Renderer
}

func (h *AccountStatusHandler) GetData(r *http.Request, rw http.ResponseWriter, graph *interaction.Graph) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	if node, ok := graph.CurrentNode().(*nodes.NodeValidateUser); ok {
		baseViewModel.SetError(node.Error)
	}
	viewmodels.Embed(data, baseViewModel)
	return data, nil
}

func (h *AccountStatusHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.ServeWithDBTx(r.Context())

	ctrl.Get(func(ctx context.Context) error {
		graph, err := ctrl.InteractionGet(ctx)
		if err != nil {
			return err
		}

		data, err := h.GetData(r, w, graph)
		if err != nil {
			return err
		}

		webSession := webapp.GetSession(ctx)
		if webSession != nil {
			// complete the interaction when user login with account
			// which has been disabled / deactivated / scheduled deletion
			err := ctrl.DeleteSession(ctx, webSession.ID)
			if err != nil {
				return err
			}
		}

		h.Renderer.RenderHTML(w, r, TemplateWebAccountStatusHTML, data)
		return nil
	})
}
