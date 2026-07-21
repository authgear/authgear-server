package transport

import (
	"context"
	"net/http"
	"strconv"

	"github.com/authgear/authgear-server/pkg/api/siteadmin"
	service "github.com/authgear/authgear-server/pkg/siteadmin/service"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureAuditLogsListRoute(route httproute.Route) httproute.Route {
	return route.WithMethods("OPTIONS", "GET").
		WithPathPattern("/api/v1/audit-logs")
}

type AuditLogsListService interface {
	ListAuditLogs(ctx context.Context, params service.ListAuditLogsParams) (*service.ListAuditLogsResult, error)
}

type AuditLogsListHandler struct {
	AuditLogsList AuditLogsListService
}

type auditLogsListParams struct {
	Page          uint64
	PageSize      uint64
	AffectedAppID string
	Order         siteadmin.OrderDirection
}

func parseAuditLogsListParams(r *http.Request) auditLogsListParams {
	q := r.URL.Query()

	page := uint64(1)
	if v := q.Get("page"); v != "" {
		if n, err := strconv.ParseUint(v, 10, 64); err == nil && n >= 1 {
			page = n
		}
	}

	pageSize := uint64(20)
	if v := q.Get("page_size"); v != "" {
		if n, err := strconv.ParseUint(v, 10, 64); err == nil && n >= 1 {
			pageSize = min(n, service.AuditLogsMaxPageSize)
		}
	}

	order := siteadmin.OrderDirection(q.Get("order"))

	return auditLogsListParams{
		Page:          page,
		PageSize:      pageSize,
		AffectedAppID: q.Get("affected_app_id"),
		Order:         order,
	}
}

func entryToSiteAdminAuditLog(e service.AuditLogEntry) siteadmin.SiteAdminAuditLog {
	a := siteadmin.SiteAdminAuditLog{
		Id:           e.ID,
		CreatedAt:    e.CreatedAt,
		ActivityType: e.ActivityType,
	}
	if e.IPAddress != "" {
		a.IpAddress = &e.IPAddress
	}
	if e.UserAgent != "" {
		a.UserAgent = &e.UserAgent
	}
	if e.ActorUserID != "" {
		a.ActorUserId = &e.ActorUserID
	}
	if e.AffectedAppID != "" {
		a.AffectedAppId = &e.AffectedAppID
	}
	return a
}

func (h *AuditLogsListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	params := parseAuditLogsListParams(r)

	result, err := h.AuditLogsList.ListAuditLogs(r.Context(), service.ListAuditLogsParams{
		Page:          params.Page,
		PageSize:      params.PageSize,
		AffectedAppID: params.AffectedAppID,
		Order:         params.Order,
	})
	if err != nil {
		writeError(w, r, err)
		return
	}

	entries := make([]siteadmin.SiteAdminAuditLog, len(result.Entries))
	for i, e := range result.Entries {
		entries[i] = entryToSiteAdminAuditLog(e)
	}

	SiteAdminAPISuccessResponse{Body: siteadmin.SiteAdminAuditLogsListResponse{
		AuditLogs:  entries,
		TotalCount: result.TotalCount,
		Page:       params.Page,
		PageSize:   params.PageSize,
	}}.WriteTo(w)
}
