package transport

import (
	"encoding/json"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/siteadmin"
)

// SiteAdminAPISuccessResponse writes a 200 JSON response with Content-Type set.
// All siteadmin handlers must use this to return success responses.
type SiteAdminAPISuccessResponse struct {
	Body any
}

func (resp SiteAdminAPISuccessResponse) WriteTo(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp.Body)
}

func writeError(w http.ResponseWriter, r *http.Request, err error) {
	apiErr := apierrors.AsAPIErrorWithContext(r.Context(), err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(apiErr.Code)
	_ = json.NewEncoder(w).Encode(siteadmin.ErrorEnvelope{Error: *apiErr})
}
