package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/api"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oauthsession"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oidc"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func ConfigureAuthenticationFlowV1CreateRoute(route httproute.Route) httproute.Route {
	return route.WithMethods("OPTIONS", "POST").WithPathPattern("/api/v1/authentication_flows")
}

var AuthenticationFlowV1NonRestfulCreateRequestSchema = validation.NewSimpleSchema(`
{
	"type": "object",
	"required": ["flow_type", "flow_name"],
	"properties": {
		"flow_type": {
			"type": "string",
			"enum": ["signup_flow", "login_flow"]
		},
		"flow_name": {
			"type": "string"
		},
		"url_query": { "type": "string" },
		"batch_input": {
			"type": "array",
			"items": {
				"type": "object"
			}
		}
	}
}
`)

type AuthenticationFlowV1NonRestfulCreateRequest struct {
	FlowType   authflow.FlowType `json:"flow_type,omitempty"`
	FlowName   string            `json:"flow_name,omitempty"`
	URLQuery   string            `json:"url_query,omitempty"`
	BatchInput []json.RawMessage `json:"batch_input,omitempty"`
}

func (r *AuthenticationFlowV1NonRestfulCreateRequest) GetFlowReference() *authflow.FlowReference {
	return &authflow.FlowReference{
		Type: r.FlowType,
		Name: r.FlowName,
	}
}

type AuthenticationFlowV1OAuthSessionService interface {
	Get(entryID string) (*oauthsession.Entry, error)
}

type AuthenticationFlowV1UIInfoResolver interface {
	GetOAuthSessionID(req *http.Request, urlQuery string) (string, bool)
	RemoveOAuthSessionID(w http.ResponseWriter, r *http.Request)
	ResolveForUI(r protocol.AuthorizationRequest) (*oidc.UIInfo, error)
}

type AuthenticationFlowV1CreateHandler struct {
	LoggerFactory  *log.Factory
	RedisHandle    *appredis.Handle
	JSON           JSONResponseWriter
	Cookies        AuthenticationFlowV1CookieManager
	Workflows      AuthenticationFlowV1WorkflowService
	OAuthSessions  AuthenticationFlowV1OAuthSessionService
	UIInfoResolver AuthenticationFlowV1UIInfoResolver
}

func (h *AuthenticationFlowV1CreateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	var request AuthenticationFlowV1NonRestfulCreateRequest
	err = httputil.BindJSONBody(r, w, AuthenticationFlowV1NonRestfulCreateRequestSchema.Validator(), &request)
	if err != nil {
		h.JSON.WriteResponse(w, &api.Response{Error: err})
		return
	}

	h.create(w, r, request)
}

func (h *AuthenticationFlowV1CreateHandler) create(w http.ResponseWriter, r *http.Request, request AuthenticationFlowV1NonRestfulCreateRequest) {
	output, err := h.create0(w, r, request)
	if err != nil {
		h.JSON.WriteResponse(w, &api.Response{Error: err})
		return
	}

	if len(request.BatchInput) > 0 {
		instanceID := output.Flow.InstanceID

		output, err = batchInput0(h.Workflows, w, r, instanceID, request.BatchInput)
		if err != nil {
			apiResp, apiRespErr := prepareErrorResponse(h.Workflows, instanceID, err)
			if apiRespErr != nil {
				// failed to get the workflow when preparing the error response
				h.JSON.WriteResponse(w, &api.Response{Error: apiRespErr})
				return
			}
			h.JSON.WriteResponse(w, apiResp)
			return
		}
	}

	for _, c := range output.Cookies {
		httputil.UpdateCookie(w, c)
	}

	result := output.ToFlowResponse()
	h.JSON.WriteResponse(w, &api.Response{Result: result})
}

func (h *AuthenticationFlowV1CreateHandler) create0(w http.ResponseWriter, r *http.Request, request AuthenticationFlowV1NonRestfulCreateRequest) (*authflow.ServiceOutput, error) {
	flow, err := authflow.InstantiateFlow(*request.GetFlowReference())
	if err != nil {
		return nil, err
	}

	var sessionOptionsFromOAuth *authflow.SessionOptions
	if oauthSessionID, ok := h.UIInfoResolver.GetOAuthSessionID(r, request.URLQuery); ok {
		sessionOptionsFromOAuth, err = h.makeSessionOptionsFromOAuth(oauthSessionID)
		if errors.Is(err, oauthsession.ErrNotFound) {
			// Clear the oauth session if it invalid or expired
			h.UIInfoResolver.RemoveOAuthSessionID(w, r)
		} else if err != nil {
			// Still return error for any other errors.
			return nil, err
		}

		// Do not clear the oauth session so that a new session can be created again.
	}

	// Accept client_id, state, ui_locales from query.
	// This is essential if the templates of some features require these paramenters.
	sessionOptionsFromQuery := h.makeSessionOptionsFromQuery(request.URLQuery)

	// The query overrides the cookie.
	sessionOptions := sessionOptionsFromOAuth.PartiallyMergeFrom(sessionOptionsFromQuery)

	output, err := h.Workflows.CreateNewFlow(flow, sessionOptions)
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (h *AuthenticationFlowV1CreateHandler) makeSessionOptionsFromQuery(urlQuery string) *authflow.SessionOptions {
	q, _ := url.ParseQuery(urlQuery)
	return &authflow.SessionOptions{
		ClientID:  q.Get("client_id"),
		State:     q.Get("state"),
		XState:    q.Get("x_state"),
		UILocales: q.Get("ui_locales"),
	}
}

func (h *AuthenticationFlowV1CreateHandler) makeSessionOptionsFromOAuth(oauthSessionID string) (*authflow.SessionOptions, error) {
	entry, err := h.OAuthSessions.Get(oauthSessionID)
	if err != nil {
		return nil, err
	}
	req := entry.T.AuthorizationRequest

	uiInfo, err := h.UIInfoResolver.ResolveForUI(req)
	if err != nil {
		return nil, err
	}

	sessionOptions := &authflow.SessionOptions{
		OAuthSessionID:           oauthSessionID,
		ClientID:                 uiInfo.ClientID,
		RedirectURI:              uiInfo.RedirectURI,
		Prompt:                   uiInfo.Prompt,
		SuppressIDPSessionCookie: uiInfo.SuppressIDPSessionCookie,
		State:                    uiInfo.State,
		XState:                   uiInfo.XState,
		UILocales:                req.UILocalesRaw(),
	}

	return sessionOptions, nil
}
