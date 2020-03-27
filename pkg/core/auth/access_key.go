package auth

import (
	"context"
	"crypto/subtle"
	"net/http"
	"strconv"

	"github.com/skygeario/skygear-server/pkg/core/config"
)

const headerAccessKey = "X-Skygear-Api-Key"
const headerIsMasterKey = "X-Skygear-Is-Master-Key"

type AccessKey struct {
	IsMasterKey bool
	Client      config.OAuthClientConfiguration
}

func NewAccessKeyFromRequest(r *http.Request) *AccessKey {
	return &AccessKey{
		IsMasterKey: r.Header.Get(headerIsMasterKey) == "true",
	}
}

func (a AccessKey) WriteTo(rw http.ResponseWriter) {
	rw.Header().Set(headerIsMasterKey, strconv.FormatBool(a.IsMasterKey))
}

type accessKeyContextKeyType struct{}

var accessKeyContextKey = accessKeyContextKeyType{}

func WithAccessKey(ctx context.Context, ak AccessKey) context.Context {
	return context.WithValue(ctx, accessKeyContextKey, ak)
}

func GetAccessKey(ctx context.Context) AccessKey {
	key, _ := ctx.Value(accessKeyContextKey).(AccessKey)
	return key
}

type AccessKeyMiddleware struct {
	TenantConfig *config.TenantConfiguration
}

func (m *AccessKeyMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		accessKey := m.resolve(r)
		r = r.WithContext(WithAccessKey(r.Context(), accessKey))
		next.ServeHTTP(rw, r)
	})
}

func (m *AccessKeyMiddleware) resolve(r *http.Request) AccessKey {
	accessKey := r.Header.Get(headerAccessKey)

	if subtle.ConstantTimeCompare([]byte(accessKey), []byte(m.TenantConfig.AppConfig.MasterKey)) == 1 {
		return AccessKey{IsMasterKey: true}
	}

	for _, clientConfig := range m.TenantConfig.AppConfig.Clients {
		if accessKey == clientConfig.ClientID() {
			return AccessKey{Client: clientConfig}
		}
	}

	return AccessKey{}
}
