package webapp

import (
	"fmt"
	htmltemplate "html/template"
	"net/http"
	"net/url"
	"time"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	coreimage "github.com/authgear/authgear-server/pkg/util/image"
	"github.com/authgear/authgear-server/pkg/util/secretcode"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebSetupTOTPHTML = template.RegisterHTML(
	"web/setup_totp.html",
	Components...,
)

var SetupTOTPSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"x_totp_code": {
				"type": "string",
				"format": "x_totp_code"
			}
		},
		"required": ["x_totp_code"]
	}
`)

func ConfigureSetupTOTPRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/flows/setup_totp")
}

type SetupTOTPViewModel struct {
	ImageURI         htmltemplate.URL
	Secret           string
	AlternativeSteps []viewmodels.AlternativeStep
}

type SetupTOTPNode interface {
	GetTOTPAuthenticator() *authenticator.Info
}

type SetupTOTPEndpointsProvider interface {
	Origin() *url.URL
}
type SetupTOTPHandler struct {
	ControllerFactory         ControllerFactory
	BaseViewModel             *viewmodels.BaseViewModeler
	AlternativeStepsViewModel *viewmodels.AlternativeStepsViewModeler
	Renderer                  Renderer
	Clock                     clock.Clock
	Endpoints                 SetupTOTPEndpointsProvider
}

func (h *SetupTOTPHandler) MakeViewModel(session *webapp.Session, graph *interaction.Graph) (*SetupTOTPViewModel, error) {
	var node SetupTOTPNode
	if !graph.FindLastNode(&node) {
		panic(fmt.Errorf("setup_totp: expected graph has node implementing SetupTOTPNode"))
	}

	a := node.GetTOTPAuthenticator()
	secret := a.TOTP.Secret
	totp, err := secretcode.NewTOTPFromSecret(secret)
	if err != nil {
		return nil, err
	}

	issuer := h.Endpoints.Origin().String()
	// FIXME(mfa): decide a proper account name.
	// We cannot use graph.MustGetUserLastIdentity because
	// In settings, the interaction may not have identity.
	accountName := "user"
	opts := secretcode.QRCodeImageOptions{
		Issuer:      issuer,
		AccountName: accountName,
		Width:       512,
		Height:      512,
	}

	img, err := totp.QRCodeImage(opts)
	if err != nil {
		return nil, err
	}

	dataURI, err := coreimage.DataURIFromImage(coreimage.CodecPNG, img)
	if err != nil {
		return nil, err
	}

	alternatives, err := h.AlternativeStepsViewModel.CreateAuthenticatorAlternatives(graph, webapp.SessionStepSetupTOTP)
	if err != nil {
		return nil, err
	}

	return &SetupTOTPViewModel{
		Secret: secret,
		// dataURI is generated here and not user generated,
		// so it is safe to use htmltemplate.URL with it.
		// nolint:gosec
		ImageURI:         htmltemplate.URL(dataURI),
		AlternativeSteps: alternatives.AlternativeSteps,
	}, nil
}

func (h *SetupTOTPHandler) GetData(r *http.Request, rw http.ResponseWriter, session *webapp.Session, graph *interaction.Graph) (map[string]interface{}, error) {
	data := map[string]interface{}{}
	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	viewModel, err := h.MakeViewModel(session, graph)
	if err != nil {
		return nil, err
	}

	viewmodels.Embed(data, baseViewModel)
	viewmodels.Embed(data, *viewModel)
	return data, nil
}

func (h *SetupTOTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.ServeWithDBTx()

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

		h.Renderer.RenderHTML(w, r, TemplateWebSetupTOTPHTML, data)
		return nil
	})

	ctrl.PostAction("", func() error {
		result, err := ctrl.InteractionPost(func() (input interface{}, err error) {
			err = SetupTOTPSchema.Validator().ValidateValue(FormToJSON(r.Form))
			if err != nil {
				return
			}

			now := h.Clock.NowUTC()

			// FIXME(mfa): decide a proper display name.
			displayName := fmt.Sprintf("TOTP @ %s", now.Format(time.RFC3339))

			input = &InputSetupTOTP{
				Code:        r.Form.Get("x_totp_code"),
				DisplayName: displayName,
			}
			return
		})
		if err != nil {
			return err
		}

		result.WriteResponse(w, r)
		return nil
	})

	handleAlternativeSteps(ctrl)
}
