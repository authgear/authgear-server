package webapp

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

type ClientIDMiddleware struct {
	TenantConfig *config.TenantConfiguration
}

func (m *ClientIDMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accessKey := m.resolve(r)
		r = r.WithContext(auth.WithAccessKey(r.Context(), accessKey))
		next.ServeHTTP(w, r)
	})
}

func (m *ClientIDMiddleware) resolve(r *http.Request) auth.AccessKey {
	clientID := r.URL.Query().Get("client_id")
	for _, client := range m.TenantConfig.AppConfig.Clients {
		if clientID == client.ClientID() {
			return auth.AccessKey{Client: client}
		}
	}
	return auth.AccessKey{}
}
