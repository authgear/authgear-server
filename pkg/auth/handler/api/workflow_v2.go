package api

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oauthsession"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oidc"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func ConfigureWorkflowV2Route(route httproute.Route) httproute.Route {
	return route.
		WithMethods("POST", "OPTIONS").
		WithPathPattern("/api/v2/workflows")
}

var WorkflowV2RequestSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"action": {
				"enum": ["create", "input", "batch_input", "get"]
			}
		},
		"required": ["action"],
		"allOf": [
			{
				"if": {
					"properties": {
						"action": { "const": "create" }
					},
					"required": ["action"]
				},
				"then": {
					"properties": {
						"url_query": { "type": "string" },
						"intent": {
							"type": "object",
							"properties": {
								"kind": { "type": "string" },
								"data": { "type": "object" }
							},
							"required": ["kind", "data"]
						},
						"bind_user_agent": { "type": "boolean" },
						"batch_input": {
							"type": "array",
							"items": {
								"type": "object",
								"properties": {
									"kind": { "type": "string" },
									"data": { "type": "object" }
								},
								"required": ["kind", "data"]
							}
						}
					},
					"required": ["intent"]
				}
			},
			{
				"if": {
					"properties": {
						"action": { "const": "input" }
					},
					"required": ["action"]
				},
				"then": {
					"properties": {
						"workflow_id": { "type": "string" },
						"instance_id": { "type": "string" },
						"input": {
							"type": "object",
							"properties": {
								"kind": { "type": "string" },
								"data": { "type": "object" }
							},
							"required": ["kind", "data"]
						}
					},
					"required": ["workflow_id", "instance_id", "input"]
				}
			},
			{
				"if": {
					"properties": {
						"action": { "const": "batch_input" }
					},
					"required": ["action"]
				},
				"then": {
					"properties": {
						"workflow_id": { "type": "string" },
						"instance_id": { "type": "string" },
						"batch_input": {
							"type": "array",
							"items": {
								"type": "object",
								"properties": {
									"kind": { "type": "string" },
									"data": { "type": "object" }
								},
								"required": ["kind", "data"]
							},
							"minItems": 1
						}
					},
					"required": ["workflow_id", "instance_id", "batch_input"]
				}
			},
			{
				"if": {
					"properties": {
						"action": { "const": "get" }
					},
					"required": ["action"]
				},
				"then": {
					"properties": {
						"workflow_id": { "type": "string" },
						"instance_id": { "type": "string" }
					},
					"required": ["workflow_id", "instance_id"]
				}
			}
		]
	}
`)

type WorkflowV2Action string

const (
	WorkflowV2ActionCreate     WorkflowV2Action = "create"
	WorkflowV2ActionInput      WorkflowV2Action = "input"
	WorkflowV2ActionGet        WorkflowV2Action = "get"
	WorkflowV2ActionBatchInput WorkflowV2Action = "batch_input"
)

type WorkflowV2Request struct {
	Action WorkflowV2Action `json:"action"`

	// Create
	URLQuery      string               `json:"url_query,omitempty"`
	Intent        *workflow.IntentJSON `json:"intent,omitempty"`
	BindUserAgent *bool                `json:"bind_user_agent,omitempty"`

	// Input, Get, or BatchInput
	WorkflowID string `json:"workflow_id,omitempty"`
	InstanceID string `json:"instance_id,omitempty"`

	// Input
	Input *workflow.InputJSON `json:"input,omitempty"`

	// BatchInput, or Create
	BatchInput []*workflow.InputJSON `json:"batch_input,omitempty"`
}

func (r *WorkflowV2Request) SetDefaults() {
	if r.BindUserAgent == nil {
		defaultBindUserAgent := true
		r.BindUserAgent = &defaultBindUserAgent
	}
}

type WorkflowV2WorkflowService interface {
	CreateNewWorkflow(ctx context.Context, intent workflow.Intent, sessionOptions *workflow.SessionOptions) (*workflow.ServiceOutput, error)
	Get(ctx context.Context, workflowID string, instanceID string, userAgentID string) (*workflow.ServiceOutput, error)
	FeedInput(ctx context.Context, workflowID string, instanceID string, userAgentID string, input workflow.Input) (*workflow.ServiceOutput, error)
}

type WorkflowV2CookieManager interface {
	GetCookie(r *http.Request, def *httputil.CookieDef) (*http.Cookie, error)
	ClearCookie(def *httputil.CookieDef) *http.Cookie
	ValueCookie(def *httputil.CookieDef, value string) *http.Cookie
}

type WorkflowV2OAuthSessionService interface {
	Get(ctx context.Context, entryID string) (*oauthsession.Entry, error)
}

type WorkflowV2UIInfoResolver interface {
	GetOAuthSessionIDLegacy(req *http.Request, urlQuery string) (string, bool)
	RemoveOAuthSessionID(w http.ResponseWriter, r *http.Request)
	ResolveForUI(ctx context.Context, r protocol.AuthorizationRequest) (*oidc.UIInfo, error)
}

type WorkflowV2Handler struct {
	JSON           JSONResponseWriter
	Cookies        WorkflowV2CookieManager
	Workflows      WorkflowV2WorkflowService
	OAuthSessions  WorkflowV2OAuthSessionService
	UIInfoResolver WorkflowV2UIInfoResolver
}

func (h *WorkflowV2Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	var request WorkflowV2Request
	err = httputil.BindJSONBody(r, w, WorkflowV2RequestSchema.Validator(), &request)
	if err != nil {
		h.JSON.WriteResponse(w, &api.Response{Error: err})
		return
	}

	ctx := r.Context()
	switch request.Action {
	case WorkflowV2ActionCreate:
		h.handleActionCreate(ctx, w, r, request)
		return
	case WorkflowV2ActionInput:
		h.handleActionInput(ctx, w, r, request)
		return
	case WorkflowV2ActionBatchInput:
		h.handleActionBatchInput(ctx, w, r, request)
		return
	case WorkflowV2ActionGet:
		h.handleActionGet(ctx, w, r, request)
		return
	}
}

func (h *WorkflowV2Handler) handleActionCreate(ctx context.Context, w http.ResponseWriter, r *http.Request, request WorkflowV2Request) {
	output, err := h.create(ctx, w, r, request)
	if err != nil {
		h.JSON.WriteResponse(w, &api.Response{Error: err})
		return
	}

	if len(request.BatchInput) > 0 {
		workflowID := output.Workflow.WorkflowID
		instanceID := output.Workflow.InstanceID
		userAgentID := output.Session.UserAgentID

		output, err = h.batchInput(ctx, w, r, workflowID, instanceID, userAgentID, request)
		if err != nil {
			apiResp, apiRespErr := h.prepareErrorResponse(ctx, workflowID, instanceID, userAgentID, err)
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

	result := WorkflowResponse{
		Action:   output.Action,
		Workflow: output.WorkflowOutput,
	}
	h.JSON.WriteResponse(w, &api.Response{Result: result})
}

func (h *WorkflowV2Handler) handleActionInput(ctx context.Context, w http.ResponseWriter, r *http.Request, request WorkflowV2Request) {
	workflowID := request.WorkflowID
	instanceID := request.InstanceID
	userAgentID := getOrCreateUserAgentID(h.Cookies, w, r)

	output, err := h.input(ctx, w, r, workflowID, instanceID, userAgentID, request)
	if err != nil {
		apiResp, apiRespErr := h.prepareErrorResponse(ctx, workflowID, instanceID, userAgentID, err)
		if apiRespErr != nil {
			// failed to get the workflow when preparing the error response
			h.JSON.WriteResponse(w, &api.Response{Error: apiRespErr})
			return
		}
		h.JSON.WriteResponse(w, apiResp)
		return
	}

	for _, c := range output.Cookies {
		httputil.UpdateCookie(w, c)
	}

	result := WorkflowResponse{
		Action:   output.Action,
		Workflow: output.WorkflowOutput,
	}
	h.JSON.WriteResponse(w, &api.Response{Result: result})
}

func (h *WorkflowV2Handler) handleActionBatchInput(ctx context.Context, w http.ResponseWriter, r *http.Request, request WorkflowV2Request) {
	workflowID := request.WorkflowID
	instanceID := request.InstanceID
	userAgentID := getOrCreateUserAgentID(h.Cookies, w, r)

	output, err := h.batchInput(ctx, w, r, workflowID, instanceID, userAgentID, request)
	if err != nil {
		apiResp, apiRespErr := h.prepareErrorResponse(ctx, workflowID, instanceID, userAgentID, err)
		if apiRespErr != nil {
			// failed to get the workflow when preparing the error response
			h.JSON.WriteResponse(w, &api.Response{Error: apiRespErr})
			return
		}
		h.JSON.WriteResponse(w, apiResp)
		return
	}

	for _, c := range output.Cookies {
		httputil.UpdateCookie(w, c)
	}

	result := WorkflowResponse{
		Action:   output.Action,
		Workflow: output.WorkflowOutput,
	}
	h.JSON.WriteResponse(w, &api.Response{Result: result})
}

func (h *WorkflowV2Handler) handleActionGet(ctx context.Context, w http.ResponseWriter, r *http.Request, request WorkflowV2Request) {
	workflowID := request.WorkflowID
	instanceID := request.InstanceID
	userAgentID := getOrCreateUserAgentID(h.Cookies, w, r)

	output, err := h.Workflows.Get(ctx, workflowID, instanceID, userAgentID)
	if err != nil {
		h.JSON.WriteResponse(w, &api.Response{Error: err})
		return
	}

	result := WorkflowResponse{
		Action:   output.Action,
		Workflow: output.WorkflowOutput,
	}
	h.JSON.WriteResponse(w, &api.Response{Result: result})
}

func (h *WorkflowV2Handler) create(ctx context.Context, w http.ResponseWriter, r *http.Request, request WorkflowV2Request) (*workflow.ServiceOutput, error) {
	intent, err := workflow.InstantiateIntentFromPublicRegistry(*request.Intent)
	if err != nil {
		return nil, err
	}

	userAgentID := getOrCreateUserAgentID(h.Cookies, w, r)

	var sessionOptionsFromOAuth *workflow.SessionOptions
	if oauthSessionID, ok := h.UIInfoResolver.GetOAuthSessionIDLegacy(r, request.URLQuery); ok {
		sessionOptionsFromOAuth, err = h.makeSessionOptionsFromOAuth(ctx, oauthSessionID)
		if errors.Is(err, oauthsession.ErrNotFound) {
			// Clear the oauth session if it invalid or expired
			h.UIInfoResolver.RemoveOAuthSessionID(w, r)
		} else if err != nil {
			// Still return error for any other errors.
			return nil, err
		}

		// Do not clear oauth session so that a new session can be created again.
	}

	// Accept client_id, state, ui_locales from query.
	// This is essential if the templates of some features require these paramenters.
	sessionOptionsFromQuery := h.makeSessionOptionsFromQuery(request.URLQuery)

	// The query overrides the cookie.
	sessionOptions := sessionOptionsFromOAuth.PartiallyMergeFrom(sessionOptionsFromQuery)

	if *request.BindUserAgent {
		sessionOptions.UserAgentID = userAgentID
	}

	output, err := h.Workflows.CreateNewWorkflow(ctx, intent, sessionOptions)
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (h *WorkflowV2Handler) makeSessionOptionsFromOAuth(ctx context.Context, oauthSessionID string) (*workflow.SessionOptions, error) {
	entry, err := h.OAuthSessions.Get(ctx, oauthSessionID)
	if err != nil {
		return nil, err
	}
	req := entry.T.AuthorizationRequest

	uiInfo, err := h.UIInfoResolver.ResolveForUI(ctx, req)
	if err != nil {
		return nil, err
	}

	sessionOptions := &workflow.SessionOptions{
		OAuthSessionID:           oauthSessionID,
		ClientID:                 uiInfo.ClientID,
		RedirectURI:              uiInfo.RedirectURI,
		SuppressIDPSessionCookie: uiInfo.SuppressIDPSessionCookie,
		State:                    uiInfo.State,
		XState:                   uiInfo.XState,
		UserIDHint:               uiInfo.UserIDHint,
		UILocales:                req.UILocalesRaw(),
	}

	return sessionOptions, nil
}

func (h *WorkflowV2Handler) makeSessionOptionsFromQuery(urlQuery string) *workflow.SessionOptions {
	q, _ := url.ParseQuery(urlQuery)
	return &workflow.SessionOptions{
		ClientID:  q.Get("client_id"),
		State:     q.Get("state"),
		XState:    q.Get("x_state"),
		UILocales: q.Get("ui_locales"),
	}
}

func (h *WorkflowV2Handler) input(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	workflowID string,
	instanceID string,
	userAgentID string,
	request WorkflowV2Request,
) (*workflow.ServiceOutput, error) {
	input, err := workflow.InstantiateInputFromPublicRegistry(*request.Input)
	if err != nil {
		return nil, err
	}

	output, err := h.Workflows.FeedInput(ctx, workflowID, instanceID, userAgentID, input)
	if err != nil && errors.Is(err, workflow.ErrNoChange) {
		err = workflow.ErrInvalidInputKind
	}
	if err != nil && !errors.Is(err, workflow.ErrEOF) {
		return nil, err
	}

	return output, nil
}

func (h *WorkflowV2Handler) batchInput(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	workflowID string,
	instanceID string,
	userAgentID string,
	request WorkflowV2Request,
) (output *workflow.ServiceOutput, err error) {
	// Collect all cookies
	var cookies []*http.Cookie
	var input workflow.Input
	for _, inputJSON := range request.BatchInput {
		input, err = workflow.InstantiateInputFromPublicRegistry(*inputJSON)
		if err != nil {
			return nil, err
		}

		output, err = h.Workflows.FeedInput(ctx, workflowID, instanceID, userAgentID, input)
		if err != nil && errors.Is(err, workflow.ErrNoChange) {
			err = workflow.ErrInvalidInputKind
		}
		if err != nil && !errors.Is(err, workflow.ErrEOF) {
			return nil, err
		}

		// Feed the next input to the latest instance.
		instanceID = output.Workflow.InstanceID
		cookies = append(cookies, output.Cookies...)
	}
	if err != nil && errors.Is(err, workflow.ErrEOF) {
		err = nil
	}
	if err != nil {
		return
	}

	// Return all collected cookies.
	output.Cookies = cookies
	return
}

func (h *WorkflowV2Handler) prepareErrorResponse(
	ctx context.Context,
	workflowID string,
	instanceID string,
	userAgentID string,
	workflowErr error,
) (*api.Response, error) {
	output, err := h.Workflows.Get(ctx, workflowID, instanceID, userAgentID)
	if err != nil {
		return nil, err
	}

	result := WorkflowResponse{
		Action:   output.Action,
		Workflow: output.WorkflowOutput,
	}
	return &api.Response{
		Error:  workflowErr,
		Result: result,
	}, nil
}
