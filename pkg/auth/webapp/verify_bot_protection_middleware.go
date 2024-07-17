package webapp

import (
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

type VerifyBotProtectionMiddlewareEndpointsProvider interface {
	ErrorEndpointURL(uiImpl config.UIImplementation) *url.URL
}

type VerifyBotProtectionMiddleware struct {
	Endpoints   VerifyBotProtectionMiddlewareEndpointsProvider
	UIConfig    *config.UIConfig
	Cookies     CookieManager
	ErrorCookie *ErrorCookie
}

func (m *VerifyBotProtectionMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}
