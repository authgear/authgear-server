package webapp

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth/config"
	"github.com/skygeario/skygear-server/pkg/core/auth"
)

type ClientIDMiddleware struct {
	Clients []config.OAuthClientConfig
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
	for _, client := range m.Clients {
		if clientID == client.ClientID() {
			// FIXME(config): Remove auth.AccessKey
			//return auth.AccessKey{Client: client}
		}
	}
	return auth.AccessKey{}
}
