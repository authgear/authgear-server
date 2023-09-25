package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type AuthenticationFlowV1WorkflowService interface {
	CreateNewFlow(intent authflow.PublicFlow, sessionOptions *authflow.SessionOptions) (*authflow.ServiceOutput, error)
	Get(stateToken string) (*authflow.ServiceOutput, error)
	FeedInput(stateToken string, rawMessage json.RawMessage) (*authflow.ServiceOutput, error)
}

type AuthenticationFlowV1CookieManager interface {
	GetCookie(r *http.Request, def *httputil.CookieDef) (*http.Cookie, error)
	ClearCookie(def *httputil.CookieDef) *http.Cookie
	ValueCookie(def *httputil.CookieDef, value string) *http.Cookie
}

func batchInput0(
	service AuthenticationFlowV1WorkflowService,
	w http.ResponseWriter,
	r *http.Request,
	stateToken string,
	rawMessages []json.RawMessage,
) (output *authflow.ServiceOutput, err error) {
	// Collect all cookies
	var cookies []*http.Cookie
	for _, rawMessage := range rawMessages {
		output, err = service.FeedInput(stateToken, rawMessage)
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

func prepareErrorResponse(service AuthenticationFlowV1WorkflowService, stateToken string, flowErr error) (*api.Response, error) {
	output, err := service.Get(stateToken)
	if err != nil {
		return nil, err
	}

	result := output.ToFlowResponse()
	return &api.Response{
		Error:  flowErr,
		Result: result,
	}, nil
}
