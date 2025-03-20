package proofofphonenumberverification

import (
	"context"
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/hook"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type WebhookMiddlewareLogger struct{ *log.Logger }

func NewWebhookMiddlewareLogger(lf *log.Factory) WebhookMiddlewareLogger {
	return WebhookMiddlewareLogger{lf.New("proof-of-phone-number-verification-webhook")}
}

type ProofOfPhoneNumberVerificationWebHook struct {
	hook.WebHook
	Client HookHTTPClient
	Logger WebhookMiddlewareLogger
}

func (h *ProofOfPhoneNumberVerificationWebHook) Call(ctx context.Context, u *url.URL, hookReq *HookRequest) (*HookResponse, error) {
	req, err := h.PrepareRequest(ctx, u, hookReq)
	if err != nil {
		return nil, err
	}

	resp, err := h.PerformWithResponse(h.Client.Client, req)
	defer func() {
		if resp != nil {
			resp.Body.Close()
		}
	}()

	if err != nil {
		h.Logger.WithError(err).Error("failed to call webhook")
		return nil, err
	}

	var hookResp *HookResponse
	hookResp, err = ParseHookResponse(ctx, resp.Body)
	if err != nil {
		apiError := apierrors.AsAPIError(err)
		err = hook.WebHookInvalidResponse.NewWithInfo("invalid response body", apiError.Info_ReadOnly)
		return nil, err
	}
	return hookResp, nil
}

var _ Hook = &ProofOfPhoneNumberVerificationWebHook{}

type HookHTTPClient struct {
	*http.Client
}

func NewHookHTTPClient(cfg *config.ProofOfPhoneNumberVerificationHookConfig) HookHTTPClient {
	return HookHTTPClient{
		httputil.NewExternalClient(cfg.Timeout.Duration()),
	}
}
