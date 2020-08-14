package webapp

import (
	"fmt"
	htmltemplate "html/template"
	"net/http"
	"net/url"
	"time"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/template"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/interaction/nodes"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	coreimage "github.com/authgear/authgear-server/pkg/util/image"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

const (
	TemplateItemTypeAuthUISetupTOTPHTML config.TemplateItemType = "auth_ui_setup_totp.html"
)

var TemplateAuthUISetupTOTPHTML = template.Spec{
	Type:        TemplateItemTypeAuthUISetupTOTPHTML,
	IsHTML:      true,
	Translation: TemplateItemTypeAuthUITranslationJSON,
	Defines:     defines,
	Components:  components,
}

const SetupTOTPRequestSchema = "SetupTOTPRequestSchema"

var SetupTOTPSchema = validation.NewMultipartSchema("").
	Add(SetupTOTPRequestSchema, `
	{
		"type": "object",
		"properties": {
			"x_code": { "type": "string" }
		},
		"required": ["x_code"]
	}
	`).Instantiate()

func ConfigureSetupTOTPRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/setup_totp")
}

type SetupTOTPViewModel struct {
	ImageURI     htmltemplate.URL
	Secret       string
	Alternatives []CreateAuthenticatorAlternative
}

type SetupTOTPNode interface {
	GetTOTPAuthenticator() *authenticator.Info
}

type SetupTOTPInput struct {
	Code        string
	DisplayName string
}

var _ nodes.InputCreateAuthenticatorTOTP = &SetupTOTPInput{}

// GetTOTP implements InputCreateAuthenticatorTOTP.
func (i *SetupTOTPInput) GetTOTP() string {
	return i.Code
}

// GetTOTPDisplayName implements InputCreateAuthenticatorTOTP.
func (i *SetupTOTPInput) GetTOTPDisplayName() string {
	return i.DisplayName
}

type SetupTOTPEndpointsProvider interface {
	BaseURL() *url.URL
}

type SetupTOTPHandler struct {
	Database      *db.Handle
	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      Renderer
	WebApp        WebAppService
	Clock         clock.Clock
	Endpoints     SetupTOTPEndpointsProvider
}

func (h *SetupTOTPHandler) MakeViewModel(stateID string, graph *interaction.Graph) (*SetupTOTPViewModel, error) {
	var node SetupTOTPNode
	if !graph.FindLastNode(&node) {
		panic(fmt.Errorf("setup_totp: expected graph has node implementing SetupTOTPNode"))
	}

	a := node.GetTOTPAuthenticator()
	secret := a.Secret

	issuer := h.Endpoints.BaseURL().String()
	// FIXME(mfa): decide a proper account name.
	// We cannot use graph.MustGetUserLastIdentity because
	// In settings, the interaction may not have identity.
	accountName := "user"
	opts := otp.MakeTOTPKeyOptions{
		Issuer:      issuer,
		AccountName: accountName,
		Secret:      secret,
	}
	key, err := otp.MakeTOTPKey(opts)
	if err != nil {
		return nil, err
	}

	img, err := key.Image(512, 512)
	if err != nil {
		return nil, err
	}

	dataURI, err := coreimage.DataURIFromImage(coreimage.CodecPNG, img)
	if err != nil {
		return nil, err
	}

	alternatives, err := DeriveCreateAuthenticatorAlternatives(
		stateID,
		graph,
		authn.AuthenticatorTypeTOTP,
	)
	if err != nil {
		return nil, err
	}

	return &SetupTOTPViewModel{
		Secret: secret,
		// dataURI is generated here and not user generated,
		// so it is safe to use htmltemplate.URL with it.
		// nolint:gosec
		ImageURI:     htmltemplate.URL(dataURI),
		Alternatives: alternatives,
	}, nil
}

func (h *SetupTOTPHandler) GetData(r *http.Request, state *webapp.State, graph *interaction.Graph) (map[string]interface{}, error) {
	data := map[string]interface{}{}

	var anyError interface{}
	if state != nil {
		anyError = state.Error
	}

	baseViewModel := h.BaseViewModel.ViewModel(r, anyError)
	// Use previous state ID because the current node should be NodeCreateAuthenticatorTOTPSetup.
	viewModel, err := h.MakeViewModel(state.PrevID, graph)
	if err != nil {
		return nil, err
	}

	viewmodels.Embed(data, baseViewModel)
	viewmodels.Embed(data, *viewModel)
	return data, nil
}

func (h *SetupTOTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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

			h.Renderer.RenderHTML(w, r, TemplateItemTypeAuthUISetupTOTPHTML, data)
			return nil
		})
	}

	if r.Method == "POST" {
		h.Database.WithTx(func() error {
			result, err := h.WebApp.PostInput(StateID(r), func() (input interface{}, err error) {
				err = SetupTOTPSchema.PartValidator(SetupTOTPRequestSchema).ValidateValue(FormToJSON(r.Form))
				if err != nil {
					return
				}

				now := h.Clock.NowUTC()

				// FIXME(mfa): decide a proper display name.
				displayName := fmt.Sprintf("TOTP @ %s", now.Format(time.RFC3339))

				input = &SetupTOTPInput{
					Code:        r.Form.Get("x_code"),
					DisplayName: displayName,
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
