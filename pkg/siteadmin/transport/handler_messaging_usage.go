package transport

import (
	"encoding/json"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/siteadmin"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureMessagingUsageRoute(route httproute.Route) httproute.Route {
	return route.WithMethods("GET").
		WithPathPattern("/api/v1/projects/:projectID/usage/messaging")
}

type MessagingUsageHandler struct {
	// Add service dependencies here as needed
}

type MessagingUsageParams struct {
	ProjectID string
	StartDate string
	EndDate   string
}

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

	return MessagingUsageParams{
		ProjectID: httproute.GetParam(r, "projectID"),
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

	// TODO: Replace with real data source. Return dummy data for now.
	usage := siteadmin.MessagingUsage{
		StartDate:                 params.StartDate,
		EndDate:                   params.EndDate,
		SmsNorthAmericaCount:      120,
		SmsOtherRegionsCount:      45,
		WhatsappNorthAmericaCount: 30,
		WhatsappOtherRegionsCount: 15,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(usage)
}
