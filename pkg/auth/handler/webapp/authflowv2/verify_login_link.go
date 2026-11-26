package authflowv2

import (
	"context"
	"net/http"

	"net/url"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	authapi "github.com/authgear/authgear-server/pkg/auth/api"
	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
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

type AuthflowV2VerifyLoginLinkOTPHandler struct {
	// APIController makes this page use the login link API, so a Custom UI
	// can do the same.
	APIController           *authapi.APIController
	ControllerFactory       handlerwebapp.ControllerFactory
	BaseViewModel           *viewmodels.BaseViewModeler
	AuthenticationViewModel *viewmodels.AuthenticationViewModeler
	Renderer                handlerwebapp.Renderer
}

func (h *AuthflowV2VerifyLoginLinkOTPHandler) GetData(r *http.Request, rw http.ResponseWriter) (map[string]any, error) {
	data := make(map[string]any)
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
			// The API rejects an empty code as an invalid request; show it as an invalid link.
			if code == "" {
				finishWithState(LoginLinkOTPPageQueryStateInvalidCode)
				return nil
			}

			err := h.APIController.InspectLoginLink(ctx, map[string]any{"code": code})
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
		if code == "" {
			finishWithState(LoginLinkOTPPageQueryStateInvalidCode)
			return nil
		}

		err = h.APIController.ApproveLoginLink(ctx, map[string]any{"code": code})
		if apierrors.IsKind(err, otp.InvalidOTPCode) {
			finishWithState(LoginLinkOTPPageQueryStateInvalidCode)
			return nil
		} else if err != nil {
			return err
		}

		finishWithState(LoginLinkOTPPageQueryStateMatched)
		return nil
	})
}
