package webapp

import (
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/setutil"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebConfirmTerminateOtherSessionsHTML = template.RegisterHTML(
	"web/confirm_terminate_other_sessions.html",
	Components...,
)

var ConfirmTerminateOtherSessionsSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"x_response": { "type": "string" }
		},
		"required": ["x_response"]
	}
`)

func ConfigureConfirmTerminateOtherSessionsRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/flows/confirm_terminate_other_sessions")
}

type ConfirmTerminateOtherSessionsEndpointsProvider interface {
	SelectAccountEndpointURL() *url.URL
}

type ConfirmTerminateOtherSessionsHandler struct {
	ControllerFactory ControllerFactory
	BaseViewModel     *viewmodels.BaseViewModeler
	Renderer          Renderer
	Endpoints         ConfirmTerminateOtherSessionsEndpointsProvider
}

func (h *ConfirmTerminateOtherSessionsHandler) GetData(r *http.Request, rw http.ResponseWriter, session *webapp.Session, graph *interaction.Graph) (map[string]interface{}, error) {
	data := map[string]interface{}{}

	baseViewModel := h.BaseViewModel.ViewModel(r, rw)

	viewmodels.Embed(data, baseViewModel)

	return data, nil
}

func (h *ConfirmTerminateOtherSessionsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

		data, err := h.GetData(r, w, session, graph)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebConfirmTerminateOtherSessionsHTML, data)
		return nil
	})

	ctrl.PostAction("", func() error {
		err = ConfirmTerminateOtherSessionsSchema.Validator().ValidateValue(FormToJSON(r.Form))
		if err != nil {
			return err
		}

		session, err := ctrl.InteractionSession()
		if err != nil {
			return err
		}

		response := r.Form.Get("x_response")
		var isConfirmed = response == "confirm"

		if !isConfirmed {
			// If cancelled, forget all existing steps
			session.Steps = []webapp.SessionStep{}
			if err = ctrl.Page.UpdateSession(session); err != nil {
				return err
			}
			u := h.Endpoints.SelectAccountEndpointURL()
			result := &webapp.Result{
				RedirectURI: u.String(),
				RemoveQueries: setutil.Set[string]{
					"x_step": struct{}{},
				},
			}
			result.WriteResponse(w, r)
			return nil
		}

		result, err := ctrl.InteractionPost(func() (input interface{}, err error) {
			input = &InputConfirmTerminateOtherSessions{
				IsConfirm: isConfirmed,
			}
			return
		})
		if err != nil {
			return err
		}

		result.WriteResponse(w, r)
		return nil
	})
}
