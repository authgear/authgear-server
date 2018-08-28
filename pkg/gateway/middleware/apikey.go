package middleware

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/gateway/model"
)

var keyMap map[string]model.Key

func init() {
	keyMap = map[string]model.Key{
		"skygear": model.Key{
			APIKey:    "apikey",
			MasterKey: "masterkey",
		},
		"skygear-next": model.Key{
			APIKey:    "apikey-next",
			MasterKey: "masterkey-next",
		},
	}
}

type APIKeyMiddleware struct {
}

func (a APIKeyMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := model.GetAPIKey(r)
		keyType, appName := keyMapLookUp(apiKey)
		if keyType == model.NoAccessKey {
			http.Error(w, "API key not set", http.StatusBadRequest)
			return
		}

		model.SetAccessKeyType(r, keyType)
		model.SetAppName(r, appName)
		next.ServeHTTP(w, r)
	})
}

func keyMapLookUp(apiKey string) (model.KeyType, string) {
	for appName, key := range keyMap {
		if apiKey == key.APIKey {
			return model.APIAccessKey, appName
		}

		if apiKey == key.MasterKey {
			return model.MasterAccessKey, appName
		}
	}

	return model.NoAccessKey, ""
}
