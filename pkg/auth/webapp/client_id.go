package webapp

import (
	"context"
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/httputil"
)

// ClientIDCookieDef is a HTTP session cookie.
var ClientIDCookieDef = httputil.CookieDef{
	Name:     "client_id",
	Path:     "/",
	SameSite: http.SameSiteNoneMode,
}

type ClientIDMiddleware struct {
	CookieFactory CookieFactory
}

func (m *ClientIDMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		clientID := q.Get("client_id")

		// Persist client_id into cookie.
		// So that client_id no longer need to be present on the query.
		if clientID != "" {
			cookie := m.CookieFactory.ValueCookie(&ClientIDCookieDef, clientID)
			httputil.UpdateCookie(w, cookie)
		}

		// Restore client_id from cookie
		if clientID == "" {
			cookie, err := r.Cookie(ClientIDCookieDef.Name)
			if err == nil {
				clientID = cookie.Value
			}
		}

		// Restore client_id into the request context.
		if clientID != "" {
			ctx := WithClientID(r.Context(), clientID)
			r = r.WithContext(ctx)
		}

		next.ServeHTTP(w, r)
	})
}

type clientIDContextKeyType struct{}

var clientIDContextKey = clientIDContextKeyType{}

type clientIDContext struct {
	ClientID string
}

func WithClientID(ctx context.Context, clientID string) context.Context {
	v, ok := ctx.Value(clientIDContextKey).(*clientIDContext)
	if ok {
		v.ClientID = clientID
		return ctx
	}

	return context.WithValue(ctx, clientIDContextKey, &clientIDContext{
		ClientID: clientID,
	})
}

func GetClientID(ctx context.Context) string {
	v, ok := ctx.Value(clientIDContextKey).(*clientIDContext)
	if ok {
		return v.ClientID
	}
	return ""
}
