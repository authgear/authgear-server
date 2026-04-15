package transport

import (
	"context"
	"net/http"
	"time"

	"github.com/authgear/authgear-server/pkg/api/siteadmin"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func ConfigureMessagingUsageRoute(route httproute.Route) httproute.Route {
	return route.WithMethods("OPTIONS", "GET").
		WithPathPattern("/api/v1/apps/:appID/usage/messaging")
}

type MessagingUsageService interface {
	GetMessagingUsage(ctx context.Context, appID string, startDate string, endDate string) (*siteadmin.MessagingUsage, error)
}

type MessagingUsageHandler struct {
	Service MessagingUsageService
}

type MessagingUsageParams struct {
	AppID     string
	StartDate string
	EndDate   string
}

// parseMessagingUsageParams validates that:
//  1. startDate <= endDate
//  2. the range does not exceed 1 year (end <= start.AddDate(1,0,0))
//
// Both checks live here (not in the service) because makeValidationError is
// transport-package-local.
func parseMessagingUsageParams(r *http.Request) (MessagingUsageParams, error) {
	q := r.URL.Query()

	startDate, err := getDateParam(q, "start_date")
	if err != nil {
		return MessagingUsageParams{}, err
	}

	endDate, err := getDateParam(q, "end_date")
	if err != nil {
		return MessagingUsageParams{}, err
	}

	start, _ := time.Parse("2006-01-02", startDate)
	end, _ := time.Parse("2006-01-02", endDate)
	if start.After(end) {
		return MessagingUsageParams{}, makeValidationError(func(ctx *validation.Context) {
			ctx.Child("end_date").EmitError("range", map[string]interface{}{"details": "end_date must not be before start_date"})
		})
	}
	if end.After(start.AddDate(1, 0, 0)) {
		return MessagingUsageParams{}, makeValidationError(func(ctx *validation.Context) {
			ctx.Child("end_date").EmitError("range", map[string]interface{}{"details": "date range must not exceed 1 year"})
		})
	}

	return MessagingUsageParams{
		AppID:     httproute.GetParam(r, "appID"),
		StartDate: startDate,
		EndDate:   endDate,
	}, nil
}

func (h *MessagingUsageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	params, err := parseMessagingUsageParams(r)
	if err != nil {
		writeError(w, r, err)
		return
	}

	usage, err := h.Service.GetMessagingUsage(r.Context(), params.AppID, params.StartDate, params.EndDate)
	if err != nil {
		writeError(w, r, err)
		return
	}

	SiteAdminAPISuccessResponse{Body: usage}.WriteTo(w)
}
