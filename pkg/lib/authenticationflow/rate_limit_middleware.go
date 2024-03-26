package authenticationflow

import (
	"errors"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type JSONResponseWriter interface {
	WriteResponse(rw http.ResponseWriter, resp *api.Response)
}

type RateLimitMiddleware struct {
	RateLimiter RateLimiter
	RemoteIP    httputil.RemoteIP
	JSON        JSONResponseWriter
	Config      *config.AppConfig
}

const (
	AuthowAPIPerIP ratelimit.BucketName = "AuthflowAPIPerIP"
)

func (m *RateLimitMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		spec := ratelimit.NewBucketSpec(m.Config.AuthenticationFlow.RateLimits.PerIP, AuthowAPIPerIP, string(m.RemoteIP))
		err := m.RateLimiter.Allow(spec)
		if errors.Is(err, ratelimit.ErrRateLimited(spec.Name)) {
			m.JSON.WriteResponse(w, &api.Response{
				Error: apierrors.NewTooManyRequest("Reach Rate Limit"),
			})
		} else if err != nil {
			panic(err)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
