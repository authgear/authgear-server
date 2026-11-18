package api

import (
	"context"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

//go:generate go tool mockgen -source=login_link_v1_approve.go -destination=login_link_v1_approve_mock_test.go -package api

func ConfigureLoginLinkV1ApproveRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST").
		WithPathPattern("/api/v1/login_link/approve")
}

var LoginLinkV1ApproveRequestSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"additionalProperties": false,
		"properties": {
			"code": { "type": "string", "minLength": 1 }
		},
		"required": ["code"]
	}
`)

type LoginLinkV1ApproveRequest struct {
	Code string `json:"code"`
}

type LoginLinkV1ApproveOTPCodeService interface {
	LookupCode(ctx context.Context, purpose otp.Purpose, code string) (target string, err error)
	VerifyOTP(ctx context.Context, kind otp.Kind, target string, otp string, opts *otp.VerifyOptions) error
	SetSubmittedCode(ctx context.Context, kind otp.Kind, target string, code string) (*otp.State, error)
}

type LoginLinkV1ApproveEventStore interface {
	Publish(ctx context.Context, websocketChannelName string, e authflow.Event) error
}

var LoginLinkV1ApproveHandlerLogger = slogutil.NewLogger("handler-login-link-v1-approve")

type LoginLinkV1ApproveHandler struct {
	Config   *config.AppConfig
	OTPCodes LoginLinkV1ApproveOTPCodeService
	Events   LoginLinkV1ApproveEventStore
}

func (h *LoginLinkV1ApproveHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var request LoginLinkV1ApproveRequest
	err := httputil.BindJSONBody(r, w, LoginLinkV1ApproveRequestSchema.Validator(), &request)
	if err != nil {
		httputil.WriteJSONResponse(ctx, w, &api.Response{Error: err})
		return
	}

	err = h.approve(ctx, request.Code)
	if err != nil {
		if !apierrors.IsAPIError(err) {
			LoginLinkV1ApproveHandlerLogger.GetLogger(ctx).WithError(err).Error(ctx, "failed to approve login link")
		}
		httputil.WriteJSONResponse(ctx, w, &api.Response{Error: err})
		return
	}

	httputil.WriteJSONResponse(ctx, w, &api.Response{Result: struct{}{}})
}

func (h *LoginLinkV1ApproveHandler) approve(ctx context.Context, code string) error {
	kind := otp.KindOOBOTPLink(h.Config, model.AuthenticatorOOBChannelEmail)

	target, err := h.OTPCodes.LookupCode(ctx, kind.Purpose(), code)
	if err != nil {
		return err
	}

	// Verify without consuming, so that rate limits and failed attempts apply.
	// The flow consumes the code when it completes the step.
	err = h.OTPCodes.VerifyOTP(ctx, kind, target, code, &otp.VerifyOptions{SkipConsume: true})
	if err != nil {
		return err
	}

	state, err := h.OTPCodes.SetSubmittedCode(ctx, kind, target, code)
	if err != nil {
		return err
	}

	if state.AuthenticationFlowWebsocketChannelName != "" {
		err = h.Events.Publish(ctx, state.AuthenticationFlowWebsocketChannelName, authflow.NewEventRefresh())
		if err != nil {
			return err
		}
	}

	return nil
}
