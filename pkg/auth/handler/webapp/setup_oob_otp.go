package webapp

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction/nodes"
	"github.com/authgear/authgear-server/pkg/auth/dependency/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/core/authn"
	"github.com/authgear/authgear-server/pkg/core/phone"
	"github.com/authgear/authgear-server/pkg/db"
	"github.com/authgear/authgear-server/pkg/httproute"
	"github.com/authgear/authgear-server/pkg/template"
	"github.com/authgear/authgear-server/pkg/validation"
)

const (
	TemplateItemTypeAuthUISetupOOBOTPHTML config.TemplateItemType = "auth_ui_setup_oob_otp.html"
)

var TemplateAuthUISetupOOBOTPHTML = template.Spec{
	Type:        TemplateItemTypeAuthUISetupOOBOTPHTML,
	IsHTML:      true,
	Translation: TemplateItemTypeAuthUITranslationJSON,
	Defines:     defines,
	Components:  components,
}

const SetupOOBOTPRequestSchema = "SetupOOBOTPRequestSchema"

var SetupOOBOTPSchema = validation.NewMultipartSchema("").
	Add(SetupOOBOTPRequestSchema, `
	{
		"type": "object",
		"properties": {
			"x_input_type": { "type": "string", "enum": ["email", "phone"] },
			"x_calling_code": { "type": "string" },
			"x_national_number": { "type": "string" },
			"x_email": { "type": "string" }
		},
		"required": ["x_input_type"],
		"allOf": [
			{
				"if": {
					"properties": {
						"x_input_type": { "type": "string", "const": "phone" }
					}
				},
				"then": {
					"required": ["x_calling_code", "x_national_number"]
				}
			},
			{
				"if": {
					"properties": {
						"x_input_type": { "type": "string", "enum": ["email"] }
					}
				},
				"then": {
					"required": ["x_email"]
				}
			}
		]
	}
	`).Instantiate()

func ConfigureSetupOOBOTPRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/setup_oob_otp")
}

type SetupOOBOTPNode interface {
	GetAllowedChannels() ([]authn.AuthenticatorOOBChannel, error)
}

type SetupOOBOTPViewModel struct {
	// InputType is either phone or email.
	InputType string
}

func NewSetupOOBOTPViewModel(graph *newinteraction.Graph, inputType string) SetupOOBOTPViewModel {
	var node SetupOOBOTPNode
	if !graph.FindLastNode(&node) {
		panic("setup_oob_otp: expected graph has node implementing SetupOOBOTPNode")
	}

	allowedChannels, err := node.GetAllowedChannels()
	if err != nil {
		panic(fmt.Errorf("setup_oob_otp: unexpected error: %w", err))
	}

	phoneAllowed := false
	emailAllowed := false

	for _, channel := range allowedChannels {
		switch channel {
		case authn.AuthenticatorOOBChannelEmail:
			emailAllowed = true
		case authn.AuthenticatorOOBChannelSMS:
			phoneAllowed = true
		}
	}

	if !phoneAllowed && !emailAllowed {
		panic("webapp: expected allowed channels to be non-empty")
	}

	switch inputType {
	case "phone":
		if !phoneAllowed {
			inputType = "email"
		}
	case "email":
		if !emailAllowed {
			inputType = "phone"
		}
	default:
		if phoneAllowed {
			inputType = "phone"
		} else if emailAllowed {
			inputType = "email"
		}
	}

	return SetupOOBOTPViewModel{
		InputType: inputType,
	}
}

type SetupOOBOTPInput struct {
	InputType string
	Target    string
}

var _ nodes.InputCreateAuthenticatorOOBSetup = &SetupOOBOTPInput{}

func (i *SetupOOBOTPInput) GetOOBChannel() authn.AuthenticatorOOBChannel {
	switch i.InputType {
	case "email":
		return authn.AuthenticatorOOBChannelEmail
	case "phone":
		return authn.AuthenticatorOOBChannelSMS
	default:
		panic("webapp: unknown input type: " + i.InputType)
	}
}

func (i *SetupOOBOTPInput) GetOOBTarget() string {
	return i.Target
}

type SetupOOBOTPHandler struct {
	Database      *db.Handle
	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      Renderer
	WebApp        WebAppService
}

func (h *SetupOOBOTPHandler) GetData(r *http.Request, state *webapp.State, graph *newinteraction.Graph) (map[string]interface{}, error) {
	data := map[string]interface{}{}

	var anyError interface{}
	if state != nil {
		anyError = state.Error
	}

	baseViewModel := h.BaseViewModel.ViewModel(r, anyError)
	viewModel := NewSetupOOBOTPViewModel(graph, r.Form.Get("x_input_type"))

	viewmodels.Embed(data, baseViewModel)
	viewmodels.Embed(data, viewModel)
	return data, nil
}

func (h *SetupOOBOTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Populate default values.
	if _, ok := r.Form["x_input_type"]; !ok {
		r.Form.Set("x_input_type", "email")
	}

	if r.Method == "GET" {
		h.Database.WithTx(func() error {
			state, graph, err := h.WebApp.Get(StateID(r))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return err
			}

			data, err := h.GetData(r, state, graph)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return err
			}

			h.Renderer.RenderHTML(w, r, TemplateItemTypeAuthUISetupOOBOTPHTML, data)
			return nil
		})
	}

	if r.Method == "POST" {
		h.Database.WithTx(func() error {
			result, err := h.WebApp.PostInput(StateID(r), func() (input interface{}, err error) {
				err = SetupOOBOTPSchema.PartValidator(SetupOOBOTPRequestSchema).ValidateValue(FormToJSON(r.Form))
				if err != nil {
					return
				}

				target, inputType, err := FormToOOBTarget(r.Form)

				input = &SetupOOBOTPInput{
					InputType: inputType,
					Target:    target,
				}
				return
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return err
			}
			result.WriteResponse(w, r)
			return nil
		})
	}
}

func FormToOOBTarget(form url.Values) (target string, inputType string, err error) {
	if form.Get("x_input_type") == "phone" {
		nationalNumber := form.Get("x_national_number")
		countryCallingCode := form.Get("x_calling_code")
		var e164 string
		e164, err = phone.Parse(nationalNumber, countryCallingCode)
		if err != nil {
			err = &validation.AggregatedError{
				Errors: []validation.Error{{
					Keyword:  "format",
					Location: "/x_national_number",
					Info: map[string]interface{}{
						"format": "phone",
					},
				}},
			}
			return
		}

		target = e164
		inputType = "phone"
		return
	}

	target = form.Get("x_email")
	inputType = "email"
	return
}
