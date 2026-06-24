package transport

import (
	"context"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/siteadmin"
	service "github.com/authgear/authgear-server/pkg/siteadmin/service"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureAuditLogGetRoute(route httproute.Route) httproute.Route {
	return route.WithMethods("GET").
		WithPathPattern("/api/v1/audit-logs/:id")
}

type AuditLogGetService interface {
	GetAuditLog(ctx context.Context, id string) (*service.AuditLogEntryDetail, error)
}

type AuditLogGetHandler struct {
	AuditLogGet AuditLogGetService
}

func (h *AuditLogGetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id := httproute.GetParam(r, "id")

	entry, err := h.AuditLogGet.GetAuditLog(r.Context(), id)
	if err != nil {
		writeError(w, r, err)
		return
	}

	a := entryToSiteAdminAuditLog(entry.AuditLogEntry)
	detail := siteadmin.SiteAdminAuditLogDetail{
		ActivityType:  a.ActivityType,
		ActorUserId:   a.ActorUserId,
		AffectedAppId: a.AffectedAppId,
		CreatedAt:     a.CreatedAt,
		Id:            a.Id,
		IpAddress:     a.IpAddress,
		UserAgent:     a.UserAgent,
		Data:          entry.Data,
	}

	SiteAdminAPISuccessResponse{Body: detail}.WriteTo(w)
}
