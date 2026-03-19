package transport

import (
	"encoding/json"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

func writeError(w http.ResponseWriter, r *http.Request, err error) {
	apiErr := apierrors.AsAPIErrorWithContext(r.Context(), err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(apiErr.Code)
	_ = json.NewEncoder(w).Encode(struct {
		Error *apierrors.APIError `json:"error"`
	}{apiErr})
}
