package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type AuthenticationFlowV1WorkflowService interface {
	CreateNewFlow(ctx context.Context, intent authflow.PublicFlow, sessionOptions *authflow.SessionOptions) (*authflow.ServiceOutput, error)
	Get(ctx context.Context, stateToken string) (*authflow.ServiceOutput, error)
	FeedInput(ctx context.Context, stateToken string, rawMessage json.RawMessage) (*authflow.ServiceOutput, error)
}

type AuthenticationFlowV1CookieManager interface {
	GetCookie(r *http.Request, def *httputil.CookieDef) (*http.Cookie, error)
	ClearCookie(def *httputil.CookieDef) *http.Cookie
	ValueCookie(def *httputil.CookieDef, value string) *http.Cookie
}

func batchInput0(
	ctx context.Context,
	service AuthenticationFlowV1WorkflowService,
	w http.ResponseWriter,
	r *http.Request,
	stateToken string,
	rawMessages []json.RawMessage,
) (output *authflow.ServiceOutput, err error) {
	// Collect all cookies
	var cookies []*http.Cookie
	for _, rawMessage := range rawMessages {
		output, err = service.FeedInput(ctx, stateToken, rawMessage)
		if err != nil && !errors.Is(err, authflow.ErrEOF) {
			return nil, err
		}

		// Feed the next input to the latest state.
		stateToken = output.Flow.StateToken
		cookies = append(cookies, output.Cookies...)
	}
	if err != nil && errors.Is(err, authflow.ErrEOF) {
		err = nil
	}
	if err != nil {
		return
	}

	// Return all collected cookies.
	output.Cookies = cookies
	return
}

func prepareErrorResponse(ctx context.Context, service AuthenticationFlowV1WorkflowService, stateToken string, flowErr error) (*api.Response, error) {
	// In reset password, it is possible that stateToken is empty
	// Simply return the error we encountered
	if stateToken == "" {
		return &api.Response{
			Error: flowErr,
		}, nil
	}

	// If the error is flow not found, we do not try to get the flow
	if errors.Is(flowErr, authflow.ErrFlowNotFound) {
		return &api.Response{
			Error: flowErr,
		}, nil
	}

	output, err := service.Get(ctx, stateToken)
	if err != nil {
		// This is unknown error because we already checked the error we encountered is not FlowNotFound
		return nil, errors.Join(
			apierrors.NewInternalError("unknown error when trying to handle flow error"),
			err,
			flowErr,
		)
	}

	result := output.ToFlowResponse()
	return &api.Response{
		Error:  flowErr,
		Result: result,
	}, nil
}
