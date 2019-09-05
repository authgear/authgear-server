package middleware

import (
	"net/http"

	"github.com/gomodule/redigo/redis"

	coreRedis "github.com/skygeario/skygear-server/pkg/core/redis"
)

// DBMiddleware setup Redis context in request
type RedisMiddleware struct {
	*redis.Pool
}

func (m RedisMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = r.WithContext(coreRedis.WithRedis(r.Context(), m.Pool))
		defer coreRedis.CloseConn(r.Context())
		next.ServeHTTP(w, r)
	})
}
