package webapp

import (
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebVerifyLoginLinkOTPHTML = template.RegisterHTML(
	"web/verify_login_link.html",
	Components...,
)

var VerifyLoginLinkOTPSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"x_oob_otp_code": { "type": "string" }
		},
		"required": ["x_oob_otp_code"]
	}
`)

func ConfigureVerifyLoginLinkOTPRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/flows/verify_login_link")
}

type VerifyLoginLinkOTPViewModel struct {
	Code       string
	StateQuery LoginLinkOTPPageQueryState
}

func NewVerifyLoginLinkOTPViewModel(r *http.Request) VerifyLoginLinkOTPViewModel {
	code := r.URL.Query().Get("code")

	return VerifyLoginLinkOTPViewModel{
		Code:       code,
		StateQuery: GetLoginLinkStateFromQuery(r),
	}
}

type WorkflowWebsocketEventStore interface {
	Publish(workflowID string, e workflow.Event) error
}

type AuthenticationFlowWebsocketEventStore interface {
	Publish(websocketChannelName string, e authflow.Event) error
}

type VerifyLoginLinkOTPHandler struct {
	LoginLinkOTPCodeService     OTPCodeService
	GlobalSessionServiceFactory *GlobalSessionServiceFactory
	ControllerFactory           ControllerFactory
	BaseViewModel               *viewmodels.BaseViewModeler
	AuthenticationViewModel     *viewmodels.AuthenticationViewModeler
	Renderer                    Renderer
	WorkflowEvents              WorkflowWebsocketEventStore
	AuthenticationFlowEvents    AuthenticationFlowWebsocketEventStore
	Config                      *config.AppConfig
}

func (h *VerifyLoginLinkOTPHandler) GetData(r *http.Request, rw http.ResponseWriter) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	viewmodels.Embed(data, NewVerifyLoginLinkOTPViewModel(r))
	viewmodels.Embed(data, baseViewModel)
	return data, nil
}

//nolint:gocognit
func (h *VerifyLoginLinkOTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.ServeWithDBTx()

	finishWithState := func(state LoginLinkOTPPageQueryState) {
		url := url.URL{}
		url.Path = r.URL.Path
		query := r.URL.Query()
		query.Set(LoginLinkOTPPageQueryStateKey, string(state))
		url.RawQuery = query.Encode()

		result := webapp.Result{
			RedirectURI:      url.String(),
			NavigationAction: "replace",
		}
		result.WriteResponse(w, r)
	}

	ctrl.Get(func() error {
		data, err := h.GetData(r, w)
		if err != nil {
			return err
		}

		if GetLoginLinkStateFromQuery(r) == LoginLinkOTPPageQueryStateInitial {
			code := r.URL.Query().Get("code")
			kind := otp.KindOOBOTPLink(h.Config, model.AuthenticatorOOBChannelEmail)

			target, err := h.LoginLinkOTPCodeService.LookupCode(kind.Purpose(), code)
			if apierrors.IsKind(err, otp.InvalidOTPCode) {
				finishWithState(LoginLinkOTPPageQueryStateInvalidCode)
				return nil
			} else if err != nil {
				return err
			}

			err = h.LoginLinkOTPCodeService.VerifyOTP(
				kind, target, code, &otp.VerifyOptions{SkipConsume: true},
			)
			if apierrors.IsKind(err, otp.InvalidOTPCode) {
				finishWithState(LoginLinkOTPPageQueryStateInvalidCode)
				return nil
			} else if err != nil {
				return err
			}
		}

		h.Renderer.RenderHTML(w, r, TemplateWebVerifyLoginLinkOTPHTML, data)
		return nil
	})

	ctrl.PostAction("", func() error {
		err := VerifyLoginLinkOTPSchema.Validator().ValidateValue(FormToJSON(r.Form))
		if err != nil {
			return err
		}

		code := r.Form.Get("x_oob_otp_code")
		kind := otp.KindOOBOTPLink(h.Config, model.AuthenticatorOOBChannelEmail)

		target, err := h.LoginLinkOTPCodeService.LookupCode(kind.Purpose(), code)
		if apierrors.IsKind(err, otp.InvalidOTPCode) {
			finishWithState(LoginLinkOTPPageQueryStateInvalidCode)
			return nil
		} else if err != nil {
			return err
		}

		state, err := h.LoginLinkOTPCodeService.SetSubmittedCode(kind, target, code)
		if apierrors.IsKind(err, otp.InvalidOTPCode) {
			finishWithState(LoginLinkOTPPageQueryStateInvalidCode)
			return nil
		} else if err != nil {
			return err
		}

		// Update the web session and trigger the refresh event
		if state.WebSessionID != "" {
			webSessionProvider := h.GlobalSessionServiceFactory.NewGlobalSessionService(
				h.Config.ID,
			)
			webSession, err := webSessionProvider.GetSession(state.WebSessionID)
			if err != nil {
				return err
			}
			err = webSessionProvider.UpdateSession(webSession)
			if err != nil {
				return err
			}
		}

		// For legacy workflow
		if state.WorkflowID != "" {
			err = h.WorkflowEvents.Publish(state.WorkflowID, workflow.NewEventRefresh())
			if err != nil {
				return err
			}
		}

		// For authentication flow
		if state.AuthenticationFlowWebsocketChannelName != "" {
			err = h.AuthenticationFlowEvents.Publish(state.AuthenticationFlowWebsocketChannelName, authflow.NewEventRefresh())
			if err != nil {
				return err
			}
		}

		finishWithState(LoginLinkOTPPageQueryStateMatched)
		return nil
	})
}
