package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebEnterPasswordHTML = template.RegisterHTML(
	"web/enter_password.html",
	components...,
)

const EnterPasswordRequestSchema = "EnterPasswordRequestSchema"

var EnterPasswordSchema = validation.NewMultipartSchema("").
	Add(EnterPasswordRequestSchema, `
		{
			"type": "object",
			"properties": {
				"x_password": { "type": "string" }
			},
			"required": ["x_password"]
		}
	`).Instantiate()

func ConfigureEnterPasswordRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/enter_password")
}

type EnterPasswordViewModel struct {
	IdentityDisplayID string
	Alternatives      []AuthenticationAlternative
}

type EnterPasswordHandler struct {
	Database      *db.Handle
	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      Renderer
	WebApp        WebAppService
}

func (h *EnterPasswordHandler) GetData(r *http.Request, rw http.ResponseWriter, state *webapp.State, graph *interaction.Graph) (map[string]interface{}, error) {
	data := map[string]interface{}{}

	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	identityInfo := graph.MustGetUserLastIdentity()

	alternatives, err := DeriveAuthenticationAlternatives(
		// Use current state ID because the current node should be NodeAuthenticationBegin.
		state.ID,
		graph,
		AuthenticationTypePassword,
		"",
	)
	if err != nil {
		return nil, err
	}

	enterPasswordViewModel := EnterPasswordViewModel{
		IdentityDisplayID: identityInfo.DisplayID(),
		Alternatives:      alternatives,
	}

	viewmodels.Embed(data, baseViewModel)
	viewmodels.Embed(data, enterPasswordViewModel)

	return data, nil
}

func (h *EnterPasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if r.Method == "GET" {
		err := h.Database.WithTx(func() error {
			state, graph, err := h.WebApp.Get(StateID(r))
			if err != nil {
				return err
			}

			data, err := h.GetData(r, w, state, graph)
			if err != nil {
				return err
			}

			h.Renderer.RenderHTML(w, r, TemplateWebEnterPasswordHTML, data)
			return nil
		})
		if err != nil {
			panic(err)
		}
	}

	if r.Method == "POST" {
		err := h.Database.WithTx(func() error {
			result, err := h.WebApp.PostInput(StateID(r), func() (input interface{}, err error) {
				err = EnterPasswordSchema.PartValidator(EnterPasswordRequestSchema).ValidateValue(FormToJSON(r.Form))
				if err != nil {
					return
				}

				plainPassword := r.Form.Get("x_password")
				deviceToken := r.Form.Get("x_device_token") == "true"

				input = &InputAuthPassword{
					Password:    plainPassword,
					DeviceToken: deviceToken,
				}
				return
			})
			if err != nil {
				return err
			}
			result.WriteResponse(w, r)
			return nil
		})
		if err != nil {
			panic(err)
		}
	}
}
