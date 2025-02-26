package authflowv2

import (
	"context"
	"net/http"

	"net/url"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebVerifyLoginLinkOTPHTML = template.RegisterHTML(
	"web/authflowv2/verify_login_link.html",
	handlerwebapp.Components...,
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

func ConfigureAuthflowV2VerifyLoginLinkOTPRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(AuthflowV2RouteVerifyLink)
}

const LoginLinkOTPPageQueryStateKey = "state"

type LoginLinkOTPPageQueryState string

const (
	LoginLinkOTPPageQueryStateInitial     LoginLinkOTPPageQueryState = ""
	LoginLinkOTPPageQueryStateInvalidCode LoginLinkOTPPageQueryState = "invalid_code"
	LoginLinkOTPPageQueryStateMatched     LoginLinkOTPPageQueryState = "matched"
)

func (s *LoginLinkOTPPageQueryState) IsValid() bool {
	return *s == LoginLinkOTPPageQueryStateInitial ||
		*s == LoginLinkOTPPageQueryStateInvalidCode ||
		*s == LoginLinkOTPPageQueryStateMatched
}

func GetLoginLinkStateFromQuery(r *http.Request) LoginLinkOTPPageQueryState {
	p := LoginLinkOTPPageQueryState(
		r.URL.Query().Get(LoginLinkOTPPageQueryStateKey),
	)
	if p.IsValid() {
		return p
	}
	return LoginLinkOTPPageQueryStateInitial
}

type VerifyLoginLinkOTPViewModel struct {
	Code  string
	State LoginLinkOTPPageQueryState
}

func NewVerifyLoginLinkOTPViewModel(r *http.Request) VerifyLoginLinkOTPViewModel {
	code := r.URL.Query().Get("code")

	return VerifyLoginLinkOTPViewModel{
		Code:  code,
		State: GetLoginLinkStateFromQuery(r),
	}
}

type AuthenticationFlowWebsocketEventStore interface {
	Publish(ctx context.Context, websocketChannelName string, e authflow.Event) error
}

type AuthflowV2VerifyLoginLinkOTPHandler struct {
	LoginLinkOTPCodeService     handlerwebapp.OTPCodeService
	GlobalSessionServiceFactory *handlerwebapp.GlobalSessionServiceFactory
	ControllerFactory           handlerwebapp.ControllerFactory
	BaseViewModel               *viewmodels.BaseViewModeler
	AuthenticationViewModel     *viewmodels.AuthenticationViewModeler
	Renderer                    handlerwebapp.Renderer
	AuthenticationFlowEvents    AuthenticationFlowWebsocketEventStore
	Config                      *config.AppConfig
}

func (h *AuthflowV2VerifyLoginLinkOTPHandler) GetData(r *http.Request, rw http.ResponseWriter) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	viewmodels.Embed(data, NewVerifyLoginLinkOTPViewModel(r))
	viewmodels.Embed(data, baseViewModel)
	return data, nil
}

//nolint:gocognit
func (h *AuthflowV2VerifyLoginLinkOTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.ServeWithoutDBTx(r.Context())

	finishWithState := func(state LoginLinkOTPPageQueryState) {
		url := url.URL{}
		url.Path = r.URL.Path
		query := r.URL.Query()
		query.Set(LoginLinkOTPPageQueryStateKey, string(state))
		url.RawQuery = query.Encode()

		result := webapp.Result{
			RedirectURI:      url.String(),
			NavigationAction: webapp.NavigationActionReplace,
		}
		result.WriteResponse(w, r)
	}

	ctrl.Get(func(ctx context.Context) error {
		data, err := h.GetData(r, w)
		if err != nil {
			return err
		}

		if GetLoginLinkStateFromQuery(r) == LoginLinkOTPPageQueryStateInitial {
			code := r.URL.Query().Get("code")
			kind := otp.KindOOBOTPLink(h.Config, model.AuthenticatorOOBChannelEmail)

			target, err := h.LoginLinkOTPCodeService.LookupCode(ctx, kind.Purpose(), code)
			if apierrors.IsKind(err, otp.InvalidOTPCode) {
				finishWithState(LoginLinkOTPPageQueryStateInvalidCode)
				return nil
			} else if err != nil {
				return err
			}

			err = h.LoginLinkOTPCodeService.VerifyOTP(
				ctx, kind, target, code, &otp.VerifyOptions{SkipConsume: true},
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

	ctrl.PostAction("", func(ctx context.Context) error {
		err := VerifyLoginLinkOTPSchema.Validator().ValidateValue(ctx, handlerwebapp.FormToJSON(r.Form))
		if err != nil {
			return err
		}

		code := r.Form.Get("x_oob_otp_code")
		kind := otp.KindOOBOTPLink(h.Config, model.AuthenticatorOOBChannelEmail)

		target, err := h.LoginLinkOTPCodeService.LookupCode(ctx, kind.Purpose(), code)
		if apierrors.IsKind(err, otp.InvalidOTPCode) {
			finishWithState(LoginLinkOTPPageQueryStateInvalidCode)
			return nil
		} else if err != nil {
			return err
		}

		state, err := h.LoginLinkOTPCodeService.SetSubmittedCode(ctx, kind, target, code)
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
			webSession, err := webSessionProvider.GetSession(ctx, state.WebSessionID)
			if err != nil {
				return err
			}
			err = webSessionProvider.UpdateSession(ctx, webSession)
			if err != nil {
				return err
			}
		}

		// For authentication flow
		if state.AuthenticationFlowWebsocketChannelName != "" {
			err = h.AuthenticationFlowEvents.Publish(ctx, state.AuthenticationFlowWebsocketChannelName, authflow.NewEventRefresh())
			if err != nil {
				return err
			}
		}

		finishWithState(LoginLinkOTPPageQueryStateMatched)
		return nil
	})
}
