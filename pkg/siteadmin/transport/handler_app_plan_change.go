package transport

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/siteadmin"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func ConfigureAppPlanChangeRoute(route httproute.Route) httproute.Route {
	return route.WithMethods("OPTIONS", "POST").
		WithPathPattern("/api/v1/apps/:appID/plan")
}

var AppPlanChangeRequestSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"plan_name": { "type": "string", "minLength": 1 }
		},
		"required": ["plan_name"]
	}
`)

type AppPlanChangeService interface {
	ChangeAppPlan(ctx context.Context, appID string, planName string) (*siteadmin.App, error)
}

type AppPlanChangeHandler struct {
	Service AppPlanChangeService
}

type AppPlanChangeParams struct {
	AppID string
	siteadmin.ChangeAppPlanRequest
}

func parseAppPlanChangeParams(r *http.Request) (AppPlanChangeParams, error) {
	var body siteadmin.ChangeAppPlanRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return AppPlanChangeParams{}, err
	}

	if err := AppPlanChangeRequestSchema.Validator().ValidateValue(r.Context(), body); err != nil {
		return AppPlanChangeParams{}, err
	}

	return AppPlanChangeParams{
		AppID:                httproute.GetParam(r, "appID"),
		ChangeAppPlanRequest: body,
	}, nil
}

func (h *AppPlanChangeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, err := parseAppPlanChangeParams(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	http.NotFound(w, r)
}
