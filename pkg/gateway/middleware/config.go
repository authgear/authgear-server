package middleware

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/gateway/model"
)

var configMap map[string]model.Config

func init() {
	configMap = map[string]model.Config{
		"skygear": model.Config{
			Auth: model.AuthConfig{
				PasswordLength: 10,
			},
		},
		"skygear-next": model.Config{
			Auth: model.AuthConfig{
				PasswordLength: 6,
			},
		},
	}
}

type ConfigMiddleware struct {
}

func (a ConfigMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		appName := model.GetAppName(r)
		config := configMap[appName]
		model.SetConfig(r, config)
		next.ServeHTTP(w, r)
	})
}
