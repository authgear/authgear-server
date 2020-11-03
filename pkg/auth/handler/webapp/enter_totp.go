package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/interaction/nodes"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebEnterTOTPHTML = template.RegisterHTML(
	"web/enter_totp.html",
	components...,
)

const EnterTOTPRequestSchema = "EnterTOTPRequestSchema"

var EnterTOTPSchema = validation.NewMultipartSchema("").
	Add(EnterTOTPRequestSchema, `
		{
			"type": "object",
			"properties": {
				"x_code": { "type": "string" }
			},
			"required": ["x_code"]
		}
	`).Instantiate()

func ConfigureEnterTOTPRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/enter_totp")
}

type EnterTOTPViewModel struct {
	Alternatives []AuthenticationAlternative
}

type EnterTOTPHandler struct {
	Database      *db.Handle
	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      Renderer
	WebApp        WebAppService
}

func (h *EnterTOTPHandler) GetData(r *http.Request, rw http.ResponseWriter, state *webapp.State, graph *interaction.Graph) (map[string]interface{}, error) {
	data := map[string]interface{}{}

	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	alternatives, err := DeriveAuthenticationAlternatives(
		// Use current state ID because the current node should be NodeAuthenticationBegin.
		state.ID,
		graph,
		AuthenticationTypeTOTP,
		"",
	)
	if err != nil {
		return nil, err
	}

	viewModel := EnterTOTPViewModel{
		Alternatives: alternatives,
	}

	viewmodels.Embed(data, baseViewModel)
	viewmodels.Embed(data, viewModel)

	return data, nil
}

type EnterTOTPInput struct {
	Code        string
	DeviceToken bool
}

var _ nodes.InputAuthenticationTOTP = &EnterTOTPInput{}
var _ nodes.InputCreateDeviceToken = &EnterTOTPInput{}

// GetTOTP implements InputAuthenticationTOTP.
func (i *EnterTOTPInput) GetTOTP() string {
	return i.Code
}

// CreateDeviceToken implements InputCreateDeviceToken.
func (i *EnterTOTPInput) CreateDeviceToken() bool {
	return i.DeviceToken
}

func (h *EnterTOTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

			h.Renderer.RenderHTML(w, r, TemplateWebEnterTOTPHTML, data)
			return nil
		})
		if err != nil {
			panic(err)
		}
	}

	if r.Method == "POST" {
		err := h.Database.WithTx(func() error {
			result, err := h.WebApp.PostInput(StateID(r), func() (input interface{}, err error) {
				err = EnterTOTPSchema.PartValidator(EnterTOTPRequestSchema).ValidateValue(FormToJSON(r.Form))
				if err != nil {
					return
				}

				code := r.Form.Get("x_code")
				deviceToken := r.Form.Get("x_device_token") == "true"

				input = &EnterTOTPInput{
					Code:        code,
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
