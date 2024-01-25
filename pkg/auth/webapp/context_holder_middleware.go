package webapp

import (
	"github.com/authgear/authgear-server/pkg/lib/web"
	"github.com/authgear/authgear-server/pkg/util/log"
	"net/http"
)

type ContextHolderMiddlewareLogger struct{ *log.Logger }

func NewContextHolderMiddlewareLogger(lf *log.Factory) ContextHolderMiddlewareLogger {
	return ContextHolderMiddlewareLogger{lf.New("webapp-context-holder-middleware")}
}

type ContextHolderMiddleware struct {
	Logger ContextHolderMiddlewareLogger
}

func (m *ContextHolderMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This middleware only creates the holder of context.
		// This enables the holder to be mutated later in other places.
		var empty web.CSPNonceContextValue
		ctx := web.WithCSPNonce(r.Context(), &empty)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
