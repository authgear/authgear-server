package authenticationflow

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type RateLimitMiddleware struct {
	RateLimiter RateLimiter
	RemoteIP    httputil.RemoteIP
	Config      *config.AppConfig
}

const (
	AuthowAPIPerIP ratelimit.BucketName = "AuthflowAPIPerIP"
)

func (m *RateLimitMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		spec := ratelimit.NewBucketSpec("", "", m.Config.AuthenticationFlow.RateLimits.PerIP, AuthowAPIPerIP, string(m.RemoteIP))
		failedReservation, err := m.RateLimiter.Allow(ctx, spec)
		if err != nil {
			panic(err)
		} else if ratelimitErr := failedReservation.Error(); ratelimitErr != nil && ratelimit.IsRateLimitErrorWithBucketName(ratelimitErr, spec.Name) {
			httputil.WriteJSONResponse(ctx, w, &api.Response{
				Error: apierrors.NewTooManyRequest("Reach Rate Limit"),
			})
		} else {

			next.ServeHTTP(w, r)
		}
	})
}
