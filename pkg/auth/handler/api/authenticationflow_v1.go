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
	Get(instanceID string) (*authflow.ServiceOutput, error)
	FeedInput(instanceID string, rawMessage json.RawMessage) (*authflow.ServiceOutput, error)
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
	instanceID string,
	rawMessages []json.RawMessage,
) (output *authflow.ServiceOutput, err error) {
	// Collect all cookies
	var cookies []*http.Cookie
	for _, rawMessage := range rawMessages {
		output, err = service.FeedInput(instanceID, rawMessage)
		if err != nil && !errors.Is(err, authflow.ErrEOF) {
			return nil, err
		}

		// Feed the next input to the latest instance.
		instanceID = output.Flow.InstanceID
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

func prepareErrorResponse(service AuthenticationFlowV1WorkflowService, instanceID string, flowErr error) (*api.Response, error) {
	output, err := service.Get(instanceID)
	if err != nil {
		return nil, err
	}

	result := output.ToFlowResponse()
	return &api.Response{
		Error:  flowErr,
		Result: result,
	}, nil
}
