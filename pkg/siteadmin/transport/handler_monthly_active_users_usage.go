package transport

import (
	"context"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/siteadmin"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func ConfigureMonthlyActiveUsersUsageRoute(route httproute.Route) httproute.Route {
	return route.WithMethods("OPTIONS", "GET").
		WithPathPattern("/api/v1/apps/:appID/usage/monthly-active-users")
}

type MonthlyActiveUsersUsageService interface {
	GetMonthlyActiveUsersUsage(ctx context.Context, appID string, startYear int, startMonth int, endYear int, endMonth int) (*siteadmin.MonthlyActiveUsersUsage, error)
}

type MonthlyActiveUsersUsageHandler struct {
	Service MonthlyActiveUsersUsageService
}

type MonthlyActiveUsersUsageParams struct {
	AppID      string
	StartYear  int
	StartMonth int
	EndYear    int
	EndMonth   int
}

func parseMonthlyActiveUsersUsageParams(r *http.Request) (MonthlyActiveUsersUsageParams, error) {
	q := r.URL.Query()

	startYear, err := getIntParam(q, "start_year")
	if err != nil {
		return MonthlyActiveUsersUsageParams{}, err
	}

	startMonth, err := getIntParam(q, "start_month")
	if err != nil {
		return MonthlyActiveUsersUsageParams{}, err
	}
	if err := validateMonth("start_month", startMonth); err != nil {
		return MonthlyActiveUsersUsageParams{}, err
	}

	endYear, err := getIntParam(q, "end_year")
	if err != nil {
		return MonthlyActiveUsersUsageParams{}, err
	}

	endMonth, err := getIntParam(q, "end_month")
	if err != nil {
		return MonthlyActiveUsersUsageParams{}, err
	}
	if err := validateMonth("end_month", endMonth); err != nil {
		return MonthlyActiveUsersUsageParams{}, err
	}

	totalMonths := (endYear-startYear)*12 + (endMonth - startMonth)
	if totalMonths < 0 || totalMonths > 11 {
		return MonthlyActiveUsersUsageParams{}, makeValidationError(func(ctx *validation.Context) {
			ctx.EmitError("range", map[string]interface{}{"details": "date range must be within 1 year"})
		})
	}

	return MonthlyActiveUsersUsageParams{
		AppID:      httproute.GetParam(r, "appID"),
		StartYear:  startYear,
		StartMonth: startMonth,
		EndYear:    endYear,
		EndMonth:   endMonth,
	}, nil
}

func (h *MonthlyActiveUsersUsageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	params, err := parseMonthlyActiveUsersUsageParams(r)
	if err != nil {
		writeError(w, r, err)
		return
	}

	usage, err := h.Service.GetMonthlyActiveUsersUsage(
		r.Context(), params.AppID,
		params.StartYear, params.StartMonth,
		params.EndYear, params.EndMonth,
	)
	if err != nil {
		writeError(w, r, err)
		return
	}

	SiteAdminAPISuccessResponse{Body: usage}.WriteTo(w)
}
