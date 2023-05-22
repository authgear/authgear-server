package api

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func ConfigureWorkflowGetRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("GET").
		WithPathPattern("/api/v1/workflows/:workflowid/instances/:instanceid")
}

type WorkflowGetWorkflowService interface {
	Get(workflowID string, instanceID string, userAgentID string) (*workflow.ServiceOutput, error)
}

type WorkflowGetCookieManager interface {
	GetCookie(r *http.Request, def *httputil.CookieDef) (*http.Cookie, error)
	ValueCookie(def *httputil.CookieDef, value string) *http.Cookie
}

type WorkflowGetHandler struct {
	JSON      JSONResponseWriter
	Workflows WorkflowGetWorkflowService
	Cookies   WorkflowGetCookieManager
}

func (h *WorkflowGetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	workflowID := httproute.GetParam(r, "workflowid")
	instanceID := httproute.GetParam(r, "instanceid")

	userAgentID := getOrCreateUserAgentID(h.Cookies, w, r)

	output, err := h.Workflows.Get(workflowID, instanceID, userAgentID)
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
