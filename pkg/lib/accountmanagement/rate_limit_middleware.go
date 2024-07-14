package accountmanagement

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type RateLimitMiddlewareJSONResponseWriter interface {
	WriteResponse(rw http.ResponseWriter, resp *api.Response)
}

type RateLimitMiddlewareRateLimiter interface {
	Allow(spec ratelimit.BucketSpec) error
}

type RateLimitMiddleware struct {
	RateLimiter RateLimitMiddlewareRateLimiter
	RemoteIP    httputil.RemoteIP
	JSON        RateLimitMiddlewareJSONResponseWriter
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
		spec := ratelimit.NewBucketSpec(accountManagementAPIPerIPConfig, AccountManagementAPIPerIP, string(m.RemoteIP))
		err := m.RateLimiter.Allow(spec)
		if ratelimit.IsRateLimitErrorWithBucketName(err, spec.Name) {
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
