package transport

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/uiparam"
	"github.com/authgear/authgear-server/pkg/util/intl"
)

type UIParamMiddleware struct{}

func (m *UIParamMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var uiParam uiparam.T

		q := r.URL.Query()

		// client_id
		clientID := q.Get("client_id")
		uiParam.ClientID = clientID

		// ui_locales
		uiLocales := q.Get("ui_locales")
		uiParam.UILocales = uiLocales

		// state
		state := q.Get("state")
		uiParam.State = state

		// Put uiParam into context
		ctx := r.Context()
		ctx = uiparam.WithUIParam(ctx, &uiParam)
		if uiParam.UILocales != "" {
			tags := intl.ParseUILocales(uiParam.UILocales)
			ctx = intl.WithPreferredLanguageTags(ctx, tags)
		}
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
