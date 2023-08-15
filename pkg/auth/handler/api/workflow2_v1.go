package api

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oauthsession"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oidc"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func ConfigureWorkflow2V1Route(route httproute.Route) httproute.Route {
	return route.
		WithMethods("POST", "OPTIONS").
		WithPathPattern("/api/v1/workflow2s")
}

var Workflow2V1RequestSchema = validation.NewSimpleSchema(`
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

type Workflow2V1Action string

const (
	Workflow2V1ActionCreate     Workflow2V1Action = "create"
	Workflow2V1ActionInput      Workflow2V1Action = "input"
	Workflow2V1ActionGet        Workflow2V1Action = "get"
	Workflow2V1ActionBatchInput Workflow2V1Action = "batch_input"
)

type Workflow2V1Request struct {
	Action Workflow2V1Action `json:"action"`

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

func (r *Workflow2V1Request) SetDefaults() {
	if r.BindUserAgent == nil {
		defaultBindUserAgent := true
		r.BindUserAgent = &defaultBindUserAgent
	}
}

type Workflow2V1WorkflowService interface {
	CreateNewWorkflow(intent workflow.Intent, sessionOptions *workflow.SessionOptions) (*workflow.ServiceOutput, error)
	Get(workflowID string, instanceID string, userAgentID string) (*workflow.ServiceOutput, error)
	FeedInput(workflowID string, instanceID string, userAgentID string, input workflow.Input) (*workflow.ServiceOutput, error)
}

type Workflow2V1CookieManager interface {
	GetCookie(r *http.Request, def *httputil.CookieDef) (*http.Cookie, error)
	ClearCookie(def *httputil.CookieDef) *http.Cookie
	ValueCookie(def *httputil.CookieDef, value string) *http.Cookie
}

type Workflow2V1OAuthSessionService interface {
	Get(entryID string) (*oauthsession.Entry, error)
}

type Workflow2V1UIInfoResolver interface {
	ResolveForUI(r protocol.AuthorizationRequest) (*oidc.UIInfo, error)
}

type Workflow2V1Handler struct {
	JSON           JSONResponseWriter
	Cookies        Workflow2V1CookieManager
	Workflows      Workflow2V1WorkflowService
	OAuthSessions  Workflow2V1OAuthSessionService
	UIInfoResolver Workflow2V1UIInfoResolver
}

func (h *Workflow2V1Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	var request Workflow2V1Request
	err = httputil.BindJSONBody(r, w, Workflow2V1RequestSchema.Validator(), &request)
	if err != nil {
		h.JSON.WriteResponse(w, &api.Response{Error: err})
		return
	}

	switch request.Action {
	case Workflow2V1ActionCreate:
		output, err := h.create(w, r, request)
		if err != nil {
			h.JSON.WriteResponse(w, &api.Response{Error: err})
			return
		}

		if len(request.BatchInput) > 0 {
			workflowID := output.Workflow.WorkflowID
			instanceID := output.Workflow.InstanceID
			userAgentID := output.Session.UserAgentID

			output, err = h.batchInput(w, r, workflowID, instanceID, userAgentID, request)
			if err != nil {
				apiResp, apiRespErr := h.prepareErrorResponse(workflowID, instanceID, userAgentID, err)
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

		result := Workflow2Response{
			Action:     output.Action,
			WorkflowID: output.Workflow.WorkflowID,
			InstanceID: output.Workflow.InstanceID,
			Data:       output.Data,
		}
		h.JSON.WriteResponse(w, &api.Response{Result: result})
	case Workflow2V1ActionInput:
		workflowID := request.WorkflowID
		instanceID := request.InstanceID
		userAgentID := workflow2getOrCreateUserAgentID(h.Cookies, w, r)

		output, err := h.input(w, r, workflowID, instanceID, userAgentID, request)
		if err != nil {
			apiResp, apiRespErr := h.prepareErrorResponse(workflowID, instanceID, userAgentID, err)
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

		result := Workflow2Response{
			Action:     output.Action,
			WorkflowID: output.Workflow.WorkflowID,
			InstanceID: output.Workflow.InstanceID,
			Data:       output.Data,
		}
		h.JSON.WriteResponse(w, &api.Response{Result: result})
	case Workflow2V1ActionBatchInput:
		workflowID := request.WorkflowID
		instanceID := request.InstanceID
		userAgentID := workflow2getOrCreateUserAgentID(h.Cookies, w, r)

		output, err := h.batchInput(w, r, workflowID, instanceID, userAgentID, request)
		if err != nil {
			apiResp, apiRespErr := h.prepareErrorResponse(workflowID, instanceID, userAgentID, err)
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

		result := Workflow2Response{
			Action:     output.Action,
			WorkflowID: output.Workflow.WorkflowID,
			InstanceID: output.Workflow.InstanceID,
			Data:       output.Data,
		}
		h.JSON.WriteResponse(w, &api.Response{Result: result})
	case Workflow2V1ActionGet:
		workflowID := request.WorkflowID
		instanceID := request.InstanceID
		userAgentID := workflow2getOrCreateUserAgentID(h.Cookies, w, r)

		output, err := h.Workflows.Get(workflowID, instanceID, userAgentID)
		if err != nil {
			h.JSON.WriteResponse(w, &api.Response{Error: err})
			return
		}

		result := Workflow2Response{
			Action:     output.Action,
			WorkflowID: output.Workflow.WorkflowID,
			InstanceID: output.Workflow.InstanceID,
			Data:       output.Data,
		}
		h.JSON.WriteResponse(w, &api.Response{Result: result})
	}
}

func (h *Workflow2V1Handler) create(w http.ResponseWriter, r *http.Request, request Workflow2V1Request) (*workflow.ServiceOutput, error) {
	intent, err := workflow.InstantiateIntentFromPublicRegistry(*request.Intent)
	if err != nil {
		return nil, err
	}

	userAgentID := workflow2getOrCreateUserAgentID(h.Cookies, w, r)

	var sessionOptionsFromCookie *workflow.SessionOptions
	oauthCookie, err := h.Cookies.GetCookie(r, oauthsession.UICookieDef)
	if err == nil {
		sessionOptionsFromCookie, err = h.makeSessionOptionsFromCookie(oauthCookie)
		if errors.Is(err, oauthsession.ErrNotFound) {
			// Clear the cookie if it invalid or expired
			httputil.UpdateCookie(w, h.Cookies.ClearCookie(oauthsession.UICookieDef))
		} else if err != nil {
			// Still return error for any other errors.
			return nil, err
		}

		// Do not clear the UI cookie so that a new session can be created again.
		// httputil.UpdateCookie(w, h.Cookies.ClearCookie(oauthsession.UICookieDef))
	}

	// Accept client_id, state, ui_locales from query.
	// This is essential if the templates of some features require these paramenters.
	sessionOptionsFromQuery := h.makeSessionOptionsFromQuery(request.URLQuery)

	// The query overrides the cookie.
	sessionOptions := sessionOptionsFromCookie.PartiallyMergeFrom(sessionOptionsFromQuery)

	if *request.BindUserAgent {
		sessionOptions.UserAgentID = userAgentID
	}

	output, err := h.Workflows.CreateNewWorkflow(intent, sessionOptions)
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (h *Workflow2V1Handler) makeSessionOptionsFromCookie(oauthSessionCookie *http.Cookie) (*workflow.SessionOptions, error) {
	entry, err := h.OAuthSessions.Get(oauthSessionCookie.Value)
	if err != nil {
		return nil, err
	}
	req := entry.T.AuthorizationRequest

	uiInfo, err := h.UIInfoResolver.ResolveForUI(req)
	if err != nil {
		return nil, err
	}

	sessionOptions := &workflow.SessionOptions{
		ClientID:                 uiInfo.ClientID,
		RedirectURI:              uiInfo.RedirectURI,
		SuppressIDPSessionCookie: uiInfo.SuppressIDPSessionCookie,
		State:                    uiInfo.State,
		XState:                   uiInfo.XState,
		UILocales:                req.UILocalesRaw(),
	}

	return sessionOptions, nil
}

func (h *Workflow2V1Handler) makeSessionOptionsFromQuery(urlQuery string) *workflow.SessionOptions {
	q, _ := url.ParseQuery(urlQuery)
	return &workflow.SessionOptions{
		ClientID:  q.Get("client_id"),
		State:     q.Get("state"),
		XState:    q.Get("x_state"),
		UILocales: q.Get("ui_locales"),
	}
}

func (h *Workflow2V1Handler) input(
	w http.ResponseWriter,
	r *http.Request,
	workflowID string,
	instanceID string,
	userAgentID string,
	request Workflow2V1Request,
) (*workflow.ServiceOutput, error) {
	input, err := workflow.InstantiateInputFromPublicRegistry(*request.Input)
	if err != nil {
		return nil, err
	}

	output, err := h.Workflows.FeedInput(workflowID, instanceID, userAgentID, input)
	if err != nil && errors.Is(err, workflow.ErrNoChange) {
		err = workflow.ErrInvalidInputKind
	}
	if err != nil && !errors.Is(err, workflow.ErrEOF) {
		return nil, err
	}

	return output, nil
}

func (h *Workflow2V1Handler) batchInput(
	w http.ResponseWriter,
	r *http.Request,
	workflowID string,
	instanceID string,
	userAgentID string,
	request Workflow2V1Request,
) (output *workflow.ServiceOutput, err error) {
	// Collect all cookies
	var cookies []*http.Cookie
	var input workflow.Input
	for _, inputJSON := range request.BatchInput {
		input, err = workflow.InstantiateInputFromPublicRegistry(*inputJSON)
		if err != nil {
			return nil, err
		}

		output, err = h.Workflows.FeedInput(workflowID, instanceID, userAgentID, input)
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

func (h *Workflow2V1Handler) prepareErrorResponse(
	workflowID string,
	instanceID string,
	userAgentID string,
	workflowErr error,
) (*api.Response, error) {
	output, err := h.Workflows.Get(workflowID, instanceID, userAgentID)
	if err != nil {
		return nil, err
	}

	result := Workflow2Response{
		Action:     output.Action,
		WorkflowID: output.Workflow.WorkflowID,
		InstanceID: output.Workflow.InstanceID,
		Data:       output.Data,
	}
	return &api.Response{
		Error:  workflowErr,
		Result: result,
	}, nil
}
