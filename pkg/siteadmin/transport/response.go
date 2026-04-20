package transport

import (
	"encoding/json"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/siteadmin"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

var JSONResponseWriterLogger = slogutil.NewLogger("siteadmin-json-response-writer")

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
	ctx := r.Context()
	apiErr := apierrors.AsAPIErrorWithContext(ctx, err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(apiErr.Code)
	if e := json.NewEncoder(w).Encode(siteadmin.ErrorEnvelope{Error: *apiErr}); e != nil {
		panic(e)
	}

	if apiErr != nil && apiErr.Code >= 500 && apiErr.Code < 600 {
		logger := JSONResponseWriterLogger.GetLogger(ctx)
		logger.WithError(err).Error(ctx, "unexpected error occurred")
	}
}
