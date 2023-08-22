package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oauthsession"
	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func ConfigureWorkflow2V1CreateRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("POST", "OPTIONS").
		WithPathPattern("/api/v1/workflow2s/create")
}

var Workflow2V1CreateRequestSchema = validation.NewSimpleSchema(`
{
	"type": "object",
	"required": ["flow_reference"],
	"properties": {
		"flow_reference": {
			"type": "object",
			"properties": {
				"type": {
					"type": "string",
					"enum": ["signup_flow", "login_flow"]
				},
				"id": {
					"type": "string"
				}
			},
			"required": ["type", "id"]
		},
		"url_query": { "type": "string" },
		"bind_user_agent": { "type": "boolean" },
		"batch_input": {
			"type": "array",
			"items": {
				"type": "object"
			}
		}
	}
}
`)

type Workflow2V1CreateRequest struct {
	FlowReference *workflow.FlowReference `json:"flow_reference,omitempty"`
	URLQuery      string                  `json:"url_query,omitempty"`
	BindUserAgent *bool                   `json:"bind_user_agent,omitempty"`
	BatchInput    []json.RawMessage       `json:"batch_input,omitempty"`
}

func (r *Workflow2V1CreateRequest) SetDefaults() {
	if r.BindUserAgent == nil {
		defaultBindUserAgent := true
		r.BindUserAgent = &defaultBindUserAgent
	}
}

type Workflow2V1CreateHandler struct {
	JSON           JSONResponseWriter
	Cookies        Workflow2V1CookieManager
	Workflows      Workflow2V1WorkflowService
	OAuthSessions  Workflow2V1OAuthSessionService
	UIInfoResolver Workflow2V1UIInfoResolver
}

func (h *Workflow2V1CreateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	var request Workflow2V1CreateRequest
	err = httputil.BindJSONBody(r, w, Workflow2V1CreateRequestSchema.Validator(), &request)
	if err != nil {
		h.JSON.WriteResponse(w, &api.Response{Error: err})
		return
	}

	output, err := h.create(w, r, request)
	if err != nil {
		h.JSON.WriteResponse(w, &api.Response{Error: err})
		return
	}

	if len(request.BatchInput) > 0 {
		instanceID := output.Workflow.InstanceID
		userAgentID := output.Session.UserAgentID

		output, err = h.batchInput(w, r, instanceID, userAgentID, request)
		if err != nil {
			apiResp, apiRespErr := h.prepareErrorResponse(instanceID, userAgentID, err)
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

	result := workflow.FlowResponse{
		ID:         output.Workflow.InstanceID,
		Data:       output.Data,
		Finished:   output.Finished,
		JSONSchema: output.SchemaBuilder,
	}
	h.JSON.WriteResponse(w, &api.Response{Result: result})
}

func (h *Workflow2V1CreateHandler) create(w http.ResponseWriter, r *http.Request, request Workflow2V1CreateRequest) (*workflow.ServiceOutput, error) {
	flow, err := workflow.InstantiateFlow(*request.FlowReference)
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

	output, err := h.Workflows.CreateNewWorkflow(flow, sessionOptions)
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (h *Workflow2V1CreateHandler) makeSessionOptionsFromCookie(oauthSessionCookie *http.Cookie) (*workflow.SessionOptions, error) {
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

func (h *Workflow2V1CreateHandler) makeSessionOptionsFromQuery(urlQuery string) *workflow.SessionOptions {
	q, _ := url.ParseQuery(urlQuery)
	return &workflow.SessionOptions{
		ClientID:  q.Get("client_id"),
		State:     q.Get("state"),
		XState:    q.Get("x_state"),
		UILocales: q.Get("ui_locales"),
	}
}

func (h *Workflow2V1CreateHandler) batchInput(
	w http.ResponseWriter,
	r *http.Request,
	instanceID string,
	userAgentID string,
	request Workflow2V1CreateRequest,
) (output *workflow.ServiceOutput, err error) {
	// Collect all cookies
	var cookies []*http.Cookie
	for _, rawMessage := range request.BatchInput {
		output, err = h.Workflows.FeedInput(instanceID, userAgentID, rawMessage)
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

func (h *Workflow2V1CreateHandler) prepareErrorResponse(
	instanceID string,
	userAgentID string,
	workflowErr error,
) (*api.Response, error) {
	output, err := h.Workflows.Get(instanceID, userAgentID)
	if err != nil {
		return nil, err
	}

	result := workflow.FlowResponse{
		ID:         output.Workflow.InstanceID,
		Data:       output.Data,
		Finished: output.Finished,
		JSONSchema: output.SchemaBuilder,
	}
	return &api.Response{
		Error:  workflowErr,
		Result: result,
	}, nil
}
