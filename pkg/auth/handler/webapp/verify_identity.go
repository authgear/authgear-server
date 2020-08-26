package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/interaction/intents"
	"github.com/authgear/authgear-server/pkg/util/errorutil"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

const (
	TemplateItemTypeAuthUIVerifyIdentityHTML string = "auth_ui_verify_identity.html"
)

var TemplateAuthUIVerifyIdentityHTML = template.T{
	Type:                    TemplateItemTypeAuthUIVerifyIdentityHTML,
	IsHTML:                  true,
	TranslationTemplateType: TemplateItemTypeAuthUITranslationJSON,
	Defines:                 defines,
	ComponentTemplateTypes:  components,
}

const VerifyIdentityRequestSchema = "VerifyIdentityRequestSchema"

var VerifyIdentitySchema = validation.NewMultipartSchema("").
	Add(VerifyIdentityRequestSchema, `
		{
			"type": "object",
			"properties": {
				"x_password": { "type": "string" }
			},
			"required": ["x_password"]
		}
	`).Instantiate()

func ConfigureVerifyIdentityRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/verify_identity")
}

type VerifyIdentityViewModel struct {
	VerificationCode             string
	VerificationCodeSendCooldown int
	VerificationCodeLength       int
	VerificationCodeChannel      string
	IdentityDisplayID            string
}

type VerifyIdentityHandler struct {
	Database      *db.Handle
	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      Renderer
	WebApp        WebAppService
}

type VerifyIdentityNode interface {
	GetVerificationIdentity() *identity.Info
	GetVerificationCodeChannel() string
	GetVerificationCodeSendCooldown() int
	GetVerificationCodeLength() int
}

func (h *VerifyIdentityHandler) MakeIntent(r *http.Request) *webapp.Intent {
	return &webapp.Intent{
		RedirectURI: "/verify_identity/success",
		KeepState:   true,
		Intent:      intents.NewIntentVerifyIdentityResume(r.Form.Get("state")),
	}
}

func (h *VerifyIdentityHandler) GetData(r *http.Request, state *webapp.State, graph *interaction.Graph) (map[string]interface{}, error) {
	data := map[string]interface{}{}

	baseViewModel := h.BaseViewModel.ViewModel(r, state.Error)
	viewModel := VerifyIdentityViewModel{
		VerificationCode: r.Form.Get("code"),
	}
	var n VerifyIdentityNode
	if graph.FindLastNode(&n) {
		viewModel.IdentityDisplayID = n.GetVerificationIdentity().DisplayID()
		viewModel.VerificationCodeSendCooldown = n.GetVerificationCodeSendCooldown()
		viewModel.VerificationCodeLength = n.GetVerificationCodeLength()
		viewModel.VerificationCodeChannel = n.GetVerificationCodeChannel()
	}

	viewmodels.Embed(data, baseViewModel)
	viewmodels.Embed(data, viewModel)

	return data, nil
}

func (h *VerifyIdentityHandler) GetErrorData(r *http.Request, err error) (map[string]interface{}, error) {
	data := map[string]interface{}{}

	baseViewModel := h.BaseViewModel.ViewModel(r, err)
	viewModel := VerifyIdentityViewModel{}

	viewmodels.Embed(data, baseViewModel)
	viewmodels.Embed(data, viewModel)

	return data, nil
}

type VerificationCodeResend struct{}

func (i *VerificationCodeResend) DoResend() {}

type VerificationCodeInput struct {
	Code string
}

// GetVerificationCode implements InputVerifyIdentityCheckCode.
func (i *VerificationCodeInput) GetVerificationCode() string {
	return i.Code
}

func (h *VerifyIdentityHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	inInteraction := true
	id := StateID(r)
	if id == "" {
		// Navigated from the link in verification message
		id = r.Form.Get("state")

		_, err := h.WebApp.GetState(id)
		if errorutil.Is(err, webapp.ErrInvalidState) {
			inInteraction = false
		} else if err != nil {
			panic(err)
		} else {
			// State still valid, resume the interaction
			inInteraction = true
		}
	}

	if r.Method == "GET" && inInteraction {
		err := h.Database.WithTx(func() error {
			state, graph, err := h.WebApp.Get(id)
			if err != nil {
				return err
			}

			data, err := h.GetData(r, state, graph)
			if err != nil {
				return err
			}

			h.Renderer.RenderHTML(w, r, TemplateItemTypeAuthUIVerifyIdentityHTML, data)
			return nil
		})
		if err != nil {
			panic(err)
		}
	}

	if r.Method == "GET" && !inInteraction {
		err := h.Database.WithTx(func() error {
			state, graph, err := h.WebApp.GetIntent(h.MakeIntent(r))
			var data map[string]interface{}
			if err != nil {
				data, err = h.GetErrorData(r, err)
			} else {
				data, err = h.GetData(r, state, graph)
			}

			if err != nil {
				return err
			}

			h.Renderer.RenderHTML(w, r, TemplateItemTypeAuthUIVerifyIdentityHTML, data)
			return nil
		})
		if err != nil {
			panic(err)
		}
	}

	trigger := r.Form.Get("trigger") == "true"

	if r.Method == "POST" && trigger {
		err := h.Database.WithTx(func() error {
			result, err := h.WebApp.PostInput(id, func() (input interface{}, err error) {
				input = &VerificationCodeResend{}
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
		return
	}

	if r.Method == "POST" {
		err := h.Database.WithTx(func() error {
			inputer := func() (input interface{}, err error) {
				err = VerifyIdentitySchema.PartValidator(VerifyIdentityRequestSchema).ValidateValue(FormToJSON(r.Form))
				if err != nil {
					return
				}

				code := r.Form.Get("x_password")

				input = &VerificationCodeInput{
					Code: code,
				}
				return
			}

			var result *webapp.Result
			var err error
			if inInteraction {
				result, err = h.WebApp.PostInput(id, inputer)
			} else {
				result, err = h.WebApp.PostIntent(h.MakeIntent(r), inputer)
			}
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
