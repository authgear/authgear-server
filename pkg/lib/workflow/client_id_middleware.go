package workflow

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/clientid"
)

type ClientIDMiddleware struct{}

func (m *ClientIDMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This middleware does not know the client ID.
		// What this middleware does is to create the client ID holder.
		// The client ID will be populated by the workflow session.
		emptyClientID := ""
		ctx := clientid.WithClientID(r.Context(), emptyClientID)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
