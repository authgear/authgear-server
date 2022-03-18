package web

import (
	"context"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/base32"
	"github.com/authgear/authgear-server/pkg/util/rand"
)

type dynamicCSPContextKeyType struct{}

var dynamicCSPContextKey = dynamicCSPContextKeyType{}

func WithCSPNonce(ctx context.Context, nonce string) context.Context {
	return context.WithValue(ctx, dynamicCSPContextKey, nonce)
}

func GetCSPNonce(ctx context.Context) string {
	nonce, _ := ctx.Value(dynamicCSPContextKey).(string)
	return nonce
}

type DynamicCSPMiddleware struct {
	HTTPConfig *config.HTTPConfig
}

func (m *DynamicCSPMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nonce := rand.StringWithAlphabet(32, base32.Alphabet, rand.SecureRand)
		r = r.WithContext(WithCSPNonce(r.Context(), nonce))

		cspDirectives, err := CSPDirectives(m.HTTPConfig.PublicOrigin, nonce)
		if err != nil {
			panic(err)
		}

		w.Header().Set("Content-Security-Policy", CSPJoin(cspDirectives))
		next.ServeHTTP(w, r)
	})
}
