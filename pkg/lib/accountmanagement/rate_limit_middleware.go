package accountmanagement

import (
	"context"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type RateLimitMiddlewareRateLimiter interface {
	Allow(ctx context.Context, spec ratelimit.BucketSpec) (*ratelimit.FailedReservation, error)
}

type RateLimitMiddleware struct {
	RateLimiter RateLimitMiddlewareRateLimiter
	RemoteIP    httputil.RemoteIP
}

const (
	AccountManagementAPIPerIP ratelimit.BucketName = "AccountManagementAPIPerIP"
)

var accountManagementAPIPerIPConfigEnabled bool = true
var accountManagementAPIPerIPConfig *config.RateLimitConfig = &config.RateLimitConfig{
	Enabled: &accountManagementAPIPerIPConfigEnabled,
	Period:  "1m",
	Burst:   1200,
}

func (m *RateLimitMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		spec := ratelimit.NewBucketSpec("", "", accountManagementAPIPerIPConfig, AccountManagementAPIPerIP, string(m.RemoteIP))
		failed, err := m.RateLimiter.Allow(ctx, spec)
		if err != nil {
			panic(err)
		} else if ratelimitErr := failed.Error(); ratelimitErr != nil && ratelimit.IsRateLimitErrorWithBucketName(err, spec.Name) {
			httputil.WriteJSONResponse(ctx, w, &api.Response{
				Error: apierrors.NewTooManyRequest("Reach Rate Limit"),
			})
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
