package transport

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/siteadmin"
	"github.com/authgear/authgear-server/pkg/siteadmin/service"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func ConfigureAppFeatureConfigPreviewRoute(route httproute.Route) httproute.Route {
	return route.WithMethods("OPTIONS", "POST").
		WithPathPattern("/api/v1/apps/:appID/feature-config/preview")
}

var AppFeatureConfigPreviewRequestSchema = validation.NewSimpleSchema(`
{
    "type": "object",
    "properties": { "app_feature_config_yaml": { "type": "string" } },
    "required": ["app_feature_config_yaml"]
}
`)

type AppFeatureConfigPreviewService interface {
	PreviewAppFeatureConfig(ctx context.Context, appID string, rawYAML string) (*service.AppFeatureConfigResult, error)
}

type AppFeatureConfigPreviewHandler struct {
	Service AppFeatureConfigPreviewService
}

type AppFeatureConfigPreviewParams struct {
	AppID string
	siteadmin.PreviewAppFeatureConfigRequest
}

func parseAppFeatureConfigPreviewParams(r *http.Request) (AppFeatureConfigPreviewParams, error) {
	var body siteadmin.PreviewAppFeatureConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return AppFeatureConfigPreviewParams{}, err
	}
	if err := AppFeatureConfigPreviewRequestSchema.Validator().ValidateValue(r.Context(), body); err != nil {
		return AppFeatureConfigPreviewParams{}, err
	}
	return AppFeatureConfigPreviewParams{
		AppID:                          httproute.GetParam(r, "appID"),
		PreviewAppFeatureConfigRequest: body,
	}, nil
}

func (h *AppFeatureConfigPreviewHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	params, err := parseAppFeatureConfigPreviewParams(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	result, err := h.Service.PreviewAppFeatureConfig(r.Context(), params.AppID, params.AppFeatureConfigYaml)
	if err != nil {
		writeError(w, r, err)
		return
	}
	SiteAdminAPISuccessResponse{Body: featureConfigResultToResponse(result)}.WriteTo(w)
}
