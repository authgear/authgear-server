package api

import (
	"errors"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oauthsession"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oidc"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func ConfigureWorkflowNewRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("POST", "OPTIONS").
		WithPathPattern("/api/v1/workflows")
}

var WorkflowNewRequestSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"additionalProperties": false,
		"properties": {
			"intent": {
				"type": "object",
				"properties": {
					"kind": { "type": "string" },
					"data": { "type": "object" }
				},
				"required": ["kind", "data"]
			},
			"bind_user_agent": { "type": "boolean" }
		},
		"required": ["intent"]
	}
`)

type WorkflowNewRequest struct {
	Intent        workflow.IntentJSON `json:"intent"`
	BindUserAgent *bool               `json:"bind_user_agent"`
}

func (c *WorkflowNewRequest) SetDefaults() {
	if c.BindUserAgent == nil {
		defaultBindUserAgent := true
		c.BindUserAgent = &defaultBindUserAgent
	}
}

type WorkflowNewWorkflowService interface {
	CreateNewWorkflow(intent workflow.Intent, sessionOptions *workflow.SessionOptions) (*workflow.ServiceOutput, error)
}

type WorkflowNewCookieManager interface {
	GetCookie(r *http.Request, def *httputil.CookieDef) (*http.Cookie, error)
	ClearCookie(def *httputil.CookieDef) *http.Cookie
	ValueCookie(def *httputil.CookieDef, value string) *http.Cookie
}

type WorkflowNewOAuthSessionService interface {
	Get(entryID string) (*oauthsession.Entry, error)
}

type WorkflowNewUIInfoResolver interface {
	GetOAuthSessionIDLegacy(req *http.Request, urlQuery string) (string, bool)
	RemoveOAuthSessionID(w http.ResponseWriter, r *http.Request)
	ResolveForUI(r protocol.AuthorizationRequest) (*oidc.UIInfo, error)
}

type WorkflowNewHandler struct {
	JSON           JSONResponseWriter
	Cookies        WorkflowNewCookieManager
	Workflows      WorkflowNewWorkflowService
	OAuthSessions  WorkflowNewOAuthSessionService
	UIInfoResolver WorkflowNewUIInfoResolver
}

func (h *WorkflowNewHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	var request WorkflowNewRequest
	err = httputil.BindJSONBody(r, w, WorkflowNewRequestSchema.Validator(), &request)
	if err != nil {
		h.JSON.WriteResponse(w, &api.Response{Error: err})
		return
	}

	output, err := h.handle(w, r, request)
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

func (h *WorkflowNewHandler) handle(w http.ResponseWriter, r *http.Request, request WorkflowNewRequest) (*workflow.ServiceOutput, error) {
	intent, err := workflow.InstantiateIntentFromPublicRegistry(request.Intent)
	if err != nil {
		return nil, err
	}

	userAgentID := getOrCreateUserAgentID(h.Cookies, w, r)

	var sessionOptionsFromOAuth *workflow.SessionOptions
	if oauthSessionID, ok := h.UIInfoResolver.GetOAuthSessionIDLegacy(r, ""); ok {
		sessionOptionsFromOAuth, err = h.makeSessionOptionsFromOAuth(oauthSessionID)
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
	sessionOptionsFromQuery := h.makeSessionOptionsFromQuery(r)

	// The query overrides the cookie.
	sessionOptions := sessionOptionsFromOAuth.PartiallyMergeFrom(sessionOptionsFromQuery)

	if *request.BindUserAgent {
		sessionOptions.UserAgentID = userAgentID
	}

	output, err := h.Workflows.CreateNewWorkflow(intent, sessionOptions)
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (h *WorkflowNewHandler) makeSessionOptionsFromOAuth(oauthSessionID string) (*workflow.SessionOptions, error) {
	entry, err := h.OAuthSessions.Get(oauthSessionID)
	if err != nil {
		return nil, err
	}
	req := entry.T.AuthorizationRequest

	uiInfo, err := h.UIInfoResolver.ResolveForUI(req)
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

func (h *WorkflowNewHandler) makeSessionOptionsFromQuery(r *http.Request) *workflow.SessionOptions {
	return &workflow.SessionOptions{
		ClientID:  r.FormValue("client_id"),
		State:     r.FormValue("state"),
		XState:    r.FormValue("x_state"),
		UILocales: r.FormValue("ui_locales"),
	}
}
