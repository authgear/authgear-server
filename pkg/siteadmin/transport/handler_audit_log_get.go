package transport

import (
	"context"
	"net/http"

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
	http.NotFound(w, r)
}
