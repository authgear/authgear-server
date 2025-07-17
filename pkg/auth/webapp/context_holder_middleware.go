package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/lib/uiparam"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

var ContextHolderMiddlewareLogger = slogutil.NewLogger("webapp-context-holder-middleware")

type ContextHolderMiddleware struct{}

func (m *ContextHolderMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This middleware only creates the holder of context.
		// This enables the holder to be mutated later in other places.
		var emptyUIParamContext uiparam.T
		ctx := httputil.WithCSPNonce(r.Context(), "")
		ctx = uiparam.WithUIParam(ctx, &emptyUIParamContext)
		ctx = ratelimit.WithRateLimitWeights(ctx)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
